package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Invite struct {
	Email       string     `json:"email"`
	InviteEmail string     `json:"inviteEmail"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func FindInvitationsByAnyEmail(email string) ([]Invite, error) {
	var res []Invite
	query := `SELECT from_email, to_email, confirmed_at, created_at 
              FROM invite 
              WHERE to_email=$1 OR from_email=$1`
	rows, err := DB.Query(query, email)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer CloseAndLog(rows)
		for rows.Next() {
			var inv Invite
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

func FindMyInvitations(email string) ([]Invite, error) {
	var res []Invite
	query := `SELECT from_email, to_email, confirmed_at, created_at 
              FROM invite 
              WHERE from_email=$1`
	rows, err := DB.Query(query, email)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer CloseAndLog(rows)
		for rows.Next() {
			var inv Invite
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

func DeleteInvite(fromEmail string, toEmail string) error {
	stmt, err := DB.Prepare("DELETE FROM invite WHERE from_email = $1 AND to_email = $2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM invite %v to %v, statement failed: %v", fromEmail, toEmail, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(fromEmail, toEmail)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func InsertInvite(id uuid.UUID, fromEmail string, toEmail string, now time.Time) error {
	stmt, err := DB.Prepare("INSERT INTO invite (id, from_email, to_email, created_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO invite for %v to %v, statement failed: %v", fromEmail, toEmail, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(id, fromEmail, toEmail, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateConfirmInviteAt(fromEmail string, toEmail string, now time.Time) error {
	stmt, err := DB.Prepare("UPDATE invite SET confirmed_at = $1 WHERE from_email = $2 and to_email=$3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE invite statement failed: %v", err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(now, fromEmail, toEmail)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}
