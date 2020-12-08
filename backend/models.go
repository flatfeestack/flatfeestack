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

// JSON handling for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(data []byte)  error {
	var s string
	if err := json.Unmarshal(data, &s); err !=nil {
		return err
	}
	ns.String = s
	ns.Valid = true
	return nil
}

// User schema of the user table
type User struct {
	ID       string `json:"id"`
	StripeId NullString `json:"-"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Subscription NullString `json:"subscription"`
	SubscriptionState NullString `json:"subscription_state"`
}

// Swaggo does not support sql.Nullstring
type UserDTO struct {
	ID       string `json:"id"`
	StripeId string `json:"-"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Subscription string `json:"subscription"`
}

type RepoDTO struct {
	ID   int32 `json:"id"`
	Url  string `json:"html_url"`
	Name string `json:"full_name"`
	Description string `json:"description"`
}

type Repo struct {
	ID   int32 `json:"id"`
	Url  string `json:"html_url"`
	Name string `json:"full_name"`
	Description NullString `json:"description"`
}

// TODO: change timestampe to time.Time
type SponsorEvent struct {
	ID        string `json:"id"`
	Uid       string `json:"uid"`
	RepoId    string `json:"repo_id"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
}


type DailyRepoBalance struct {
	ID         int       `json:"id"`
	RepoId     string    `json:"repo_id"`
	Uid        string    `json:"uid"`
	ComputedAt time.Time `json:"computed_at"`
	Balance    int       `json:"balance"`
}


type Payment struct {
	Uid string
	Amount int64
	From time.Time
	To time.Time
	Sub string
}