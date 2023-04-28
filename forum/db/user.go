package database

import (
	"database/sql"
	"forum/globals"
	"forum/types"
	"github.com/google/uuid"
)

type DbUser struct {
	Id     uuid.UUID
	Email  string
	Name   string
	Role   string
	Claims types.TokenClaims
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
