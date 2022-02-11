/**
 * We cannot use go-cron or other awesome cron implementations as we want to timewarp. So this is a quick
 * and dirty cron that checks every second if there needs to be done something. Its not efficient, but
 * a very simple solution that works with timewarping.
 */

package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	jobs []Job
	ctx  = context.Background()
)

type Job struct {
	job        func(now time.Time) error
	nextExecAt time.Time
	nextExec   func(now time.Time) time.Time
}

func init() {
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				again := true
				for again {
					again = false
					for k, job := range jobs {
						if job.nextExecAt.Before(timeNow()) {
							log.Printf("About to execute job [daily] at %v", job.nextExecAt)
							err := job.job(job.nextExecAt)
							if err != nil {
								log.Printf("Error in job [daily] run at %v: %v", job.nextExecAt, err)
							}
							nextExecAt := job.nextExec(job.nextExecAt)
							jobs[k].nextExecAt = nextExecAt
							again = true
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func cronStop() {
	ctx.Done()
}

func cronJobDay(job func(now time.Time) error, now time.Time) {
	jobs = append(jobs, Job{
		job:        job,
		nextExec:   timeDayPlusOne,
		nextExecAt: timeDayPlusOne(now)})
}

func timeDayPlusOne(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func cronJobHour(job func(now time.Time) error, now time.Time) {
	jobs = append(jobs, Job{
		job:        job,
		nextExec:   timeHourPlusOne,
		nextExecAt: timeHourPlusOne(now)})
}

func timeHourPlusOne(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
}

func hourlyRunner(now time.Time) error {
	//find repos that have an analysis older than 3 days
	a, err := findAllLatestAnalysisRequest(now.AddDate(0, 0, -3))
	if err != nil {
		return err
	}
	log.Printf("Daily Analysis Check found %v entries", len(a))

	/*for _, v := range a {
		if v.Url == nil {
			log.Printf("URL is nil of %v", v.Id)
			continue
		}
		if v.Branch == nil {
			log.Printf("Branch is nil of %v", v.Id)
			continue
		}
		err = analysisRequest(v.Id, *v.Url, *v.Branch)
		if err != nil {
			return err
		}
	}*/

	log.Printf("Hourly runner finished")
	return nil
}

func dailyRunner(now time.Time) error {
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	log.Printf("Start daily runner from %v to %v", yesterdayStart, yesterdayStop)

	nr, err := runDailyContribution(yesterdayStart, yesterdayStop)
	if err != nil {
		return err
	}
	log.Printf("Daily runner inserted %v entries", nr)

	return nil
}
