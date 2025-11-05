package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Marketing struct {
	Email    string
	Balances map[string]*big.Int
}

func (db *DB) FindGitEmailsByUserId(uid uuid.UUID) ([]GitEmail, error) {
	rows, err := db.Query(
		`SELECT email, confirmed_at, created_at FROM git_email WHERE user_id=$1`,
		uid)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var gitEmails []GitEmail
	for rows.Next() {
		var gitEmail GitEmail
		err = rows.Scan(&gitEmail.Email, &gitEmail.ConfirmedAt, &gitEmail.CreatedAt)
		if err != nil {
			return nil, err
		}
		gitEmails = append(gitEmails, gitEmail)
	}
	return gitEmails, nil
}

func (db *DB) CountExistingOrConfirmedGitEmail(uid uuid.UUID, email string) (int, error) {
	var c int
	err := db.QueryRow(
		`SELECT count(*) AS c FROM git_email WHERE (user_id=$1 OR confirmed_at IS NOT NULL) AND email=$2`,
		uid, email).Scan(&c)
	
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return c, nil
	default:
		return 0, err
	}
}

func (db *DB) InsertGitEmail(id uuid.UUID, uid uuid.UUID, email string, token *string, now time.Time) error {
	_, err := db.Exec(
		`INSERT INTO git_email(id, user_id, email, token, created_at) VALUES($1, $2, $3, $4, $5)`,
		id, uid, email, token, now)
	return err
}

func (db *DB) ConfirmGitEmail(email string, token string, now time.Time) error {
	_, err := db.Exec(
		`UPDATE git_email SET token=NULL, confirmed_at=$1 WHERE email=$2 AND token=$3`,
		now, email, token)
	return err
}

func (db *DB) DeleteGitEmail(uid uuid.UUID, email string) error {
	_, err := db.Exec(
		`DELETE FROM git_email WHERE email=$1 AND user_id=$2`,
		email, uid)
	return err
}

func (db *DB) DeleteGitEmailFromUserEmailsSent(uid uuid.UUID, email string) error {
	_, err := db.Exec(
		`DELETE FROM user_emails_sent WHERE email=$1 AND user_id=$2 AND email_type=$3`,
		email, uid, "add-git"+email)
	return err
}

func (db *DB) FindUserByGitEmail(gitEmail string) (*uuid.UUID, error) {
	var uid uuid.UUID
	err := db.QueryRow(
		`SELECT user_id FROM git_email WHERE email=$1`,
		gitEmail).Scan(&uid)
	
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &uid, nil
	default:
		return nil, err
	}
}

func (db *DB) FindUsersByGitEmails(emails []string) ([]uuid.UUID, error) {
	if len(emails) == 0 {
		return nil, nil
	}

	rows, err := db.Query(
		`SELECT DISTINCT user_id FROM git_email WHERE email = ANY($1)`,
		pq.Array(emails))
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var userIds []uuid.UUID
	for rows.Next() {
		var userId uuid.UUID
		if err := rows.Scan(&userId); err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}
	return userIds, nil
}

func (db *DB) GetRepoEmails(repoId uuid.UUID) ([]string, error) {
	rows, err := db.Query(
		`SELECT DISTINCT git_email FROM analysis_response WHERE repo_id = $1`,
		repoId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func (db *DB) InsertEmailSent(id uuid.UUID, userId *uuid.UUID, email string, emailType string, now time.Time) error {
	_, err := db.Exec(
		`INSERT INTO user_emails_sent(id, user_id, email, email_type, created_at) 
		 VALUES($1, $2, $3, $4, $5)`,
		id, userId, email, emailType, now)
	return err
}

func (db *DB) CountEmailSentById(userId uuid.UUID, emailType string) (int, error) {
	var c int
	err := db.QueryRow(
		`SELECT count(*) AS c FROM user_emails_sent WHERE user_id=$1 AND email_type=$2`,
		userId, emailType).Scan(&c)
	
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return c, nil
	default:
		return 0, err
	}
}

func (db *DB) CountEmailSentByEmail(email string, emailType string) (int, error) {
	var c int
	err := db.QueryRow(
		`SELECT count(*) AS c FROM user_emails_sent WHERE email=$1 AND email_type=$2`,
		email, emailType).Scan(&c)
	
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return c, nil
	default:
		return 0, err
	}
}

func (db *DB) InsertUnclaimed(id uuid.UUID, email string, repoId uuid.UUID, balance *big.Int, currency string,
	day time.Time, now time.Time) error {
	_, err := db.Exec(
		`INSERT INTO unclaimed(id, email, repo_id, balance, currency, day, created_at) 
		 VALUES($1, $2, $3, $4, $5, $6, $7)`,
		id, email, repoId, balance.String(), currency, day, now)
	return err
}

func (db *DB) FindMarketingEmails() ([]Marketing, error) {
	rows, err := db.Query(`
		SELECT u.email, u.currency, COALESCE(sum(u.balance), 0) as balances
		FROM unclaimed u
		LEFT JOIN git_email g ON u.email = g.email 
		WHERE g.email IS NULL 
		GROUP BY u.email, u.currency
		ORDER BY u.email, u.currency`)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var ms []Marketing
	mk := Marketing{Balances: make(map[string]*big.Int)}
	var email string
	var emailOld string

	for rows.Next() {
		var c, b string
		err = rows.Scan(&email, &c, &b)
		if err != nil {
			return nil, err
		}

		if emailOld != email && emailOld != "" {
			mk.Email = emailOld
			ms = append(ms, mk)
			mk = Marketing{Balances: make(map[string]*big.Int)}
			emailOld = email
		}

		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if mk.Balances[c] == nil {
			mk.Balances[c] = b1
		} else {
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}

		if emailOld == "" {
			emailOld = email
		}
	}
	
	if email != "" {
		mk.Email = email
		ms = append(ms, mk)
	}
	
	return ms, nil
}