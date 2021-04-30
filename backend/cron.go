/**
 * We cannot use go-cron or other awesome cron implementations as we want to timewarp. So this is a quick
 * and dirty cron that checks every second if there needs to be done something. Its not efficient, but
 * a very simple solution that works with timewarping.
 */

package main

import (
	"context"
	"github.com/google/uuid"
	"log"
	"time"
)

var (
	jobs []Job
	ctx  = context.Background()
)

type Job struct {
	job        func(now time.Time) error
	addTime    int
	nextExecAt time.Time
	nextExec   func(addTime int, now time.Time) time.Time
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
							nextExecAt := job.nextExec(job.addTime, job.nextExecAt)
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
		addTime:    1,
		nextExec:   timeDay,
		nextExecAt: timeDay(1, now)})
}

func timeDay(days int, now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+days, 0, 0, 0, 0, now.Location())
}

//func cronJobHour(job func(now time.Time) error, now time.Time) {
//	jobs = append(jobs, Job{
//		job:        job,
//		nextExec:   timeHour,
//		nextExecAt: timeHour(1, now)})
//}

//func timeHour(hours int, now time.Time) time.Time {
//	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+hours, 0, 0, 0, now.Location())
//}

//func hourlyRunner(now time.Time) error {
//	log.Printf("Hourly runner finished")
//	return nil
//}

func dailyRunner(now time.Time) error {
	yesterdayStop := timeDay(0, now)
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	log.Printf("Start daily runner from %v to %v", yesterdayStart, yesterdayStop)
	nr, err := runDailyRepoHours(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily Repo Hours inserted %v entries", nr)

	nr, err = runDailyUserBalance(yesterdayStart, now)
	if err != nil {
		return err
	}
	log.Printf("Daily User Balance inserted %v entries", nr)

	nr, err = runDailyDaysLeft(yesterdayStart)
	if err != nil {
		return err
	}
	log.Printf("Daily Days Left inserted %v entries", nr)

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

	users, err := runDailyTopupReminderUser()
	if err != nil {
		return err
	}
	log.Printf("Daily Topup Reminder found %v entries", len(users))

	for _, u := range users {
		err = reminderTopup(u)
		if err != nil {
			return err
		}
	}

	log.Printf("Daily runner finished")
	return nil
}

func reminderTopup(u User) error {
	emailCountId := "topup-" + u.PaymentCycleId.String()
	c, err := countEmailSent(u.Id, emailCountId)
	if err != nil {
		return err
	}

	if c < 2 {
		err = insertEmailSent(u.Id, "topup-"+u.PaymentCycleId.String(), timeNow())
		if err != nil {
			return err
		}

		_, err = stripePaymentRecurring(u)
		if err != nil {
			return err
		}

		email := *u.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = opts.EmailLinkPrefix + "/dashboard/profile"
		other["lang"] = "en"

		e := prepareEmail(email, other,
			"template-subject-topup_", "We are about to top up your account",
			"template-plain-topup_", "Thanks for supporting with flatfeestack: "+other["url"],
			"template-html-topup_", other["lang"])

		go func(userId uuid.UUID, emailType string) {
			err := sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}(u.Id, emailCountId)
	}

	log.Printf("TOPUP, you are running out of credit %v", u)
	return nil
}
