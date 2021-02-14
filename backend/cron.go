/**
 * We cannot use go-cron or other awesome cron implementations as we want to timewarp. So this is a quick
 * and dirty cron that checks every second if there needs to be done something. Its not efficient, but
 * a very simple solution that works with timewarping.
 */

package main

import (
	"context"
	"log"
	"time"
)

var (
	jobs []Job
	ctx  = context.Background()
)

type Job struct {
	job      func(now time.Time) error
	nextExec *time.Time
}

func init() {
	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				again := true
				for again {
					again = false
					for k, job := range jobs {
						if job.nextExec.Before(timeNow()) {
							log.Printf("About to execute job [daily] at %v", job.nextExec)
							err := job.job(*job.nextExec)
							if err != nil {
								log.Printf("Error in job [daily] run at %v: %v", job.nextExec, err)
							}
							nextExec := timeDay(1, *job.nextExec)
							jobs[k].nextExec = &nextExec
							again = true
						}
					}
				}
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}

func cronStop() {
	ctx.Done()
}

func cronJob(job func(now time.Time) error, now time.Time) {
	nextExec := timeDay(1, now)
	jobs = append(jobs, Job{job, &nextExec})
}

func timeDay(days int, now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+days, 0, 0, 0, 0, now.Location())
}

func dailyRunner(now time.Time) error {
	yesterdayStop := timeDay(0, now)
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	log.Printf("Start daily runner from %v to %v", yesterdayStart, yesterdayStop)
	nr, err := runDailyRepoHours(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily Repo Hours inserted %v entries", nr)

	nr, err = runDailyRepoBalance(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily Repo Balance inserted %v entries", nr)

	nr, err = runDailyEmailPayout(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily Email Payout inserted %v entries", nr)

	nr, err = runDailyRepoWeight(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily Repo Weight inserted %v entries", nr)

	nr, err = runDailyUserPayout(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily User Payout inserted %v entries", nr)

	nr, err = runDailyFutureLeftover(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily Leftover inserted %v entries", nr)

	repos, err := runDailyAnalysisCheck(now, 5)
	if err != nil {
		return err
	}
	log.Printf("Daily Analysis Check found %v entries", len(repos))

	for _, v := range repos {
		err = analysisRequest(v.Id, *v.Url, *v.Branch)
		if err != nil {
			return err
		}
	}

	log.Printf("Daily runner finished")
	return nil
}
