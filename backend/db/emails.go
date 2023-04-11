package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

type Marketing struct {
	Email      string
	RepoIds    []uuid.UUID
	Balances   []string
	Currencies []string
}

func FindGitEmailsByUserId(uid uuid.UUID) ([]GitEmail, error) {
	var gitEmails []GitEmail
	s := "SELECT email, confirmed_at, created_at FROM git_email WHERE user_id=$1"
	rows, err := db.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

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

func InsertGitEmail(uid uuid.UUID, email string, token *string, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO git_email(user_id, email, token, created_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO git_email for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uid, email, token, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func ConfirmGitEmail(email string, token string, now time.Time) error {
	stmt, err := db.Prepare("UPDATE git_email SET token=NULL, confirmed_at=$1 WHERE email=$2 AND token=$3")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(now, email, token)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func DeleteGitEmail(uid uuid.UUID, email string) error {
	//TODO: don't delete, just mark as deleted
	stmt, err := db.Prepare("DELETE FROM git_email WHERE email=$1 AND user_id=$2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(email, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindUserByGitEmail(gitEmail string) (*uuid.UUID, error) {
	var uid uuid.UUID
	err := db.
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

func InsertEmailSent(userId *uuid.UUID, email string, emailType string, now time.Time) error {
	stmt, err := db.Prepare(`
			INSERT INTO user_emails_sent(user_id, email, email_type, created_at) 
			VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_emails_sent for %v statement event: %v", userId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(userId, email, emailType, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func CountEmailSentById(userId uuid.UUID, emailType string) (int, error) {
	var c int
	err := db.
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
	err := db.
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

//******** Marketing

func FindMarketingEmails() ([]Marketing, error) {
	ms := []Marketing{}
	rows, err := db.Query(`SELECT u.email, array_agg(DISTINCT u.repo_id) as repo_ids, array_agg(u.balance) as balances,  array_agg(u.currency) as currencies
                        FROM unclaimed u
                        LEFT JOIN git_email g ON u.email = g.email 
                        WHERE g.email IS NULL 
                        GROUP BY u.email`)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var m Marketing
		err = rows.Scan(&m.Email, pq.Array(&m.RepoIds), pq.Array(&m.Balances), pq.Array(&m.Currencies))
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, nil
}
