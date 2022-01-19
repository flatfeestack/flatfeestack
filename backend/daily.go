package main

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"time"
)

//*********************************************************************************
//******************************* Daily calculations ******************************
//*********************************************************************************

func runDailyUserRepo(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`
			INSERT INTO daily_user_repo (user_id, repo_id, day, created_at)
			SELECT user_id, repo_id, $1::date, $3 
			FROM sponsor_event
			WHERE sponsor_at < $1 AND un_sponsor_at >= $2`)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return 0, err
	}
	return handleErr(res)

}

func runDailyBalances(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	rows, err := db.Query(`		
			SELECT user_id, ARRAY_AGG(repo_id)
    		FROM daily_user_repo d JOIN users u on u.id = d.user_id
			WHERE day = $1
            GROUP BY user_id`, yesterdayStart)
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
			log.Warningf("cannot scan %v", err)
			continue
		}

		u, err := findUserById(uid)
		if err != nil {
			log.Warningf("cannot find user %v", err)
			continue
		}

		mAdd, err := findSumUserBalanceByCurrency(u.PaymentCycleId)
		if err != nil {
			log.Warningf("cannot find sum user balance %v", err)
			continue
		}

		mSub, err := findSumDailyBalanceCurrency(u.PaymentCycleId)
		if err != nil {
			log.Warningf("cannot find sum daily balance %v", err)
			continue
		}

		currency, s, err := strategyDeductRandom(mAdd, mSub)
		if err != nil {
			log.Warningf("no funds, notify user %v", err)
			continue
		}

		for _, rid := range rids {
			err = insertDailyBalance(uid, rid, u.PaymentCycleId, s, currency, now, timeNow())
			if err != nil {
				log.Warningf("no funds, notify user %v", err)
				continue
			}
		}

		success++
	}
	return success, nil
}

func runDailyContribution(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	success := int64(0)

	return success, nil
}
