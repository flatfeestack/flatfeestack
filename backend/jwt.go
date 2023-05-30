package main

import (
	"backend/api"
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
			log.Errorf("Token expired: %v, available: %v", claims.Subject, emails)
			utils.WriteErrorf(w, http.StatusUnauthorized, api.GenericErrorMessage)
			return
		} else if claims == nil && err != nil {
			log.Errorf("jwtAuthAdmin error: %v", err)
			utils.WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
			return
		}
		for _, email := range emails {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				next(w, r, email)
				return
			}
		}
		log.Errorf("ERR-01,jwtAuthAdmin error: %v != %v", claims.Subject, emails)
		utils.WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
	}
}

func jwtAuthUser(next func(w http.ResponseWriter, r *http.Request, user *db.UserDetail)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ValidateJwtInRequest(r, jwtKey)
		if claims != nil && err != nil {
			log.Errorf("Token expired: %v", claims.Subject)
			utils.WriteErrorf(w, http.StatusUnauthorized, api.GenericErrorMessage)
			return
		} else if claims == nil && err != nil {
			log.Errorf("jwtAuthAdmin error: %v", err)
			utils.WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
			return
		}

		// Fetch user from DB
		user, err := db.FindUserByEmail(claims.Subject)
		if err != nil {
			log.Errorf("ERR-08, user find error: %v", err)
			utils.WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
			return
		}

		if user == nil {
			user, err = db.CreateUser(claims.Subject, utils.TimeNow())
			if err != nil {
				log.Errorf("ERR-09, user update error: %v", err)
				utils.WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
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
