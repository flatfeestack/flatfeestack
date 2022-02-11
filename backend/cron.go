/**
 * We cannot use go-cron or other awesome cron implementations as we want to timewarp. So this is a quick
 * and dirty cron that checks every second if there needs to be done something. Its not efficient, but
 * a very simple solution that works with timewarping.
 */

package main

import (
	"context"
	"github.com/google/uuid"
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

func reminderTopup(u User, sponsorEmailNotifed string) error {
	isSponsor := u.Email == sponsorEmailNotifed
	emailCountId := "topup-"
	if u.PaymentCycleInId != nil {
		emailCountId += u.PaymentCycleInId.String()
	}
	c, err := countEmailSent(u.Id, emailCountId)
	if err != nil {
		return err
	}

	if c > 0 {
		log.Printf("TOPUP, but we already sent a notification %v", u)
		return nil
	}

	err = insertEmailSent(u.Id, emailCountId, timeNow())
	if err != nil {
		return err
	}

	//check if user has stripe
	if u.PaymentCycleInId != nil && u.StripeId != nil && u.PaymentMethod != nil {
		_, err = stripePaymentRecurring(u)
		if err != nil {
			return err
		}

		email := u.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = opts.EmailLinkPrefix + "/user/payments"
		other["lang"] = "en"

		e := prepareEmail(email, other,
			"template-subject-topup-stripe_", "We are about to top up your account",
			"template-plain-topup-stripe_", "Thanks for supporting with flatfeestack: "+other["url"],
			"template-html-topup-stripe_", other["lang"])

		go func(userId uuid.UUID, emailType string) {
			err := sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}(u.Id, emailCountId)
	} else {
		//No stripe, just send email
		email := u.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = opts.EmailLinkPrefix + "/user/payments"
		other["lang"] = "en"

		if isSponsor {
			//we are sponser, and the user beneficiaryEmail could not donate
			e := prepareEmail(email, other,
				"template-subject-topup-other-sponsor_", "Your invited users could not sponsor anymore",
				"template-plain-topup-other-sponsor_", "Please add funds at: "+other["url"],
				"template-html-topup-other-sponsor_", other["lang"])

			go func(userId uuid.UUID, emailType string) {
				err := sendEmail(opts.EmailUrl, e)
				if err != nil {
					log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
				}
			}(u.Id, emailCountId)
		} else {
			//we are user, and the user beneficiaryEmail could not donate
			if u.InvitedId != nil {
				e := prepareEmail(email, other,
					"template-subject-topup-other-user1_", sponsorEmailNotifed+" (and you) are running low on funds",
					"template-plain-topup-other-user1_", "Please add funds at: "+other["url"],
					"template-html-topup-other-user1_", other["lang"])

				go func(userId uuid.UUID, emailType string) {
					err := sendEmail(opts.EmailUrl, e)
					if err != nil {
						log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
					}
				}(u.Id, emailCountId)
			} else {
				e := prepareEmail(email, other,
					"template-subject-topup-other-user2_", "You are running low on funding",
					"template-plain-topup-other-user2_", "Please add funds at: "+other["url"],
					"template-html-topup-other-user2_", other["lang"])

				go func(userId uuid.UUID, emailType string) {
					err := sendEmail(opts.EmailUrl, e)
					if err != nil {
						log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
					}
				}(u.Id, emailCountId)
			}
		}
	}

	log.Printf("TOPUP, you are running out of credit %v", u)
	return nil
}
