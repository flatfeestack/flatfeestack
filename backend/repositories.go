package main

import (
	"database/sql"
	"errors"
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

// FindByID returns a single user
func FindUserByEmail(email string) (*User, error) {
	var u User
	err := db.
		QueryRow("SELECT id, stripe_id, email, subscription, subscription_state, payout_eth FROM users WHERE email=$1", email).
		Scan(&u.Id, &u.StripeId, &u.Email, &u.Subscription, &u.SubscriptionState, &u.PayoutETH)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &u, nil
	}
}

// Save inserts a user into the database
func SaveUser(user *User) error {
	stmt, err := db.Prepare("INSERT INTO users (id, email) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(user.Id, user.Email)
	return handleErr(res, err, "INSERT INTO auth", user)
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

// FindByID returns a single user
func FindUserByID(ID string) (*User, error) {
	var user User

	if ID == "" {
		return &user, fmt.Errorf("ID cannot be empty")
	}

	// create the select sql query
	sqlStatement := `SELECT * FROM "user" WHERE id=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, ID)

	// unmarshal the row object to user
	err := row.Scan(&user.Id, &user.StripeId, &user.Email, &user.Subscription, &user.SubscriptionState)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return &user, nil
	case nil:
		return &user, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return &user, err
}

func FindConnectedEmails(uid uuid.UUID) ([]string, error) {

	var emails []string

	sqlStatement := `SELECT email FROM git_email WHERE uid=$1;`
	rows, err := db.Query(sqlStatement, uid)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var email string
		err = rows.Scan(&email)
		if err != nil {
			fmt.Printf("could not destructure row %v", err)
		}
		emails = append(emails, email)
	}
	// return empty user on error
	return emails, err
}

func InsertConnectedEmail(uid uuid.UUID, email string) error {
	sqlStatement := `INSERT INTO git_email(email, uid) VALUES($1,$2);`
	_, err := db.Exec(sqlStatement, email, uid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func DeleteConnectedEmail(uid uuid.UUID, email string) error {
	sqlStatement := `DELETE FROM git_email WHERE email=$1 AND uid=$2`
	_, err := db.Exec(sqlStatement, email, uid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func UpdateUser(user *User) error {
	sqlStatement := `UPDATE "user" SET (email, stripe_id, subscription, "subscription_state") = ($2, $3, $4, $5, $6) 
						WHERE id=$1 RETURNING id`

	var id string
	err := db.QueryRow(sqlStatement, user.Id, user.Email, user.StripeId, user.Subscription, user.SubscriptionState).Scan(&id)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func FindRepoByID(ID int) (*Repo, error) {
	var repo Repo

	// create the select sql query
	sqlStatement := `SELECT * FROM "repo" WHERE id=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, ID)

	// unmarshal the row object to user
	err := row.Scan(&repo.ID, &repo.Url, &repo.Name, &repo.Description)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, errors.New("Not found")
	case nil:
		return &repo, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return &repo, err
}

func SaveRepo(repo *Repo) error {
	sqlStatement := `INSERT INTO "repo" (id, url, name, description) VALUES ($1, $2, $3, $4) RETURNING id`

	var id string
	err := db.QueryRow(sqlStatement, repo.ID, repo.Url, repo.Name, repo.Description).Scan(&id)

	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Inserted a single record %v", id)

	return nil
}

func Sponsor(repoID int, uid uuid.UUID) (*SponsorEvent, error) {
	var event SponsorEvent
	sqlStatement := `INSERT INTO "sponsor_event" (uid, repo_id, type, timestamp) VALUES ($1, $2, $3, $4) RETURNING id, uid, repo_id, type, timestamp`
	err := db.QueryRow(sqlStatement, uid, repoID, "SPONSOR", time.Now().Unix()).Scan(&event.ID, &event.Uid, &event.RepoId, &event.Type, &event.Timestamp)

	if err != nil {
		log.Println(err)
		return &event, err
	}

	fmt.Printf("Inserted a single record %v", &event.ID)

	return &event, nil
}

func Unsponsor(repoID int, uid uuid.UUID) (*SponsorEvent, error) {
	var event SponsorEvent
	sqlStatement := `INSERT INTO "sponsor_event" (uid, repo_id, type, timestamp) VALUES ($1, $2, $3, $4) RETURNING id, uid, repo_id, type, timestamp`
	err := db.QueryRow(sqlStatement, uid, repoID, "UNSPONSOR", time.Now().Unix()).Scan(&event.ID, &event.Uid, &event.RepoId, &event.Type, &event.Timestamp)

	if err != nil {
		log.Println(err)
		return &event, err
	}

	fmt.Printf("Inserted a single record %v", &event.ID)

	return &event, nil
}

func GetSponsoredReposById(uid uuid.UUID) ([]Repo, error) {
	var repos []Repo
	sqlStatement := `SELECT r.* FROM 
		(SELECT uid, repo_id, max("timestamp") as "timestamp" 
			FROM sponsor_event 
			GROUP BY uid, repo_id) as latest 
		JOIN sponsor_event s on latest.uid = s.uid AND latest.repo_id = s.repo_id AND latest.timestamp = s."timestamp"
		JOIN repo r on r.id = s.repo_id
		WHERE s."type" = 'SPONSOR' AND s.uid = $1`
	rows, err := db.Query(sqlStatement, uid)

	if err != nil {
		log.Println(err)
		return repos, err
	}

	defer rows.Close()

	for rows.Next() {
		var repo Repo
		err = rows.Scan(&repo.ID, &repo.Url, &repo.Name, &repo.Description)
		if err != nil {
			fmt.Printf("could not destructure row %v", err)
		}
		repos = append(repos, repo)
	}
	return repos, nil

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

func InsertOrUpdatePayoutAddress(p PayoutAddress) error {
	sqlStatement := `INSERT INTO pay_out_address(uid, address, chain_id) VALUES($1, $2, $3) ON CONFLICT(uid,chain_id) DO UPDATE SET address = EXCLUDED.address`
	_, err := db.Exec(sqlStatement, p.Uid, p.Address, p.ChainId)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SelectPayoutAddressesByUid(uid uuid.UUID) (addresses []PayoutAddress, err error) {
	sqlStatement := `SELECT uid, address, chain_id FROM pay_out_address WHERE uid=$1`
	rows, err := db.Query(sqlStatement, uid)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address PayoutAddress
		err = rows.Scan(&address.Uid, &address.Address, &address.ChainId)
		if err != nil {
			fmt.Printf("could not destructure row %v", err)
		}
		addresses = append(addresses, address)
	}
	return
}
