package types

import "github.com/go-jose/go-jose/v3/jwt"

type Opts struct {
	Port      int
	HS256     string
	Env       string
	DBPath    string
	DBDriver  string
	DBScripts string
	Admins    string
}

type TokenClaims struct {
	jwt.Claims
}
