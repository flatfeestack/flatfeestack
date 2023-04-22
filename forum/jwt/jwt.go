package jwt

import (
	"context"
	"fmt"
	"forum/api"
	database "forum/db"
	"forum/globals"
	"forum/types"
	"forum/utils"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
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

		claims, err := jwtAuth(r)
		if err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "auth error: %v", err)
			return
		}

		if claims != nil && err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "Token expired: %v", claims.Subject)
			return
		}

		unlock := globals.KM.Lock(claims.Subject)
		defer unlock()

		dbUser, err := database.FindUserByEmail(claims.Subject)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "User find error: %v", err)
			return
		}
		dbUser.Claims = *claims

		log.Info(fmt.Sprintf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method))

		switch scopesSlice[0] {
		case "Admin":
			log.Debug("Admin scope")
			for _, email := range globals.ADMINS {
				if claims.Subject == email {
					log.Info(fmt.Sprintf("Authenticated admin %s\n", email))
					dbUser.Role = "Admin"
					ctx = context.WithValue(ctx, CurrentUser, dbUser)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			utils.WriteErrorf(w, http.StatusUnauthorized, "You are not admin: %v", claims.Subject)
			return
		case "User":
			log.Debug("User scope")
			ctx = context.WithValue(ctx, CurrentUser, dbUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		default:
			utils.WriteErrorf(w, http.StatusInternalServerError, "Unknown scope")
			return
		}
	}
}

func jwtAuth(r *http.Request) (*types.TokenClaims, error) {
	authHeader := r.Header.Get("Authorization")
	var bearerToken = ""
	if authHeader == "" {
		return nil, fmt.Errorf("ERR-01, authorization header not set for %v", r.URL)
	}
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		return nil, fmt.Errorf("ERR-02, could not split token: %v", bearerToken)
	}
	bearerToken = split[1]

	tok, err := jwt.ParseSigned(bearerToken)
	if err != nil {
		return nil, fmt.Errorf("ERR-03, could not parse token: %v", bearerToken[1])
	}

	claims := &types.TokenClaims{}

	if tok.Headers[0].Algorithm == string(jose.HS256) {
		err = tok.Claims(globals.JwtKey, claims)
	} else {
		return nil, fmt.Errorf("ERR-04, unknown algorithm: %v", tok.Headers[0].Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("ERR-05, could not parse claims: err=%v for token=%v", err, bearerToken)
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(time.Now().UTC()) {
		return claims, fmt.Errorf("ERR-06, unauthorized: %v", bearerToken)
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("ERR-07, no subject: %v", claims)
	}
	return claims, nil
}
