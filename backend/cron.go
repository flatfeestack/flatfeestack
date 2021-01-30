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
	name     string
	days     int
	f        func(now time.Time) error
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
							log.Printf("About to execute job [%v] at %v", job.name, job.nextExec)
							err := job.f(*job.nextExec)
							if err != nil {
								log.Printf("Error in job %v run at %v: %v", job.name, job.nextExec, err)
							}
							next := cronNext(job.days, *job.nextExec)
							jobs[k].nextExec = &next
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

func cronJob(name string, days int, f func(now time.Time) error) {
	now := timeNow()
	next := cronNext(days, now)
	jobs = append(jobs, Job{
		name:     name,
		days:     days,
		f:        f,
		nextExec: &next,
	})
}

func cronNext(days int, now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+days, 0, 0, 0, 0, now.Location())
}
