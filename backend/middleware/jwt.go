package middleware

import (
	"backend/api"
	"backend/config"
	"backend/util"
	"github.com/go-jose/go-jose/v3/jwt"
	"log/slog"
	"net/http"
)

type JwtHandler struct {
	*config.Config
}

func NewJwtHandler(cfg *config.Config) *JwtHandler {
	return &JwtHandler{Config: cfg}
}

func (j *JwtHandler) JwtAuth(next func(w http.ResponseWriter, r *http.Request, jwt *jwt.Claims)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := util.ValidateJwtInRequest(r, j.JwtKey)
		if claims != nil && err != nil {
			slog.Error("Token expired", slog.String("subject", claims.Subject))
			util.WriteErrorf(w, http.StatusUnauthorized, api.GenericErrorMessage)
			return
		} else if claims == nil && err != nil {
			slog.Error("JwtAuthAdmin error", slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
			return
		}

		slog.Info("User request",
			slog.String("subject", claims.Subject),
			slog.Any("url", r.URL),
			slog.String("method", r.Method))
		next(w, r, claims)
	}
}

func MaxBytes(next func(w http.ResponseWriter, r *http.Request), size int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, size)
		next(w, r)
	}
}
