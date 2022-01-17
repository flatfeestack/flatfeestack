package main

import (
	"database/sql"
	"github.com/google/uuid"
	"math/big"
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
			WHERE sponsor_at <= $1 AND unsponsor_at > $2`)
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
		var uid *uuid.UUID
		var rids []*uuid.UUID
		err = rows.Scan(uid, rids)
		if err == nil {
			success++
		}

		u, err := findUserById(*uid)
		if err == nil {
			success++
		}

		mAdd, err := findSumUserBalanceByCurrency(u.PaymentCycleId)
		if err == nil {
			success++
		}

		mSub, err := findSumDailyBalanceCurrency(u.PaymentCycleId)
		if err == nil {
			success++
		}

		c, b, err := strategyDeductRandom(mAdd, mSub)
		if err == nil {
			success++
		}

		for _, rid := range rids {
			instertDailyBalance(c, new(big.Int).Div(b, big.NewInt(int64(len(rids)))), uid, rid, u.PaymentCycleId)
		}

	}
	return success, nil
}
