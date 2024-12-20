package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type Marketing struct {
	Email    string
	Balances map[string]*big.Int
}

func FindGitEmailsByUserId(uid uuid.UUID) ([]GitEmail, error) {
	var gitEmails []GitEmail
	s := "SELECT email, confirmed_at, created_at FROM git_email WHERE user_id=$1"
	rows, err := DB.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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
	err := DB.
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
	stmt, err := DB.Prepare("INSERT INTO git_email(id, user_id, email, token, created_at) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO git_email for %v statement event: %v", email, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, uid, email, token, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func ConfirmGitEmail(email string, token string, now time.Time) error {
	stmt, err := DB.Prepare("UPDATE git_email SET token=NULL, confirmed_at=$1 WHERE email=$2 AND token=$3")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(now, email, token)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func DeleteGitEmail(uid uuid.UUID, email string) error {
	//TODO: don't delete, just mark as deleted
	stmt, err := DB.Prepare("DELETE FROM git_email WHERE email=$1 AND user_id=$2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(email, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func DeleteGitEmailFromUserEmailsSent(uid uuid.UUID, email string) error {
	//TODO: find a better solution in future that allows multiple mails be sent to same mail address but still has spam protection
	stmt, err := DB.Prepare("DELETE FROM user_emails_sent WHERE email=$1 AND user_id=$2 and email_type=$3")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM user_emails_sent for %v and user_id %v statement event: %v", email, uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(email, uid, "add-git"+email)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindUserByGitEmail(gitEmail string) (*uuid.UUID, error) {
	var uid uuid.UUID
	err := DB.
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

func FindUsersByGitEmails(emails []string) ([]uuid.UUID, error) {
	if len(emails) == 0 {
		return nil, nil
	}

	query := `SELECT DISTINCT user_id FROM git_email WHERE email IN (` + GeneratePlaceholders(len(emails)) + `)`
	rows, err := DB.Query(query, ConvertToInterfaceSlice(emails)...)
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

func GetRepoEmails(repoId uuid.UUID) ([]string, error) {
	query := `SELECT DISTINCT git_email FROM analysis_response WHERE repo_id = $1`
	rows, err := DB.Query(query, repoId)
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

//******* Emails Sent

func InsertEmailSent(id uuid.UUID, userId *uuid.UUID, email string, emailType string, now time.Time) error {
	stmt, err := DB.Prepare(`
			INSERT INTO user_emails_sent(id, user_id, email, email_type, created_at) 
			VALUES($1, $2, $3, $4, $5)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_emails_sent for %v statement event: %v", userId, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, userId, email, emailType, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func CountEmailSentById(userId uuid.UUID, emailType string) (int, error) {
	var c int
	err := DB.
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
	err := DB.
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
	stmt, err := DB.Prepare(`
			INSERT INTO unclaimed(id, email, repo_id, balance, currency, day, created_at) 
			VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO unclaimed for %v statement event: %v", email, err)
	}
	defer CloseAndLog(stmt)

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
	rows, err := DB.Query(`SELECT u.email, u.currency, COALESCE(sum(u.balance), 0) as balances
                        FROM unclaimed u
                        LEFT JOIN git_email g ON u.email = g.email 
                        WHERE g.email IS NULL 
                        GROUP BY u.email, u.currency
                        ORDER BY u.email, u.currency`)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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
