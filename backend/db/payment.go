package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"time"
)

func InsertUserBalance(ub UserBalance) error {
	stmt, err := db.Prepare(`INSERT INTO user_balances(
                                            payment_cycle_in_id, 
                          	                user_id,
                                            from_user_id,
                                            balance,
                          					split,
                                            balance_type, 
                          					currency,
                                            created_at) 
                                    VALUES($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_balances for %v/%v statement event: %v", ub.UserId, ub.PaymentCycleInId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := ub.Balance.String()
	s := ub.Split.String()
	res, err = stmt.Exec(ub.PaymentCycleInId, ub.UserId, ub.FromUserId, b, s, ub.BalanceType, ub.Currency, ub.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func InsertNewPaymentCycleIn(seats int64, freq int64, createdAt time.Time) (*uuid.UUID, error) {
	stmt, err := db.Prepare(`INSERT INTO payment_cycle_in(seats, freq, created_at) 
                                    VALUES($1, $2, $3)  RETURNING id`)
	if err != nil {
		return nil, fmt.Errorf("prepareINSERT INTO payment_cycle_in statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(seats, freq, createdAt).Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}
	return &lastInsertId, nil
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

func FindPaymentCycle(paymentCycleInId *uuid.UUID) (*PaymentCycle, error) {
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
