package main

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
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
