package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"time"
)

func InsertFutureBalance(uid uuid.UUID, repoId uuid.UUID, paymentCycleInId *uuid.UUID, balance *big.Int,
	currency string, day time.Time, createdAt time.Time) error {

	stmt, err := db.Prepare(`
				INSERT INTO future_contribution(user_sponsor_id, repo_id, payment_cycle_in_id, balance, 
				                               currency, day, created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO future_contribution for %v statement event: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := balance.String()
	res, err = stmt.Exec(uid, repoId, paymentCycleInId, b, currency, day, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindSumDailyContributors(userContributorId uuid.UUID) (map[string]*Balance, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM daily_contribution 
                     WHERE user_contributor_id = $1
                     GROUP BY currency`, userContributorId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*Balance)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = &Balance{Balance: b1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func FindSumDailySponsors(userSponsorId uuid.UUID, paymentCycleInId uuid.UUID) (map[string]*Balance, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM daily_contribution 
                     WHERE user_sponsor_id = $1 AND payment_cycle_in_id = $2
                     GROUP BY currency`, userSponsorId, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m1 := make(map[string]*Balance)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m1[c] == nil {
			m1[c] = &Balance{Balance: b1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	rows, err = db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM future_contribution 
                     WHERE user_sponsor_id = $1 AND payment_cycle_in_id = $2
                     GROUP BY currency`, userSponsorId, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m2 := make(map[string]*Balance)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m2[c] == nil {
			m2[c] = &Balance{Balance: b1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	//TODO: integrate with loop above
	for k, _ := range m1 {
		if m2[k] != nil {
			m1[k].Balance = new(big.Int).Add(m2[k].Balance, m1[k].Balance)
		}
	}
	for k, _ := range m2 {
		if m1[k] == nil {
			m1[k] = m2[k]
		}
	}

	return m1, nil
}

func FindSumDailyBalanceCurrency(paymentCycleInId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.
		Query(`	SELECT currency, COALESCE(sum(balance), 0)
        				FROM daily_contribution 
                        WHERE payment_cycle_in_id = $1
                        GROUP BY currency`, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*big.Int)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = b1
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func FindSumFutureBalanceByRepoId(repoId *uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
                             FROM future_contribution 
                             WHERE repo_id = $1
                             GROUP BY currency`, repoId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*big.Int)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = b1
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func FindSumFutureBalanceByCurrency(paymentCycleInId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
                             FROM future_contribution 
                             WHERE payment_cycle_in_id = $1
                             GROUP BY currency`, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*big.Int)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = b1
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func InsertUnclaimed(email string, rid uuid.UUID, balance *big.Int, currency string, day time.Time, createdAt time.Time) error {
	stmt, err := db.Prepare(`
				INSERT INTO unclaimed(email, repo_id, balance, currency, day, created_at) 
				VALUES($1, $2, $3, $4, $5, $6)
				ON CONFLICT(email, repo_id, currency) DO UPDATE SET balance=$3, day=$5, created_at=$6`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO unclaimed for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := balance.String()
	res, err = stmt.Exec(email, rid, b, currency, day, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}
