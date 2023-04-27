package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"time"
)

type UserBalance struct {
	UserId           uuid.UUID `json:"userId"`
	Balance          *big.Int  `json:"balance"`
	DailySpending    *big.Int  `json:"dailySpending"`
	PaymentCycleInId uuid.UUID `json:"paymentCycleInId"`
	BalanceType      string    `json:"balanceType"`
	Currency         string    `json:"currency"`
	CreatedAt        time.Time `json:"createdAt"`
}

func InsertUserBalance(ub UserBalance) error {
	stmt, err := db.Prepare(`INSERT INTO payment_event(
                                     payment_cycle_in_id, user_id, balance, daily_spending, 
                          			 balance_type, currency, created_at) 
                                   VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_balances for %v/%v statement event: %v", ub.UserId, ub.PaymentCycleInId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := ub.Balance.String()
	s := ub.DailySpending.String()
	res, err = stmt.Exec(ub.PaymentCycleInId, ub.UserId, b, s, ub.BalanceType, ub.Currency, ub.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindUserBalances(userId uuid.UUID) ([]UserBalance, error) {
	s := `SELECT payment_cycle_in_id, user_id, balance, currency, balance_type, created_at FROM payment_event WHERE user_id = $1`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userBalances []UserBalance
	for rows.Next() {
		var userBalance UserBalance
		b := ""
		err = rows.Scan(&userBalance.PaymentCycleInId, &userBalance.UserId, &b, &userBalance.Currency, &userBalance.BalanceType, &userBalance.CreatedAt)
		userBalance.Balance, _ = new(big.Int).SetString(b, 10)
		if err != nil {
			return nil, err
		}
		userBalances = append(userBalances, userBalance)
	}
	return userBalances, nil
}

func FindSumUserBalanceByCurrency(paymentCycleInId uuid.UUID) (map[string]*Balance, error) {
	rows, err := db.
		Query(`SELECT currency, daily_spending, COALESCE(sum(balance), 0)
                             FROM payment_event 
                             WHERE payment_cycle_in_id = $1
                             GROUP BY currency, daily_spending`, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*Balance)
	for rows.Next() {
		var c, b, s string
		err = rows.Scan(&c, &s, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		s1, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = &Balance{Balance: b1, DailySpending: s1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func InsertNewPaymentCycleIn(id uuid.UUID, userId uuid.UUID, seats int64, freq int64, createdAt time.Time) error {
	stmt, err := db.Prepare(`INSERT INTO payment_cycle_in(id, user_id, seats, freq, created_at) VALUES($1, $2, $3, $4, $5)`)
	if err != nil {
		return fmt.Errorf("prepareINSERT INTO payment_cycle_in statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, userId, seats, freq, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func InsertNewPaymentCycleOut(id uuid.UUID, userId uuid.UUID, createdAt time.Time) error {
	stmt, err := db.Prepare(`INSERT INTO payment_cycle_out(id, userId, created_at) VALUES($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("prepareINSERT INTO payment_cycle_in statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, userId, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateFreq(paymentCycleInId *uuid.UUID, freq int) error {
	stmt, err := db.Prepare(`UPDATE payment_cycle_in SET freq = $1 WHERE id=$2`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE payment_cycle_in for %v statement event: %v", freq, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(freq, paymentCycleInId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindPaymentCycle(paymentCycleInId uuid.UUID) (*PaymentCycle, error) {
	var pc PaymentCycle
	err := db.
		QueryRow(`SELECT id, seats, freq FROM payment_cycle_in WHERE id=$1`, paymentCycleInId).
		Scan(&pc.Id, &pc.Seats, &pc.Freq)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &pc, nil
	default:
		return nil, err
	}
}

func FindBalance(paymentCycleInId uuid.UUID, userId uuid.UUID, balanceType string, currency string) (*UserBalance, error) {
	var userBalance UserBalance
	userBalance.PaymentCycleInId = paymentCycleInId
	userBalance.UserId = userId
	userBalance.BalanceType = balanceType
	userBalance.Currency = currency
	s := `SELECT balance, created_at 
          FROM payment_event 
          WHERE payment_cycle_in_id = $1 
            AND balance_type = $2 
            AND currency = $3
            AND user_id = $4`
	var balanceString string
	err := db.QueryRow(s, paymentCycleInId, balanceType, currency, userId).
		Scan(&balanceString, &userBalance.CreatedAt)
	b, test := new(big.Int).SetString(balanceString, 10)
	if !test {
		return nil, fmt.Errorf("cannot convert bigint %v", balanceString)
	}
	userBalance.Balance = b
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &userBalance, nil
	default:
		return nil, err
	}
}

func FindPaymentCycleLast(uid uuid.UUID) (PaymentCycle, error) {
	pc := PaymentCycle{}
	err := db.
		QueryRow(`SELECT p.id, p.seats, p.freq 
                        FROM payment_cycle_in p JOIN users u on p.id = u.payment_cycle_in_id
                        WHERE u.id=$1`, uid).
		Scan(&pc.Id, &pc.Seats, &pc.Freq)
	switch err {
	case sql.ErrNoRows:
		return pc, nil
	case nil:
		return pc, nil
	default:
		return pc, err
	}
}
