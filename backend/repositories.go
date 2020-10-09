package main

import (
	"database/sql"
	"fmt"
	"log"
)

/*
 *	==== USER ====
 */

// UserRepo implements UserRepository
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}
// FindByID returns a single user
func (r *UserRepo) FindByID(ID string) (*User, error) {
	var user User

	if ID == "" {
		return &user, fmt.Errorf("ID cannot be empty")
	}

	// create the select sql query
	sqlStatement := `SELECT * FROM "user" WHERE id=$1`

	// execute the sql statement
	row := r.db.QueryRow(sqlStatement, ID)

	// unmarshal the row object to user
	err := row.Scan(&user.ID, &user.StripeId, &user.Email)

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

// Save inserts a user into the database
func (r *UserRepo) Save(user *User) error {
	sqlStatement := `INSERT INTO "user" (id, email, "stripe_id") VALUES ($1, $2, $3) RETURNING id`

	var id string
	err := r.db.QueryRow(sqlStatement, user.ID, user.Email, user.StripeId).Scan(&id)

	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Inserted a single record %v", id)

	return nil
}

/*
 *	==== REPO ====
 */

// RepoRepo implements RepoRepository
type RepoRepo struct {
	db *sql.DB
}

func NewRepoRepo(db *sql.DB) *RepoRepo {
	return &RepoRepo{
		db: db,
	}
}

func (r *RepoRepo) FindByID(ID string) (*Repo, error){
	var repo Repo
	log.Println("not implemented yet, return empty repo")
	return &repo, nil
}

func (r *RepoRepo) Save(repo *Repo) error{
	sqlStatement := `INSERT INTO "repo" (id, url, name) VALUES ($1, $2, $3) RETURNING id`

	var id string
	err := r.db.QueryRow(sqlStatement, repo.ID, repo.Url, repo.Name).Scan(&id)

	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Inserted a single record %v", id)

	return nil
}