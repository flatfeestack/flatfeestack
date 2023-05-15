package database

import (
	"database/sql"
	"encoding/json"
	"forum/globals"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
)

type DbUser struct {
	Id     uuid.UUID
	Email  string
	Name   JsonNullString
	Role   string
	Claims jwt.Claims
}

func FindUserByEmail(email string) (*DbUser, error) {
	var u DbUser
	err := globals.DB.
		QueryRow(`SELECT id, email, name
                         FROM users WHERE email=$1`, email).
		Scan(&u.Id, &u.Email, &u.Name)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

// https://stackoverflow.com/a/33072822
type JsonNullString struct {
	sql.NullString
}

func (v JsonNullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}
