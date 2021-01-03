package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

type User struct {
	Id                uuid.UUID `json:"id" sql:",type:uuid"`
	StripeId          *string   `json:"-"`
	Email             *string   `json:"email"`
	Subscription      *string   `json:"subscription"`
	SubscriptionState *string   `json:"subscription_state"`
	PayoutETH         *string   `json:"payout_eth"`
}

type SponsorEvent struct {
	Id        uuid.UUID `json:"id"`
	Uid       uuid.UUID `json:"uid"`
	RepoId    uuid.UUID `json:"repo_id"`
	EventType uint8     `json:"event_type"`
	CreatedAt time.Time `json:"created_at"`
}

type Repo struct {
	Id          uuid.UUID `json:"id"`
	OrigId      int
	OrigFrom    *string
	Url         *string `json:"html_url"`
	Name        *string `json:"full_name"`
	Description *string `json:"description"`
}

// FindByID returns a single user
func FindUserByEmail(email string) (*User, error) {
	var u User
	err := db.
		QueryRow("SELECT id, stripe_id, email, subscription, subscription_state, payout_eth FROM users WHERE email=$1", email).
		Scan(&u.Id, &u.StripeId, &u.Email, &u.Subscription, &u.SubscriptionState, &u.PayoutETH)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

// FindByID returns a single user
func FindUserByID(uid uuid.UUID) (*User, error) {
	var u User
	err := db.
		QueryRow("SELECT id, stripe_id, email, subscription, subscription_state, payout_eth FROM users WHERE id=$1", uid).
		Scan(&u.Id, &u.StripeId, &u.Email, &u.Subscription, &u.SubscriptionState, &u.PayoutETH)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

// Save inserts a user into the database
func SaveUser(user *User) error {
	stmt, err := db.Prepare("INSERT INTO users (id, email, stripe_id, payout_eth) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(user.Id, user.Email, user.StripeId, user.PayoutETH)
	return handleErr(res, err, "INSERT INTO auth", user)
}

func UpdateUser(user *User) error {
	stmt, err := db.Prepare("UPDATE users SET email=$1, stripe_id=$2, subscription=$3, subscription_state=$4, payout_eth=$5 WHERE id=$6")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", user, err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(user.Email, user.StripeId, user.Subscription, user.SubscriptionState, user.PayoutETH, user.Id)
	return handleErr(res, err, "UPDATE users", user)
}

//sponsor events
func Sponsor(event *SponsorEvent) error {
	stmt, err := db.Prepare("INSERT INTO sponsor_event (id, user_id, repo_id, event_type, created_at) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO sponsor_event for %v statement event: %v", event, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(event.Id, event.Uid, event.RepoId, event.EventType, event.CreatedAt)
	return handleErr(res, err, "INSERT INTO sponsor_event", event)
}

// Repositories
func GetSponsoredReposById(uid uuid.UUID, eventType uint8) ([]Repo, error) {
	var repos []Repo
	sql := `SELECT r.id, r.orig_id, r.orig_from, r.url, r.name, r.description 
			FROM (SELECT event_type, repo_id, max(created_at) as created_at FROM sponsor_event WHERE user_id=$1 GROUP BY repo_id) as s 
			JOIN repo r on r.id = s.repo_id AND s.event_type=$2`
	rows, err := db.Query(sql, uid, eventType)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var repo Repo
		err = rows.Scan(&repo.Id, &repo.OrigId, &repo.OrigFrom, &repo.Url, &repo.Name, &repo.Description)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func SaveRepo(repo *Repo) error {
	stmt, err := db.Prepare("INSERT INTO repo (id, orig_id, orig_from, url, name, description) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(repo.Id, repo.OrigId, repo.OrigFrom, repo.Url, repo.Name, repo.Description)
	return handleErr(res, err, "INSERT INTO repo", repo)
}

func FindRepoByID(rid uuid.UUID) (*Repo, error) {
	var r Repo
	err := db.
		QueryRow("SELECT id, orig_id, orig_from, url, name, description FROM repo WHERE id=$1", rid).
		Scan(&r.Id, &r.OrigId, &r.OrigFrom, &r.Name, &r.Url, &r.Description)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &r, nil
	default:
		return nil, err
	}
}

//connected emails
func FindGitEmails(uid uuid.UUID) ([]string, error) {
	var emails []string
	sql := "SELECT email FROM git_email WHERE user_id=$1"
	rows, err := db.Query(sql, uid)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var email string
		err = rows.Scan(&email)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func SaveGitEmail(id uuid.UUID, uid uuid.UUID, email string) error {
	stmt, err := db.Prepare("INSERT INTO git_email(id, user_id, email) VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO git_email for %v statement event: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id, uid, email)
	return handleErr(res, err, "INSERT INTO git_email", email)
}

func DeleteGitEmail(uid uuid.UUID, email string) error {
	stmt, err := db.Prepare("DELETE FROM git_email WHERE email=$1 AND user_id=$2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, uid)
	return handleErr(res, err, "DELETE FROM git_email", email)
}

func handleErr(res sql.Result, err error, info string, value interface{}) error {
	if err != nil {
		return fmt.Errorf("%v query %v failed: %v", info, value, err)
	}
	nr, err := res.RowsAffected()
	if nr == 0 || err != nil {
		return fmt.Errorf("%v %v rows %v, affected or err: %v", info, nr, value, err)
	}
	return nil
}


func SavePayment(payment *Payment) error {
	var paymentId int
	sqlStatement := `INSERT INTO "payments" ("uid", "from", "to", "sub", "amount") VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(sqlStatement, payment.Uid, payment.From, payment.To, payment.Sub, payment.Amount).Scan(&paymentId)
	if err != nil {
		log.Printf("Error inserting payment %v", err)
		return err
	}
	fmt.Printf("Inserted a payment of user%v", paymentId)
	return nil
}

func SelectExchanges() ([]ExchangeEntry, error) {
	var exchanges []ExchangeEntry

	sqlStatement := `SELECT "id", "amount", "chain_id", "price", "date" FROM exchange`
	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var exchange ExchangeEntry
		err = rows.Scan(&exchange.ID, &exchange.Amount, &exchange.ChainId, &exchange.Price, &exchange.Date)
		if err != nil {
			fmt.Printf("could not destructure row %v", err)
		}
		exchanges = append(exchanges, exchange)
	}
	return exchanges, nil
}

func UpdateExchange(ex ExchangeEntryUpdate) error {
	sqlStatement := `UPDATE "exchange" SET ("date", "price") = ($2, $3) 
						WHERE id=$1 RETURNING id`

	_, err := db.Exec(sqlStatement, ex.ID, ex.Date, ex.Price)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
