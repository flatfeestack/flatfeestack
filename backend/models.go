package main

import (
	"database/sql"
	"time"
	"encoding/json"

)

// NullString is an alias for sql.NullString data type
type NullString struct {
	sql.NullString
}

// MarshalJSON for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// User schema of the user table
type User struct {
	ID       string `json:"id"`
	StripeId NullString `json:"-"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Subscription NullString `json:"subscription"`
}

// Swaggo does not support sql.Nullstring
type UserDTO struct {
	ID       string `json:"id"`
	StripeId string `json:"-"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Subscription string `json:"subscription"`
}

type UserRepository interface {
	FindByID(ID string) (*User, error)
	FindByEmail(email string) (*User, error)
	Save(user *User) error
	UpdateUser(user *User) (*User, error)
}

type Repo struct {
	ID   string `json:"id"`
	Url  string `json:"url"`
	Name string `json:"name"`
}

type RepoRepository interface {
	FindByID(ID string) (*Repo, error)
	Save(repo *Repo) error
}

type SponsorEvent struct {
	ID        string `json:"id"`
	Uid       string `json:"uid"`
	RepoId    string `json:"repo_id"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
}

type SponsorEventRepository interface {
	Sponsor(repoID string, uid string) (*SponsorEvent, error)
	Unsponsor(repoID string, uid string) (*SponsorEvent, error)
	GetSponsoredRepos(uid string) ([]Repo, error)
}

type DailyRepoBalance struct {
	ID         int       `json:"id"`
	RepoId     string    `json:"repo_id"`
	Uid        string    `json:"uid"`
	ComputedAt time.Time `json:"computed_at"`
	Balance    int       `json:"balance"`
}

type DailyRepoBalanceRepository interface {
	CalculateDailyByUser(uid string, sponsoredRepos []Repo, amountToShare int) ([]DailyRepoBalance, error)
}
