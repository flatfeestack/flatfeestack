package util

import (
	"crypto/subtle"
	"fmt"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Credentials struct {
	Username string
	Password string
}

func BasicAuth(cred Credentials, next func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(cred.Username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(cred.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="FlatFeeStack Backend"`)
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("Unauthorized.\n"))

			if err != nil {
				slog.Error("Basic Auth write error", slog.Any("error", err))
			}
			return
		}

		next(w, r)
	}
}

func ValidateJwtInRequest(r *http.Request, jwtKey []byte) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	var bearerToken = ""
	if authHeader == "" {
		return nil, fmt.Errorf("ERR-01, authorization header not set for %v", r.URL)
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

	claims := &jwt.Claims{}

	if tok.Headers[0].Algorithm == string(jose.HS256) {
		err = tok.Claims(jwtKey, claims)
	} else {
		return nil, fmt.Errorf("ERR-04, unknown algorithm: %v", tok.Headers[0].Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("ERR-05, could not parse claims: err=%v for token=%v", err, bearerToken)
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(time.Now().UTC()) {
		return claims, fmt.Errorf("ERR-06, unauthorized: %v", bearerToken)
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("ERR-07, no subject: %v", claims)
	}
	return claims, nil
}
