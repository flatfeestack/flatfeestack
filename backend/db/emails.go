package db

import (
	"database/sql"
	"fmt"
	dbLib "github.com/flatfeestack/go-lib/database"
	"github.com/google/uuid"
	"math/big"
	"time"
)

type Marketing struct {
	Email    string
	Balances map[string]*big.Int
}

func FindGitEmailsByUserId(uid uuid.UUID) ([]GitEmail, error) {
	var gitEmails []GitEmail
	s := "SELECT email, confirmed_at, created_at FROM git_email WHERE user_id=$1"
	rows, err := dbLib.DB.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(rows)

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

func CountExistingOrConfirmedGitEmail(uid uuid.UUID, email string) (int, error) {
	var c int
	err := dbLib.DB.
		QueryRow(`SELECT count(*) AS c FROM git_email WHERE (user_id=$1 or confirmed_at is not null) and email=$2`, uid, email).
		Scan(&c)
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return c, nil
	default:
		return c, err
	}
}

func InsertGitEmail(id uuid.UUID, uid uuid.UUID, email string, token *string, now time.Time) error {
	stmt, err := dbLib.DB.Prepare("INSERT INTO git_email(id, user_id, email, token, created_at) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO git_email for %v statement event: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, uid, email, token, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func ConfirmGitEmail(email string, token string, now time.Time) error {
	stmt, err := dbLib.DB.Prepare("UPDATE git_email SET token=NULL, confirmed_at=$1 WHERE email=$2 AND token=$3")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(now, email, token)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func DeleteGitEmail(uid uuid.UUID, email string) error {
	//TODO: don't delete, just mark as deleted
	stmt, err := dbLib.DB.Prepare("DELETE FROM git_email WHERE email=$1 AND user_id=$2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(email, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindUserByGitEmail(gitEmail string) (*uuid.UUID, error) {
	var uid uuid.UUID
	err := dbLib.DB.
		QueryRow(`SELECT user_id FROM git_email WHERE email=$1`, gitEmail).
		Scan(&uid)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &uid, nil
	default:
		return nil, err
	}
}

//******* Emails Sent

func InsertEmailSent(id uuid.UUID, userId *uuid.UUID, email string, emailType string, now time.Time) error {
	stmt, err := dbLib.DB.Prepare(`
			INSERT INTO user_emails_sent(id, user_id, email, email_type, created_at) 
			VALUES($1, $2, $3, $4, $5)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_emails_sent for %v statement event: %v", userId, err)
	}
	defer dbLib.CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, userId, email, emailType, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func CountEmailSentById(userId uuid.UUID, emailType string) (int, error) {
	var c int
	err := dbLib.DB.
		QueryRow(`SELECT count(*) AS c FROM user_emails_sent WHERE user_id=$1 and email_type=$2`, userId, emailType).
		Scan(&c)
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return c, nil
	default:
		return 0, err
	}
}

func CountEmailSentByEmail(email string, emailType string) (int, error) {
	var c int
	err := dbLib.DB.
		QueryRow(`SELECT count(*) AS c FROM user_emails_sent WHERE email=$1 and email_type=$2`, email, emailType).
		Scan(&c)
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return c, nil
	default:
		return 0, err
	}
}

func InsertUnclaimed(id uuid.UUID, email string, repoId uuid.UUID, balance *big.Int, currency string,
	day time.Time, now time.Time) error {
	stmt, err := dbLib.DB.Prepare(`
			INSERT INTO unclaimed(id, email, repo_id, balance, currency, day, created_at) 
			VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO unclaimed for %v statement event: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, email, repoId, balance.String(), currency, day, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

// ******** Marketing
func FindMarketingEmails() ([]Marketing, error) {
	ms := []Marketing{}
	rows, err := dbLib.DB.Query(`SELECT u.email, u.currency, COALESCE(sum(u.balance), 0) as balances
                        FROM unclaimed u
                        LEFT JOIN git_email g ON u.email = g.email 
                        WHERE g.email IS NULL 
                        GROUP BY u.email, u.currency
                        ORDER BY u.email, u.currency`)

	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(rows)

	mk := Marketing{}
	m := make(map[string]*big.Int)
	mk.Balances = m

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

			mk = Marketing{}
			m = make(map[string]*big.Int)
			mk.Balances = m
			emailOld = email
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

		if emailOld == "" {
			emailOld = email
		}
	}
	mk.Email = email
	ms = append(ms, mk)
	return ms, nil
}
