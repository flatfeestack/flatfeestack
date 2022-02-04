package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"math/big"
	"time"
)

//*********************************************************************************
//******************************* Daily calculations ******************************
//*********************************************************************************

func runDailyContribution(yesterdayStart time.Time, yesterdayStop time.Time) (int64, error) {
	rows, err := db.Query(`		
			SELECT user_id, ARRAY_AGG(repo_id)
			FROM sponsor_event
			WHERE sponsor_at < $1 AND un_sponsor_at >= $2
			GROUP BY user_id`, yesterdayStart, yesterdayStop)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(rows)

	success := int64(0)
	for rows.Next() {
		uid := uuid.UUID{}
		rids := []uuid.UUID{}
		err = rows.Scan(&uid, pq.Array(&rids))
		if err != nil {
			return 0, err
		}

		err = calcContribution(uid, rids, yesterdayStart)
		if err != nil {
			return 0, err
		}
		success++
	}

	return success, nil
}

func calcContribution(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time) error {
	u, err := findUserById(uid)
	if err != nil {
		return fmt.Errorf("cannot find user %v", err)
	}

	mAdd, err := findSumUserBalanceByCurrency(u.PaymentCycleId)
	if err != nil {
		return fmt.Errorf("cannot find sum user balance %v", err)
	}

	mFut, err := findSumFutureBalanceByCurrency(u.PaymentCycleId)
	if err != nil {
		return fmt.Errorf("cannot find sum user balance %v", err)
	}

	mSub, err := findSumDailyBalanceCurrency(u.PaymentCycleId)
	if err != nil {
		return fmt.Errorf("cannot find sum daily balance %v", err)
	}

	currency, s, err := strategyDeductRandom(mAdd, mSub)
	if err != nil {
		return fmt.Errorf("no funds, notify user %v", err)
	}

	//split the contribution among the repos
	distribute := new(big.Int).Div(s, big.NewInt(int64(len(rids))))
	for _, rid := range rids {

		//get weights for the contributors
		ars, err := findAnalysisResponse(rid, yesterdayStart)
		if err != nil {
			return err
		}
		uidMap := map[uuid.UUID]float64{}
		total := 0.0
		for _, ar := range ars {
			uidGit, err := findUserByGitEmail(ar.GitEmail)
			if err != nil {
				return err
			}
			if uidGit != nil {
				uidMap[*uidGit] += ar.Weight
				total += ar.Weight
			} else {
				//not in map, send marketing email
			}
		}

		if len(uidMap) == 0 {
			//no contribution park the sponsoring separately
			err = insertFutureBalance(uid, rid, u.PaymentCycleId, distribute, currency, yesterdayStart, timeNow())
			if err != nil {
				return err
			}
		} else {
			for k, v := range uidMap {
				//we can distribute more, as we may have future balances
				if mFut[currency] != nil {
					distribute = new(big.Int).Add(distribute, mFut[currency])
					//if we distribute more, we need to deduct this from the future balances
					deduct := new(big.Int).Neg(mFut[currency])
					err = insertFutureBalance(uid, rid, u.PaymentCycleId, deduct, currency, yesterdayStart, timeNow())
					if err != nil {
						return err
					}
				}
				f := new(big.Float).SetInt(distribute)
				amountF := new(big.Float).Mul(big.NewFloat(v/total), f)
				amount := new(big.Int)
				amountF.Int(amount)
				uidGit, err := findUserById(k)
				if err != nil {
					return err
				}

				insertContribution(uid, k, rid, u.PaymentCycleId, uidGit.PaymentCycleId, amount, currency, yesterdayStart, timeNow())
			}
		}
	}
	return nil
}
