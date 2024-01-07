package util

import (
	"backend/internal/api"
	"backend/pkg/config"
	"backend/pkg/util"
	"github.com/flatfeestack/go-lib/auth"
	"github.com/go-jose/go-jose/v3/jwt"
	log "github.com/sirupsen/logrus"
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
		claims, err := auth.ValidateJwtInRequest(r, j.JwtKey)
		if claims != nil && err != nil {
			log.Errorf("Token expired: %v", claims.Subject)
			util.WriteErrorf(w, http.StatusUnauthorized, api.GenericErrorMessage)
			return
		} else if claims == nil && err != nil {
			log.Errorf("jwtAuthAdmin error: %v", err)
			util.WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
			return
		}

		log.Printf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method)
		next(w, r, claims)

	}
}

func MaxBytes(next func(w http.ResponseWriter, r *http.Request), size int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, size)
		next(w, r)
	}
}
