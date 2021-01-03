package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"time"
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

func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ns.String = s
	ns.Valid = true
	return nil
}

// NullTime is an alias for pq.NullTime data type
type NullTime struct {
	pq.NullTime
}

// JSON handling for NullString
func (ns *NullTime) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.Time)
}

func (ns *NullTime) UnmarshalJSON(data []byte) error {
	var s time.Time
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ns.Time = s
	ns.Valid = true
	return nil
}

// User schema of the user table

// User schema of the user table
type UserWithConnectedEmails struct {
	ID                string     `json:"id"`
	StripeId          NullString `json:"-"`
	Email             string     `json:"email"`
	Username          string     `json:"username"`
	Subscription      NullString `json:"subscription"`
	SubscriptionState NullString `json:"subscription_state"`
	ConnectedEmails   []string   `json:"connected_emails"`
}

// Swaggo does not support sql.Nullstring
type UserDTO struct {
	ID           string `json:"id"`
	StripeId     string `json:"-"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Subscription string `json:"subscription"`
}

type RepoDTO struct {
	ID          int32  `json:"id"`
	Url         string `json:"html_url"`
	Name        string `json:"full_name"`
	Description string `json:"description"`
}

type DailyRepoBalance struct {
	ID         int       `json:"id"`
	RepoId     string    `json:"repo_id"`
	Uid        string    `json:"uid"`
	ComputedAt time.Time `json:"computed_at"`
	Balance    int       `json:"balance"`
}

type Payment struct {
	Uid    string
	Amount int64
	From   time.Time
	To     time.Time
	Sub    string
}

type WebhookRequest struct {
	RepositoryUrl       string `json:"repository_url"`
	Since               string `json:"since"`
	Until               string `json:"until"`
	PlatformInformation bool   `json:"platform_information"`
	Branch              string `json:"branch"`
}

type WebhookResponse struct {
	RequestId string `json:"request_id"`
}

type ExchangeEntry struct {
	ID int `json:"id"`
	// DB Driver will convert numeric into a string,
	// which is fine since we don't do calculations on the amount but only display it
	Amount  string     `json:"amount"`
	ChainId string     `json:"chain_id"`
	Date    NullTime   `json:"date"`
	Price   NullString `json:"price"`
}

type ExchangeEntryUpdate struct {
	ID    int       `json:"id"`
	Date  time.Time `json:"date"`
	Price string    `json:"price"`
}
