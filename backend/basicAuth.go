package main

import (
	"crypto/subtle"
	"net/http"
)

func basicAuth(next func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(opts.BackendUsername)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(opts.BackendPassword)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="FlatFeeStack Backend"`)
			w.WriteHeader(401)
			_, err := w.Write([]byte("Unauthorised.\n"))

			if err != nil {
				return
			}

			return
		}

		next(w, r)
	}
}
