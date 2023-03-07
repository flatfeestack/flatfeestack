package main

import (
	"fmt"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type TokenClaims struct {
	jwt.Claims
}

func jwtAuthAdmin(next func(w http.ResponseWriter, r *http.Request, email string), emails []string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwtAuth(r)
		if claims != nil && err != nil {
			writeErrorf(w, http.StatusUnauthorized, "Token expired: %v, available: %v", claims.Subject, emails)
			return
		} else if claims == nil && err != nil {
			writeErrorf(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
			return
		}
		for _, email := range emails {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				next(w, r, email)
				return
			}
		}
		writeErrorf(w, http.StatusBadRequest, "ERR-01,jwtAuthAdmin error: %v != %v", claims.Subject, emails)
	}
}

func jwtAuthUser(next func(w http.ResponseWriter, r *http.Request, user *User)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwtAuth(r)

		if claims != nil && err != nil {
			if r.Header.Get("Sec-WebSocket-Protocol") == "" {
				//no websocket
				writeErrorf(w, http.StatusUnauthorized, "Token expired: %v", claims.Subject)
			} else {
				//we use websocket
				wsNoAuth(w, r)
			}
			return
		} else if claims == nil && err != nil {
			writeErrorf(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
			return
		}

		unlock := km.Lock(claims.Subject)
		defer unlock()

		// Fetch user from DB
		user, err := findUserByEmail(claims.Subject)
		if err != nil {
			writeErrorf(w, http.StatusBadRequest, "ERR-08, user find error: %v", err)
			return
		}

		if user == nil {
			user, err = createUser(claims.Subject)
			if err != nil {
				writeErrorf(w, http.StatusBadRequest, "ERR-09, user update error: %v", err)
				return
			}
		}

		//User exists now, check if we are admin
		for _, email := range admins {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				user.Role = stringPointer("admin")
			}
		}

		user.Claims = claims
		log.Printf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method)
		next(w, r, user)
	}
}

func jwtAuth(r *http.Request) (*TokenClaims, error) {
	authHeader := r.Header.Get("Authorization")
	var bearerToken = ""
	if authHeader == "" {
		authHeader = r.Header.Get("Sec-WebSocket-Protocol")
		if authHeader == "" {
			return nil, fmt.Errorf("ERR-01, authorization header not set for %v", r.URL)
		}
	}
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		return nil, fmt.Errorf("ERR-02, could not split token: %v", bearerToken)
	}
	bearerToken = split[1]

	tok, err := jwt.ParseSigned(bearerToken)
	if err != nil {
		return nil, fmt.Errorf("ERR-03, could not parse token: %v", bearerToken[1])
	}

	claims := &TokenClaims{}

	if tok.Headers[0].Algorithm == string(jose.RS256) {
		err = tok.Claims(privRSA.Public(), claims)
	} else if tok.Headers[0].Algorithm == string(jose.HS256) {
		err = tok.Claims(jwtKey, claims)
	} else if tok.Headers[0].Algorithm == string(jose.EdDSA) {
		err = tok.Claims(privEdDSA.Public(), claims)
	} else {
		return nil, fmt.Errorf("ERR-04, unknown algorithm: %v", tok.Headers[0].Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("ERR-05, could not parse claims: err=%v for token=%v", err, bearerToken)
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(timeNow()) {
		return claims, fmt.Errorf("ERR-06, unauthorized: %v", bearerToken)
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("ERR-07, no subject: %v", claims)
	}
	return claims, nil
}
