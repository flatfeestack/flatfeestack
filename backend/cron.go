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
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"math/big"
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
	//find repos that have an analysis older than 2 days
	a, err := findAllLatestAnalysisRequest(now.AddDate(0, 0, -2))
	if err != nil {
		return err
	}
	log.Printf("Start hourly analysis check with %v entries", len(a))

	nr := 0
	for _, v := range a {
		//check if we need analysis request
		err := analysisRequest(v.RepoId, v.GitUrls)
		if err != nil {
			log.Warnf("analysis request failed %v", err)
		} else {
			nr++
		}
	}

	log.Printf("Hourly runner processed %v entries", nr)
	return nil
}

func dailyRunner(now time.Time) error {
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	log.Printf("Start daily runner from %v to %v", yesterdayStart, yesterdayStop)

	rows, err := db.Query(`		
			SELECT user_id, ARRAY_AGG(repo_id)
			FROM sponsor_event
			WHERE sponsor_at < $1 AND (un_sponsor_at IS NULL OR un_sponsor_at >= $2)
			GROUP BY user_id`, yesterdayStart, yesterdayStop)
	if err != nil {
		return err
	}
	defer closeAndLog(rows)

	nr := 0
	for ; rows.Next(); nr++ {
		uid := uuid.UUID{}
		rids := []uuid.UUID{}
		err = rows.Scan(&uid, pq.Array(&rids))
		if err != nil {
			return err
		}

		if len(rids) > 0 {
			err = calcContribution(uid, rids, yesterdayStart)
			if err != nil {
				return err
			}
		}
	}

	log.Printf("Daily runner inserted %v entries", nr)
	return nil
}

func calcContribution(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time) error {
	u, err := findUserById(uid)
	if err != nil {
		return fmt.Errorf("cannot find user %v", err)
	}
	//first check if the sponsor has enough funds
	var sponsorEmailNotifed string
	if u.InvitedId != nil {
		u1, err := findUserById(*u.InvitedId)
		if err != nil {
			return fmt.Errorf("cannot find invited user %v", err)
		}

		currency, freq, distributeDeduct, distributeAdd, deductFutureContribution, err := calcShare(u1.PaymentCycleInId, int64(len(rids)))
		if err != nil {
			return fmt.Errorf("cannot calc invited share %v", err)
		}

		if freq > 0 {
			return doDeduct(uid, rids, yesterdayStart, u1.PaymentCycleInId, currency, distributeDeduct, distributeAdd, deductFutureContribution)
		}

		sponsorEmailNotifed = u1.Email
		reminderTopUp(*u1, sponsorEmailNotifed)
	}

	currency, freq, distributeDeduct, distributeAdd, deductFutureContribution, err := calcShare(u.PaymentCycleInId, int64(len(rids)))
	if err != nil {
		return fmt.Errorf("cannot calc share %v", err)
	}

	if freq > 0 {
		return doDeduct(uid, rids, yesterdayStart, u.PaymentCycleInId, currency, distributeDeduct, distributeAdd, deductFutureContribution)
	}
	reminderTopUp(*u, sponsorEmailNotifed)
	return nil
}

func doDeduct(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time, paymentCycleInId *uuid.UUID, currency string, distributeDeduct *big.Int, distributeAdd *big.Int, deductFutureContribution *big.Int) error {
	for _, rid := range rids {
		//get weights for the contributors
		a, err := findLatestAnalysisRequest(rid)
		if err != nil {
			return err
		}
		if a == nil {
			continue
		}
		ars, err := findAnalysisResults(a.Id)
		if err != nil {
			return err
		}
		uidInMap := map[uuid.UUID]float64{}
		uidNotInMap := map[string]float64{}
		total := 0.0
		for _, ar := range ars {
			uidGit, err := findUserByGitEmail(ar.GitEmail)
			if err != nil {
				return err
			}
			if uidGit != nil {
				uidInMap[*uidGit] += ar.Weight
				total += ar.Weight
			} else {
				uidNotInMap[ar.GitEmail] += ar.Weight
			}
		}

		//TODO: write marketing email to uidNotInMap with distributeAdd*w/total

		if len(uidInMap) == 0 {
			//no contribution park the sponsoring separately
			err = insertFutureBalance(uid, rid, paymentCycleInId, distributeDeduct, currency, yesterdayStart, timeNow())
			if err != nil {
				return err
			}
		} else {
			for k, v := range uidInMap {
				//we can distribute more, as we may have future balances
				if deductFutureContribution != nil {
					err = insertFutureBalance(uid, rid, paymentCycleInId, deductFutureContribution, currency, yesterdayStart, timeNow())
					if err != nil {
						return err
					}
				}
				distributeAddF := new(big.Float).SetInt(distributeAdd)
				amountF := new(big.Float).Mul(big.NewFloat(v/total), distributeAddF)
				amount := new(big.Int)
				amountF.Int(amount)
				uidGit, err := findUserById(k)
				if err != nil {
					return err
				}

				insertContribution(uid, k, rid, paymentCycleInId, uidGit.PaymentCycleOutId, amount, currency, yesterdayStart, timeNow())

				//TODO: notify users that they have funds, use the following steps
				//use KW to send every week a mail, don't spam only if amount has
				//increased by 20% in the meantime
			}
		}
	}
	return nil
}

func calcShare(paymentCycleInId *uuid.UUID, rLen int64) (string, int64, *big.Int, *big.Int, *big.Int, error) {
	mAdd, err := findSumUserBalanceByCurrency(paymentCycleInId)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum user balance %v", err)
	}

	mFut, err := findSumFutureBalanceByCurrency(paymentCycleInId)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum user balance %v", err)
	}

	mSub, err := findSumDailyBalanceCurrency(paymentCycleInId)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum daily balance %v", err)
	}

	currency, freq, s := strategyDeductMax(mAdd, mSub)

	if s == nil {
		return currency, freq, nil, nil, nil, nil
	}
	//split the contribution among the repos
	distributeDeduct := new(big.Int).Div(s, big.NewInt(rLen))
	distributeAdd := distributeDeduct
	var deductFutureContribution *big.Int
	if mFut[currency] != nil {
		distributeAdd = new(big.Int).Add(distributeDeduct, mFut[currency])
		//if we distribute more, we need to deduct this from the future balances
		deductFutureContribution = new(big.Int).Neg(mFut[currency])
	}
	return currency, freq, distributeDeduct, distributeAdd, deductFutureContribution, nil
}

func reminderTopUp(u User, sponsorEmailNotifed string) error {
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
