package db

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	Email       string     `json:"email"`
	InviteEmail string     `json:"inviteEmail"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func (db *DB) FindInvitationsByAnyEmail(email string) ([]Invite, error) {
	rows, err := db.Query(`
		SELECT from_email, to_email, confirmed_at, created_at 
		FROM invite 
		WHERE to_email=$1 OR from_email=$1`, email)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var res []Invite
	for rows.Next() {
		var inv Invite
		err = rows.Scan(&inv.Email, &inv.InviteEmail, &inv.ConfirmedAt, &inv.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, inv)
	}
	return res, nil
}

func (db *DB) FindMyInvitations(email string) ([]Invite, error) {
	rows, err := db.Query(`
		SELECT from_email, to_email, confirmed_at, created_at 
		FROM invite 
		WHERE from_email=$1`, email)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var res []Invite
	for rows.Next() {
		var inv Invite
		err = rows.Scan(&inv.Email, &inv.InviteEmail, &inv.ConfirmedAt, &inv.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, inv)
	}
	return res, nil
}

func (db *DB) DeleteInvite(fromEmail string, toEmail string) error {
	_, err := db.Exec(
		`DELETE FROM invite WHERE from_email = $1 AND to_email = $2`,
		fromEmail, toEmail)
	return err
}

func (db *DB) InsertInvite(id uuid.UUID, fromEmail string, toEmail string, now time.Time) error {
	_, err := db.Exec(
		`INSERT INTO invite (id, from_email, to_email, created_at) VALUES ($1, $2, $3, $4)`,
		id, fromEmail, toEmail, now)
	return err
}

func (db *DB) UpdateConfirmInviteAt(fromEmail string, toEmail string, now time.Time) error {
	_, err := db.Exec(
		`UPDATE invite SET confirmed_at = $1 WHERE from_email = $2 AND to_email=$3`,
		now, fromEmail, toEmail)
	return err
}