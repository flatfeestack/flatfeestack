package main

import (
	"backend/api"
	"backend/clients"
	"backend/db"
	"backend/utils"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"math"
	"math/big"
	"time"
)

func hourlyRunner(now time.Time) error {
	//find repos that have an analysis older than 2 days
	a, err := db.FindAllLatestAnalysisRequest(now.AddDate(0, 0, -1))
	if err != nil {
		return err
	}
	log.Printf("Start hourly analysis check with %v entries", len(a))

	nr := 0
	for _, v := range a {
		//check if we need analysis request
		err := clients.AnalysisReq(v.RepoId, v.GitUrl)
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

	sponsorResults, err := db.FindSponsorsBetween(yesterdayStart, yesterdayStop)
	if err != nil {
		return err
	}

	nr := 0
	for _, sponsorResult := range sponsorResults {
		if len(sponsorResult.RepoIds) > 0 {
			err = calcContribution(sponsorResult.UserId, sponsorResult.RepoIds, yesterdayStart)
			nr++
			if err != nil {
				return err
			}
		}
	}

	log.Printf("Daily runner inserted %v entries", nr)

	//aggregate marketing emails
	ms, err := db.FindMarketingEmails()
	for _, v := range ms {
		if err != nil {
			return err
		}
		repoNames := []string{}
		//TODO: fetch repo names
		clients.SendMarketingEmail(v.Email, v.Balances, repoNames)
	}

	return nil
}

func calcContribution(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time) error {
	u, err := db.FindUserById(uid)
	if err != nil {
		return fmt.Errorf("cannot find user %v", err)
	}
	//first check if the sponsor has enough funds
	if u.InvitedId != nil {
		u1, err := db.FindUserById(*u.InvitedId)
		if err != nil {
			return fmt.Errorf("cannot find invited user %v", err)
		}

		return calcAndDeduct(u1, rids, yesterdayStart, u)
	}
	//TODO: also notify the not only the parent of insufficient funds
	return calcAndDeduct(u, rids, yesterdayStart, nil)
}

func calcAndDeduct(u *db.UserDetail, rids []uuid.UUID, yesterdayStart time.Time, uOrig *db.UserDetail) error {
	currency, freq, distributeDeduct, distributeAdd, deductFutureContribution, err := calcShare(u.Id, int64(len(rids)))
	if err != nil {
		return fmt.Errorf("cannot calc share %v", err)
	}

	if freq <= 1 {
		reminderTopUp(*u, uOrig)
	}

	if freq > 0 {
		err = doDeduct(u.Id, rids, yesterdayStart, currency, distributeDeduct, distributeAdd, deductFutureContribution)
		return err
	} else {
		log.Debugf("the user %v is out of funds", u.Id)
	}
	return nil
}

func doDeduct(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time, currency string, distributeDeduct *big.Int, distributeAdd *big.Int, deductFutureContribution *big.Int) error {
	for _, rid := range rids {
		//get weights for the contributors
		a, err := db.FindLatestAnalysisRequest(rid)
		if err != nil {
			return err
		}
		if a == nil {
			continue
		}
		ars, err := db.FindAnalysisResults(a.Id)
		if err != nil {
			return err
		}
		uidInMap := map[uuid.UUID]float64{}
		uidNotInMap := map[string]float64{}
		total := 0.0
		for _, ar := range ars {
			uidGit, err := db.FindUserByGitEmail(ar.GitEmail)
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

		for email, v := range uidNotInMap {
			amount := calcSharePerUser(distributeAdd, v, math.Max(total, v))
			id := uuid.New()
			err = db.InsertUnclaimed(id, email, rid, amount, currency, yesterdayStart, utils.TimeNow())
			if err != nil {
				log.Errorf("insertUnclaimed failed: %v, %v\n", opts.EmailUrl, err)
			}
		}

		if len(uidInMap) == 0 {
			//no contribution park the sponsoring separately
			err = db.InsertFutureContribution(uid, rid, distributeDeduct, currency, yesterdayStart, utils.TimeNow())
			if err != nil {
				return err
			}
		} else {
			for k, v := range uidInMap {
				//we can distribute more, as we may have future balances
				if deductFutureContribution != nil {
					err = db.InsertFutureContribution(uid, rid, deductFutureContribution, currency, yesterdayStart, utils.TimeNow())
					if err != nil {
						return err
					}
				}

				amount := calcSharePerUser(distributeAdd, v, total)
				err = db.InsertContribution(uid, k, rid, amount, currency, yesterdayStart, utils.TimeNow())
				if err != nil {
					return err
				}
				//TODO: notify users that they have funds, use the following steps
				//use KW to send every week a mail, don't spam only if amount has
				//increased by 20% in the meantime
			}
		}
	}
	return nil
}

func calcSharePerUser(distributeAdd *big.Int, v float64, total float64) *big.Int {
	distributeAddF := new(big.Float).SetInt(distributeAdd)
	amountF := new(big.Float).Mul(big.NewFloat(v/total), distributeAddF)
	amount := new(big.Int)
	amountF.Int(amount)
	return amount
}

func calcShare(userId uuid.UUID, repoLen int64) (string, int64, *big.Int, *big.Int, *big.Int, error) {
	//mAdd is what the user paid in the current cycle
	mAdd, err := db.FindSumPaymentByCurrency(userId, db.PayInSuccess)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum user balance %v", err)
	}

	//either the user spent it on a repo that does not have any devs who can claim
	mFut, err := db.FindSumFutureSponsors(userId)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum user balance %v", err)
	}

	//or the user spent it on for a repo with a dev who can claim
	mSub, err := db.FindSumDailySponsors(userId)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum daily balance %v", err)
	}

	currency, freq, s, err := api.StrategyDeductMax(userId, mAdd, mSub, mFut)

	if s == nil {
		return currency, freq, nil, nil, nil, nil
	}
	//split the contribution among the repos
	distributeDeduct := new(big.Int).Div(s, big.NewInt(repoLen))
	distributeAdd := distributeDeduct
	var deductFutureContribution *big.Int
	if mFut[currency] != nil {
		distributeAdd = new(big.Int).Add(distributeDeduct, mFut[currency])
		//if we distribute more, we need to deduct this from the future balances
		deductFutureContribution = new(big.Int).Neg(mFut[currency])
	}
	return currency, freq, distributeDeduct, distributeAdd, deductFutureContribution, nil
}

func reminderTopUp(u db.UserDetail, uOrig *db.UserDetail) error {

	//check if user has stripe
	if u.StripeId != nil && u.PaymentMethod != nil {
		err := api.StripePaymentRecurring(u)
		if err != nil {
			return err
		}

		err = clients.SendStripeTopUp(u)
		if err != nil {
			return err
		}
	} else {
		//No stripe, just send email
		isSponsor := uOrig != nil
		if isSponsor {
			err := clients.SendTopUpSponsor(u)
			if err != nil {
				return err
			}
		} else {
			if u.InvitedId != nil {
				err := clients.SendTopUpInvited(u)
				if err != nil {
					return err
				}
			} else {
				err := clients.SendTopUpOther(u)
				if err != nil {
					return err
				}
			}
		}
	}

	log.Printf("TOPUP, you are running out of credit %v", u)
	return nil
}
