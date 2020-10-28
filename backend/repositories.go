package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"
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
	err := row.Scan(&user.ID, &user.StripeId, &user.Email, &user.Username)

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
	err := row.Scan(&user.ID, &user.StripeId, &user.Email, &user.Username)

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
	sqlStatement := `INSERT INTO "user" (id, email, "stripe_id", username) VALUES ($1, $2, $3, $4) RETURNING id`

	var id string
	err := db.QueryRow(sqlStatement, user.ID, user.Email, user.StripeId, user.Username).Scan(&id)

	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Inserted a single record %v", id)

	return nil
}

func FindRepoByID(ID string) (*Repo, error) {
	var repo Repo

	if ID == "" {
		return &repo, fmt.Errorf("ID cannot be empty")
	}

	// create the select sql query
	sqlStatement := `SELECT * FROM "repo" WHERE id=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, ID)

	// unmarshal the row object to user
	err := row.Scan(&repo.ID, &repo.Name, &repo.Url)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return &repo, nil
	case nil:
		return &repo, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return &repo, err
}

func SaveRepo(repo *Repo) error {
	sqlStatement := `INSERT INTO "repo" (id, url, name) VALUES ($1, $2, $3) RETURNING id`

	var id string
	err := db.QueryRow(sqlStatement, repo.ID, repo.Url, repo.Name).Scan(&id)

	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Inserted a single record %v", id)

	return nil
}

func Sponsor(repoID string, uid string) (*SponsorEvent, error) {
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

func Unsponsor(repoID string, uid string) (*SponsorEvent, error) {
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
		err = rows.Scan(&repo.ID, &repo.Name, &repo.Url)
		if err != nil {
			fmt.Printf("could not destructure row %v", err)
		}
		repos = append(repos, repo)
	}
	return repos, nil

}

func CalculateDailyByUser(uid string, sponsoredRepos []Repo, amountToShare int) ([]DailyRepoBalance, error) {
	var repoBalances []DailyRepoBalance
	var n = len(sponsoredRepos)
	query := `INSERT INTO "daily_repo_balance" ("repo_id", "uid", "computed_at", "balance") VALUES`
	var values []interface{}
	for i, s := range sponsoredRepos {
		values = append(values, s.ID, uid, time.Now(), math.Floor(float64(amountToShare/n)))

		numFields := 4
		n := i * numFields

		query += `(`
		for j := 0; j < numFields; j++ {
			query += `$` + strconv.Itoa(n+j+1) + `,`
		}
		query = query[:len(query)-1] + `),`
	}
	query = query[:len(query)-1] // remove the trailing comma
	query += ` RETURNING "id", "repo_id", "uid", "computed_at", "balance"`
	rows, err := db.Query(query, values...)

	if err != nil {
		fmt.Printf("error executing query %v", err)
		return repoBalances, err
	}

	defer rows.Close()

	for rows.Next() {
		var dailyBalance DailyRepoBalance
		err = rows.Scan(&dailyBalance.ID, &dailyBalance.RepoId, &dailyBalance.Uid, &dailyBalance.ComputedAt, &dailyBalance.Balance)
		if err != nil {
			fmt.Printf("\ncould not destructure row %v", err)
		}
		repoBalances = append(repoBalances, dailyBalance)
	}

	return repoBalances, nil
}
