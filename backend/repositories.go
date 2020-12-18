package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

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
	err := row.Scan(&user.ID, &user.StripeId, &user.Email, &user.Username, &user.Subscription, &user.SubscriptionState)

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

func FindConnectedEmails(uid string) ([]string, error){
	if uid == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
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

func InsertConnectedEmail(uid string, email string)  error{
	if uid == "" {
		return  fmt.Errorf("ID cannot be empty")
	}
	sqlStatement := `INSERT INTO git_email(email, uid) VALUES($1,$2);`
	_, err := db.Exec(sqlStatement, email, uid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func DeleteConnectedEmail(uid string, email string)  error{
	if uid == "" {
		return  fmt.Errorf("ID cannot be empty")
	}
	sqlStatement := `DELETE FROM git_email WHERE email=$1 AND uid=$2`
	_, err := db.Exec(sqlStatement, email, uid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// FindByID returns a single user
func FindUserByEmail(ID string) (*User, error) {
	var user User

	if ID == "" {
		return &user, fmt.Errorf("ID cannot be empty")
	}

	// create the select sql query
	sqlStatement := `SELECT * FROM "user" WHERE email=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, ID)

	// unmarshal the row object to user
	err := row.Scan(&user.ID, &user.StripeId, &user.Email, &user.Username, &user.Subscription, &user.SubscriptionState)

	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return &user, nil
	case nil:
		return &user, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return &user, err
}

// Save inserts a user into the database
func SaveUser(user *User) error {
	sqlStatement := `INSERT INTO "user" (id, email, username) VALUES ($1, $2, $3) RETURNING id`

	var id string
	err := db.QueryRow(sqlStatement, user.ID, user.Email, user.Username).Scan(&id)

	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Inserted a single record %v", id)

	return nil
}

func UpdateUser(user *User) error{
	sqlStatement := `UPDATE "user" SET (email, username, stripe_id, subscription, "subscription_state") = ($2, $3, $4, $5, $6) 
						WHERE id=$1 RETURNING id`

	var id string
	err := db.QueryRow(sqlStatement, user.ID, user.Email, user.Username, user.StripeId, user.Subscription, user.SubscriptionState).Scan(&id)

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

func Sponsor(repoID int, uid string) (*SponsorEvent, error) {
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

func Unsponsor(repoID int, uid string) (*SponsorEvent, error) {
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

func GetSponsoredReposById(uid string) ([]Repo, error) {
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
		log.Printf("Error inserting payment %v",err)
		return err
	}
	fmt.Printf("Inserted a payment of user%v", paymentId)
	return nil
}