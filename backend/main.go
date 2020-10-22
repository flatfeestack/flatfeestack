package main

import (
	_ "api/docs"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"net/http"
	"os"
	"strings"
)

// @title Flatfeestack API
// @version 0.0.1
// @host localhost:8080
// @BasePath /
func main() {
	db := createConnection()

	userRepo := NewUserRepo(db)
	repoRepo := NewRepoRepo(db)
	sponsorEventRepo := NewSponsorEventRepo(db)
	dailyRepoBalanceRepo := NewDailyRepoBalanceRepo(db)

	h := NewBaseHandler(userRepo, repoRepo, sponsorEventRepo, dailyRepoBalanceRepo)

	// Routes
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(AuthMiddleware(h))
	apiRouter.HandleFunc("/users/me", h.GetMyUser).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{id}", h.GetUserByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{id}/sponsored", h.GetSponsoredRepos).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{id}/sponsored/calculateDaily", h.CalculateDailyRepoBalanceByUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/users", h.CreateUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos", h.CreateRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}", h.GetRepoByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/sponsor", h.SponsorRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/unsponsor", h.UnsponsorRepo).Methods("POST", "OPTIONS")
	//apiRouter.Use(AuthMiddleware)

	// Swagger
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	fmt.Println("Starting server on the port 8080...")

	log.Fatal(http.ListenAndServe(":8080", router))
}
type authMiddlewareKey string

func AuthMiddleware(h *BaseHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqToken := r.Header.Get("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
			reqToken = splitToken[1]
			token, err := jwt.ParseSigned(reqToken)
			out := make(map[string]interface{})

			// TODO: check signature of token
			if err := token.UnsafeClaimsWithoutVerification(&out); err != nil {
				panic(err)
			}
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
			}

			sub := fmt.Sprintf("%v",out["sub"])

			// Fetch user from DB
			user, userErr := h.userRepo.FindByEmail(sub)
			if userErr != nil {
				log.Printf("Could not get user %s, %v\n", userErr, user)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if user.ID == "" {
				log.Printf("need to create a user")
				var newUser User
				uid, uuidErr := uuid.NewRandom()
				if uuidErr != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				newUser.ID = uid.String()
				newUser.Email = sub
				userErr := h.userRepo.Save(&newUser)
				if userErr != nil {
					log.Printf("Could not create user %v",userErr)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				log.Printf("user created")
				user = &newUser
			}

			if user.ID == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
			} else {
				ctx := r.Context()
				ctx = context.WithValue(ctx, authMiddlewareKey("user"), user)
				log.Printf("Authenticated user %s\n", user.Email)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

// create connection with postgres db
func createConnection() *sql.DB {
	// Open the connection
	var dbString string
	if os.Getenv("ENV") != "local" {
		dbString = fmt.Sprintf("postgresql://%v:%v@postgres:5432/%v?sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
	} else {
		dbString = "postgresql://postgres:password@localhost:5432/flatfeestack?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbString)

	if err != nil {
		panic(err)
	}

	// check the connection
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	// return the connection
	return db
}
