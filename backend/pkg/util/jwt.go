package util

import (
	"backend/api"
	"backend/pkg/config"
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

func (j *JwtHandler) JwtAuthAdmin(next func(w http.ResponseWriter, r *http.Request, email string), emails []string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ValidateJwtInRequest(r, j.JwtKey)
		if claims != nil && err != nil {
			log.Errorf("Token expired: %v, available: %v", claims.Subject, emails)
			WriteErrorf(w, http.StatusUnauthorized, api.GenericErrorMessage)
			return
		} else if claims == nil && err != nil {
			log.Errorf("jwtAuthAdmin error: %v", err)
			WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
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
		WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
	}
}

func (j *JwtHandler) JwtAuthUser(next func(w http.ResponseWriter, r *http.Request, jwt *jwt.Claims)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ValidateJwtInRequest(r, j.JwtKey)
		if claims != nil && err != nil {
			log.Errorf("Token expired: %v", claims.Subject)
			WriteErrorf(w, http.StatusUnauthorized, api.GenericErrorMessage)
			return
		} else if claims == nil && err != nil {
			log.Errorf("jwtAuthAdmin error: %v", err)
			WriteErrorf(w, http.StatusBadRequest, api.NotAllowedToViewMessage)
			return
		}

		log.Printf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method)
		next(w, r, claims)

		// Fetch user from DB
		/*user, err := db.FindUserByEmail(claims.Subject)
		if err != nil {
			log.Errorf("ERR-08, user find error: %v", err)
			WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
			return
		}

		if user == nil {
			user, err = db.CreateUser(claims.Subject, TimeNow())
			if err != nil {
				log.Errorf("ERR-09, user update error: %v", err)
				WriteErrorf(w, http.StatusBadRequest, api.GenericErrorMessage)
				return
			}
		}*/

		//User exists now, check if we are admin
		/*role := "user"
		for _, email := range j.Admins {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				role = "admin"
			}
		}

		user.Claims = claims*/

	}
}

func MaxBytes(next func(w http.ResponseWriter, r *http.Request), size int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, size)
		next(w, r)
	}
}
