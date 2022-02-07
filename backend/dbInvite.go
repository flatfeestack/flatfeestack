package main

import (
	"database/sql"
	"fmt"
	"time"
)

type dbInvite struct {
	Email       string     `json:"email"`
	InviteEmail string     `json:"inviteEmail"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func findInvitationsByAnyEmail(email string) ([]dbInvite, error) {
	var res []dbInvite
	query := `SELECT email, invite_email, confirmed_at, freq, created_at 
              FROM invite 
              WHERE invite_email=$1 OR email=$1`
	rows, err := db.Query(query, email)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer closeAndLog(rows)
		for rows.Next() {
			var inv dbInvite
			err = rows.Scan(&inv.Email, &inv.InviteEmail, &inv.ConfirmedAt, &inv.CreatedAt)
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

func findMyInvitations(email string) ([]dbInvite, error) {
	var res []dbInvite
	query := `SELECT email, invite_email, confirmed_at, created_at 
              FROM invite 
              WHERE email=$1`
	rows, err := db.Query(query, email)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer closeAndLog(rows)
		for rows.Next() {
			var inv dbInvite
			err = rows.Scan(&inv.Email, &inv.InviteEmail, &inv.ConfirmedAt, &inv.CreatedAt)
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
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertInvite(email string, inviteEmail string, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO invite (email, invite_email, created_at) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO invite for %v statement failed: %v", email, err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(email, inviteEmail, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

/*func findConfirmInviteAt(email string, inviteEmail string) (*time.Time, error) {
	var confirmedAt time.Time
	err := db.
		QueryRow(`SELECT confirmedAt
                        FROM invite
                        WHERE email = $1 and invite_email=$2`, email, inviteEmail).
		Scan(&confirmedAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &confirmedAt, nil
	default:
		return nil, err
	}
}*/

func updateConfirmInviteAt(email string, inviteEmail string, now time.Time) error {
	stmt, err := db.Prepare("UPDATE invite SET confirmed_at = $1 WHERE email = $2 and invite_email=$3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE invite statement failed: %v", err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(now, email, inviteEmail)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}
