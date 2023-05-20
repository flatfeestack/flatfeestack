package jwt

import (
	"context"
	"encoding/base64"
	"fmt"
	"forum/api"
	"forum/globals"
	"forum/types"
	"forum/utils"
	"github.com/flatfeestack/go-lib/auth"
	"github.com/go-jose/go-jose/v3/json"
	log "github.com/sirupsen/logrus"
	"net/http"
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

		claims, err := auth.ValidateJwtInRequest(r, globals.JwtKey)
		if err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "auth error: %v", err)
			return
		}

		if claims != nil && err != nil {
			utils.WriteErrorf(w, http.StatusUnauthorized, "Token expired: %v", claims.Subject)
			return
		}

		user, err := findUserByEmail(claims.Subject)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "User find error: %v", err)
			return
		}
		user.Claims = *claims

		log.Info(fmt.Sprintf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method))

		user.Role = "User"
		for _, email := range globals.ADMINS {
			if claims.Subject == email {
				log.Info(fmt.Sprintf("Authenticated admin %s\n", email))
				user.Role = "Admin"

			}
		}
		ctx = context.WithValue(ctx, CurrentUser, user)

		switch scopesSlice[0] {
		case "Admin":
			if user.Role == "Admin" {
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

func findUserByEmail(email string) (*types.User, error) {
	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, globals.OPTS.BackendUrl+"users/by/"+email, nil)
	if err != nil {
		log.Printf("Could not create a HTTP request to call the backend %v", err)
		return nil, err
	}

	basicAuth := globals.OPTS.BackendUsername + ":" + globals.OPTS.BackendPassword
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basicAuth)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error sending HTTP request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	var user types.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return nil, err
	}

	return &user, nil
}
