package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"math/big"
	"time"
)

func runDailyContribution(yesterdayStart time.Time, yesterdayStop time.Time) (int64, error) {
	rows, err := db.Query(`		
			SELECT user_id, ARRAY_AGG(repo_id)
			FROM sponsor_event
			WHERE sponsor_at < $1 AND (un_sponsor_at IS NULL OR un_sponsor_at >= $2)
			GROUP BY user_id`, yesterdayStart, yesterdayStop)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(rows)

	success := int64(0)
	for ; rows.Next(); success++ {
		uid := uuid.UUID{}
		rids := []uuid.UUID{}
		err = rows.Scan(&uid, pq.Array(&rids))
		if err != nil {
			return success, err
		}

		if len(rids) > 0 {
			err = calcContribution(uid, rids, yesterdayStart)
			if err != nil {
				return success, err
			}
		}
	}

	return success, nil
}

func calcContribution(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time) error {
	u, err := findUserById(uid)
	if err != nil {
		return fmt.Errorf("cannot find user %v", err)
	}
	//first check if the sponsor has enough funds
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
	}

	//if sponsor has no funds, use our funding
	//TODO
	//write email to sponsor, he is out of funds, please topup
	currency, freq, distributeDeduct, distributeAdd, deductFutureContribution, err := calcShare(u.PaymentCycleInId, int64(len(rids)))
	if err != nil {
		return fmt.Errorf("cannot calc share %v", err)
	}

	if freq > 0 {
		return doDeduct(uid, rids, yesterdayStart, u.PaymentCycleInId, currency, distributeDeduct, distributeAdd, deductFutureContribution)
	} else {
		//TODO
		//write email to user, out of funds, please topup
	}
	return nil
}

func doDeduct(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time, paymentCycleInId *uuid.UUID, currency string, distributeDeduct *big.Int, distributeAdd *big.Int, deductFutureContribution *big.Int) error {
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
				//TODO
				//not in map, send marketing email
			}
		}

		if len(uidMap) == 0 {
			//no contribution park the sponsoring separately
			err = insertFutureBalance(uid, rid, paymentCycleInId, distributeDeduct, currency, yesterdayStart, timeNow())
			if err != nil {
				return err
			}
		} else {
			for k, v := range uidMap {
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

	currency, freq, s, err := strategyDeductMax(mAdd, mSub)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("no funds, notify user %v", err)
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
