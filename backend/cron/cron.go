/**
 * We cannot use go-cron or other awesome cron implementations as we want to timewarp. So this is a quick
 * and dirty cron that checks every second if there needs to be done something. Its not efficient, but
 * a very simple solution that works with timewarping.
 */

package cron

import (
	"backend/util"
	"log/slog"
	"time"
)

var (
	jobs []Job
	done = make(chan bool)
	t    *time.Ticker
)

type Job struct {
	job          func(now time.Time) error
	nextExecAt   time.Time
	nextExecFunc func(now time.Time) time.Time
}

func init() {
	go func() {
		t = time.NewTicker(5 * time.Second)
		for {
			select {
			case <-t.C:
				check()
			case <-done:
				return
			}
		}
	}()
}

func check() {
	for k, job := range jobs {
		for job.nextExecAt.Before(util.TimeNow()) {
			slog.Info("About to execute job",
				slog.Any("time", job.nextExecAt))
			err := job.job(job.nextExecAt)
			if err != nil {
				slog.Error("Error in job run",
					slog.Any("time", job.nextExecAt),
					slog.Any("error", err))
			}
			job.nextExecAt = job.nextExecFunc(job.nextExecAt)
			jobs[k].nextExecAt = job.nextExecAt
		}
	}
}

func CheckNow() {
	check()
}

func CronStop() {
	done <- true
	t.Stop()
}

func CronJobDay(job func(now time.Time) error, now time.Time) {
	jobs = append(jobs, Job{
		job:          job,
		nextExecFunc: timeDayPlusOne,
		nextExecAt:   now}) //run this job at startup
}

func timeDayPlusOne(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
}

func CronJobHour(job func(now time.Time) error, now time.Time) {
	jobs = append(jobs, Job{
		job:          job,
		nextExecFunc: timeHourPlusOne,
		nextExecAt:   now}) //run this job at startup
}

func timeHourPlusOne(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, now.Minute(), now.Second(), now.Nanosecond(), now.Location())
}
