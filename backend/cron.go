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

const (
	DAILY = iota + 1
	MONTHLY
)

var (
	jobs []Job
	ctx  = context.Background()
)

type Job struct {
	name     string
	myTime   int
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
							next := cronNext(job.myTime, *job.nextExec)
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

func cronJob(name string, myTime int, f func(now time.Time) error) {
	now := timeNow()
	next := cronNext(myTime, now)
	jobs = append(jobs, Job{
		name:     name,
		myTime:   myTime,
		f:        f,
		nextExec: &next,
	})
}

func cronNext(myTime int, now time.Time) time.Time {
	var next time.Time
	switch myTime {
	case MONTHLY:
		next = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	case DAILY:
		next = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	}
	return next
}
