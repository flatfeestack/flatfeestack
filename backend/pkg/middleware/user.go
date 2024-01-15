package util

import (
	"backend/internal/api"
	"backend/internal/db"
	"backend/pkg/config"
	"backend/pkg/util"
	"github.com/go-jose/go-jose/v3/jwt"
	"log/slog"
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
			slog.Error("User find error",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
			return
		}

		if user == nil {
			user, err = db.CreateUser(claims.Subject, util.TimeNow())
			if err != nil {
				slog.Error("User update error",
					slog.Any("error", err))
				util.WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
				return
			}
		}

		//User exists now, check if we are admin
		role := "user"
		for _, email := range j.AdminsParsed {
			if claims.Subject == email {
				slog.Info("Authenticated admin",
					slog.String("email", email))
				role = "admin"
			}
		}
		user.Role = role
		user.Claims = claims
		next(w, r, user)
	}
}
