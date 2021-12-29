package main

import (
	"database/sql"
	"fmt"
	"time"
)

type dbInvite struct {
	Email       string     `json:"email"`
	Freq        int64      `json:"freq"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func findInvitationsByEmail(email string) ([]dbInvite, error) {
	var res []dbInvite
	query := `SELECT email, confirmed_at, freq, created_at 
              FROM invite 
              WHERE invite_email=$1`
	rows, err := db.Query(query, email)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer closeAndLog(rows)
		for rows.Next() {
			var inv dbInvite
			err = rows.Scan(&inv.Email, &inv.ConfirmedAt, &inv.Freq, &inv.CreatedAt)
			if err != nil {
				return nil, err
			}
			res = append(res, inv)
		}
		return res, nil
	default:
		return nil, err
	}
}

func deleteInvite(myEmail string, inviteEmail string) error {
	stmt, err := db.Prepare("DELETE FROM invite WHERE email = $1 AND invite_email = $2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM invite %v statement failed: %v", myEmail, err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(myEmail, inviteEmail)
	return handleErrMustInsertOne(res)
}

func insertInvite(email string, inviteEmail string, freq int64, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO invite (email, invite_email, freq, created_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO invite for %v statement failed: %v", email, err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(email, inviteEmail, freq, now)
	return handleErrMustInsertOne(res)
}

func updateConfirmInviteAt(email string, inviteEmail string, now time.Time) error {
	stmt, err := db.Prepare("UPDATE invite SET confirmed_at = $1 WHERE email = $2 and invite_email=$3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE invite statement failed: %v", err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(now, email, inviteEmail)
	return handleErrMustInsertOne(res)
}
