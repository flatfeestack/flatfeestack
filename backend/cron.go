/**
 * We cannot use go-cron or other awesome cron implementations as we want to timewarp. So this is a quick
 * and dirty cron that checks every second if there needs to be done something. Its not efficient, but
 * a very simple solution that works with timewarping.
 */

package main

import (
	"context"
	"fmt"
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

	nr, err := runDailyUserBalance(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily User Balance inserted %v entries", nr)

	nr, err = runDailyDaysLeftDailyPayment()
	if err != nil {
		return err
	}
	log.Printf("Daily Days Left Daily Payment updated %v entries", nr)

	nr, err = runDailyDaysLeftPaymentCycle()
	if err != nil {
		return err
	}
	log.Printf("Daily Days Left Payment Cycle updated %v entries", nr)

	nr, err = runDailyRepoBalance(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	log.Printf("Daily Repo Balance inserted %v entries", nr)

	//nr, err = runDailyEmailPayout(yesterdayStart, yesterdayStop, now)
	//if err != nil {
	//	return err
	//}
	//log.Printf("Daily Email Payout inserted %v entries", nr)

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

	/*	nr, err = runDailyUserContribution(yesterdayStart, yesterdayStop, now)
		if err != nil {
			return err
		}
		log.Printf("Daily User Contribution inserted %v entries", nr)*/

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
	}

	users, err := runDailyTopupReminderUser()
	if err != nil {
		return err
	}
	log.Printf("Daily Topup Reminder found %v entries", len(users))

	for _, u := range users {

		if u.SponsorId != nil {
			//topup user
			pc, err := findPaymentCycle(u.PaymentCycleId)
			if err != nil {
				log.Printf("error1 %v", err)
				continue
			}
			sponsor, err := findUserById(*u.SponsorId)
			if err != nil {
				log.Printf("error2 %v", err)
				continue
			}
			ok, _ := topupWithSponsor(&u, pc.Freq, sponsor.Email)
			if ok {
				continue
			}
		}

		err = reminderTopup(u)
		if err != nil {
			return err
		}
	}

	/*	userRepo, err := runDailyMarketing(yesterdayStart)
		if err != nil {
			return err
		}
		log.Printf("Daily Marketing candidates found %v entries", len(users))
		for _, u := range userRepo {
			//
			log.Printf("send mail to %v", u)
		}*/

	log.Printf("Daily runner finished")
	return nil
}

func reminderTopup(u User) error {
	emailCountId := "topup-" + u.PaymentCycleId.String()
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
		"template-subject-topup_", "We are about to top up your account",
		"template-plain-topup_", "Thanks for supporting with flatfeestack: "+other["url"],
		"template-html-topup_", other["lang"])

	go func(userId uuid.UUID, emailType string) {
		err := sendEmail(opts.EmailUrl, e)
		if err != nil {
			log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
		}
	}(u.Id, emailCountId)

	log.Printf("TOPUP, you are running out of credit %v", u)
	return nil
}

// ToDo: how to run monthly?
func monthlyRunner() error {
	chunkSize := 1000
	var container = make([][]PayoutCrypto, len(supportedCurrencies))
	supportedCurrenciesWithUSD := append(supportedCurrencies, CryptoCurrency{Name: "Dollar", ShortName: "USD"})

	payouts, err := monthlyBatchJobPayout()
	if err != nil {
		return err
	}

	// group container by currency [[eth], [neo], [tez], [usd]]
	for _, payout := range payouts {
		for i, currency := range supportedCurrenciesWithUSD {
			if payout.Currency == currency.ShortName {
				container[i] = append(container[i], payout)
			}
		}
	}

	for _, payouts := range container {
		currency := payouts[0].Currency

		for i := 0; i < len(payouts); i += chunkSize {
			end := i + chunkSize
			if end > len(payouts) {
				end = len(payouts)
			}
			var pts []PayoutToServiceCrypto
			batchId := uuid.New()
			for _, payout := range payouts[i:end] {
				request := PayoutRequestDB{
					UserId:    payout.UserId,
					BatchId:   batchId,
					Currency:  currency,
					Tea:       payout.Tea,
					Address:   payout.Address,
					CreatedAt: timeNow(),
				}
				err := insertPayoutRequest(&request)
				if err != nil {
					return err
				}

				pt := PayoutToServiceCrypto{
					Address: payout.Address,
					Tea:     payout.Tea,
				}
				pts = append(pts, pt)
			}
			err := cryptoPayout(pts, batchId, currency)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func cryptoPayout(pts []PayoutToServiceCrypto, batchId uuid.UUID, currency string) error {
	res, err := cryptoPayoutRequest(pts, currency)
	res.Currency = currency
	if err != nil {
		err1 := err.Error()
		err2 := insertPayoutsResponse(&PayoutsResponse{
			BatchId:   batchId,
			Error:     &err1,
			CreatedAt: timeNow(),
		})
		return fmt.Errorf("error %v/%v", err, err2)
	}
	return insertPayoutResponse(&PayoutResponseDB{
		BatchId:   batchId,
		Error:     nil,
		CreatedAt: timeNow(),
		Payouts:   *res,
	})
}
