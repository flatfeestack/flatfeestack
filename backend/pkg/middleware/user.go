package util

import (
	"backend/internal/api"
	"backend/internal/db"
	"backend/pkg/config"
	"backend/pkg/util"
	"github.com/go-jose/go-jose/v3/jwt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type JwtUserHandler struct {
	*config.Config
}

func NewJwtUserHandler(cfg *config.Config) *JwtUserHandler {
	return &JwtUserHandler{Config: cfg}
}

func (j *JwtUserHandler) JwtUser(next func(w http.ResponseWriter, r *http.Request, u *db.UserDetail)) func(http.ResponseWriter, *http.Request, *jwt.Claims) {
	return func(w http.ResponseWriter, r *http.Request, claims *jwt.Claims) {
		// Fetch user from DB
		user, err := db.FindUserByEmail(claims.Subject)
		if err != nil {
			log.Errorf("ERR-08, user find error: %v", err)
			util.WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
			return
		}

		if user == nil {
			user, err = db.CreateUser(claims.Subject, util.TimeNow())
			if err != nil {
				log.Errorf("ERR-09, user update error: %v", err)
				util.WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
				return
			}
		}

		//User exists now, check if we are admin
		role := "user"
		for _, email := range j.AdminsParsed {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				role = "admin"
			}
		}
		user.Role = role
		user.Claims = claims
		next(w, r, user)
	}
}
func JwtAdmin(next func(w http.ResponseWriter, r *http.Request, u *db.UserDetail)) func(http.ResponseWriter, *http.Request, *db.UserDetail) {
	return func(w http.ResponseWriter, r *http.Request, u *db.UserDetail) {
		if u.Role != "admin" {
			log.Errorf("not admin : %v", u.Email)
			util.WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
			return
		}
		next(w, r, u)
	}
}
