package types

import (
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
)

type Opts struct {
	Port               int
	HS256              string
	Env                string
	DBPath             string
	DBDriver           string
	DBScripts          string
	Admins             string
	BackendUrl         string
	BackendUsername    string
	BackendPassword    string
	DaoContractAddress string
	EthWsUrl           string
}

type User struct {
	Id     uuid.UUID `json:"id"`
	Email  string    `json:"email"`
	Name   string    `json:"name,omitempty"`
	Claims jwt.Claims
	Role   string `json:"role,omitempty"`
}
