package main

import (
	"backend/db"
	"backend/utils"
	"github.com/flatfeestack/go-lib/auth"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func jwtAuthAdmin(next func(w http.ResponseWriter, r *http.Request, email string), emails []string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ValidateJwtInRequest(r, jwtKey)
		if claims != nil && err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "Token expired: %v, available: %v", claims.Subject, emails)
			return
		} else if claims == nil && err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
			return
		}
		for _, email := range emails {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				next(w, r, email)
				return
			}
		}
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-01,jwtAuthAdmin error: %v != %v", claims.Subject, emails)
	}
}

func jwtAuthUser(next func(w http.ResponseWriter, r *http.Request, user *db.UserDetail)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ValidateJwtInRequest(r, jwtKey)
		if claims != nil && err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "Token expired: %v", claims.Subject)
			return
		} else if claims == nil && err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
			return
		}

		// Fetch user from DB
		user, err := db.FindUserByEmail(claims.Subject)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "ERR-08, user find error: %v", err)
			return
		}

		if user == nil {
			user, err = db.CreateUser(claims.Subject, utils.TimeNow())
			if err != nil {
				utils.WriteErrorf(w, http.StatusBadRequest, "ERR-09, user update error: %v", err)
				return
			}
		}

		//User exists now, check if we are admin
		for _, email := range admins {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				user.Role = utils.StringPointer("admin")
			}
		}

		user.Claims = claims
		log.Printf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method)
		next(w, r, user)
	}
}
