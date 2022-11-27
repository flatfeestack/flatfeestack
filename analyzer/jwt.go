package main

import (
	"github.com/go-jose/go-jose/v3/jwt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type TokenClaims struct {
	jwt.Claims
}

func jwtAuth(next func(w http.ResponseWriter, r *http.Request, claims *TokenClaims)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErr(w, http.StatusBadRequest, "jwtAuth, authorization header not set")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			writeErr(w, http.StatusBadRequest, "jwtAuth, could not split token: %v", bearerToken)
			return
		}

		tok, err := jwt.ParseSigned(bearerToken[1])
		if err != nil {
			writeErr(w, http.StatusBadRequest, "jwtAuth, could not parse token: %v", bearerToken[1])
			return
		}

		claims := &TokenClaims{}

		if tok.Headers[0].Algorithm == "HS256" {
			err = tok.Claims(jwtKey, claims)
		} else {
			writeErr(w, http.StatusUnauthorized, "jwtAuth, unknown algorithm: %v", tok.Headers[0].Algorithm)
			return
		}

		if err != nil {
			writeErr(w, http.StatusUnauthorized, "jwtAuth, could not parse claims: %v", bearerToken[1])
			return
		}

		if claims == nil {
			writeErr(w, http.StatusBadRequest, "jwtAuth, claims are empty")
			return
		}

		if claims.Expiry != nil && !claims.Expiry.Time().After(time.Now()) {
			writeErr(w, http.StatusBadRequest, "jwtAuth, expired: %v", claims.Expiry.Time())
			return
		}

		next(w, r, claims)
	}
}

func jwtAuthServer(next func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request, *TokenClaims) {
	return func(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
		if claims.Subject == "ffs-server" {
			log.Printf("Authenticated server %s\n")
			next(w, r)
			return
		}

		writeErr(w, http.StatusBadRequest, "ERR-01,jwtAuthAdmin error: %v", claims.Subject)
	}
}
