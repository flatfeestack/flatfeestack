package jwt

import (
	"context"
	"fmt"
	"forum/api"
	database "forum/db"
	"forum/globals"
	"forum/utils"
	"github.com/flatfeestack/go-lib/auth"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	CurrentUser = "currentUser"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the context from the request
		ctx := r.Context()

		scopes := ctx.Value(api.BearerAuthScopes)

		// Route without auth
		if scopes == nil {
			next.ServeHTTP(w, r)
			return
		}

		scopesSlice, ok := scopes.([]string)
		if !ok {
			log.Error("value is not a []string")
			utils.WriteErrorf(w, http.StatusInternalServerError, "Auth-Scope is not a []string")
			return
		}

		if len(scopesSlice) == 0 {
			utils.WriteErrorf(w, http.StatusInternalServerError, "Auth-Scope is empty")
			return
		}

		claims, err := auth.ValidateJwtInRequest(r, globals.JwtKey)
		if err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "auth error: %v", err)
			return
		}

		if claims != nil && err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "Token expired: %v", claims.Subject)
			return
		}

		dbUser, err := database.FindUserByEmail(claims.Subject)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "User find error: %v", err)
			return
		}
		dbUser.Claims = *claims

		log.Info(fmt.Sprintf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method))

		dbUser.Role = "User"
		for _, email := range globals.ADMINS {
			if claims.Subject == email {
				log.Info(fmt.Sprintf("Authenticated admin %s\n", email))
				dbUser.Role = "Admin"

			}
		}
		ctx = context.WithValue(ctx, CurrentUser, dbUser)

		switch scopesSlice[0] {
		case "Admin":
			if dbUser.Role == "Admin" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			utils.WriteErrorf(w, http.StatusUnauthorized, "You are not admin: %v", claims.Subject)
			return
		case "User":
			next.ServeHTTP(w, r.WithContext(ctx))
		default:
			utils.WriteErrorf(w, http.StatusInternalServerError, "Unknown scope")
			return
		}
	}
}
