package main

import (
	_ "api/docs"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go/v72"
	httpSwagger "github.com/swaggo/http-swagger"
	"gopkg.in/square/go-jose.v2/jwt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/joho/godotenv"
)

var (
	db *sql.DB
	host string
)

// @title Flatfeestack API
// @version 0.0.1
// @host localhost:8080
// @BasePath /
func main() {
	host = "db"
	if os.Getenv("ENV") == "local" {
		//if run locally get environment file from above docker config file
		err := godotenv.Load("../.env")
		if err!=nil{
			log.Fatalf("could not find env file. Please add an .env file if you want to run it without docker.",err)
		}
		host = "localhost"
	}

	stripe.Key = os.Getenv("STRIPE_SECRET")

	db = createConnection()
	initDbErr := initDB()
	if initDbErr != nil {
		fmt.Printf("Could not init DB %v", initDbErr)
	}

	// Routes
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(AuthMiddleware())
	apiRouter.HandleFunc("/users/me", GetMyUser).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/sponsored", GetSponsoredRepos).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{id}", GetUserByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{id}/sponsored/calculateDaily", CalculateDailyRepoBalanceByUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/users", CreateUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos", CreateRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/search", SearchRepo).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}", GetRepoByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/sponsor", SponsorRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/unsponsor", UnsponsorRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/payments/subscriptions", PostSubscription).Methods("POST", "OPTIONS")

	//webhooks
	apiRouter.HandleFunc("/hooks/stripe", StripeWebhook).Methods("POST","OPTIONS")

	// Swagger
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	fmt.Println("Starting server on the port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

type authMiddlewareKey string

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/hooks"){
				// webhook routes are not protected by middleware, but verified on an individual level
				next.ServeHTTP(w, r)
				return
			}
			reqToken := r.Header.Get("Authorization")
			if !strings.HasPrefix(reqToken, "Bearer"){
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
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
				return
			}

			sub := fmt.Sprintf("%v", out["sub"])

			// Fetch user from DB
			user, userErr := FindUserByEmail(sub)
			if userErr != nil {
				log.Printf("Could not get user %s, %v\n", userErr, user)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// User needs to be created
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
				userErr := SaveUser(&newUser)
				if userErr != nil {
					log.Printf("Could not create user %v", userErr)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				log.Printf("user created")

				// Create Stripe user
				userWithStripe, stripeErr := CreateStripeCustomer(newUser)
				if stripeErr != nil {
					log.Printf("Could not create user in stripe %v", stripeErr)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				user = userWithStripe
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

	dbString = fmt.Sprintf("postgresql://%v:%v@%v:5432/%v?sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), host, os.Getenv("POSTGRES_DB"))
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

func initDB() error {
	//this will create or alter tables
	//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat("init.sql"); err == nil {
		file, err := ioutil.ReadFile("init.sql")
		if err != nil {
			return err
		}
		requests := strings.Split(string(file), ";\n\n")
		for _, request := range requests {
			request = strings.Replace(request, "\n", "", -1)
			request = strings.Replace(request, "\t", "", -1)
			if !strings.HasPrefix(request, "#") {
				_, err = db.Exec(request)
				if err != nil {
					return fmt.Errorf("[%v] %v", request, err)
				}
			}
		}
	}
	return nil
}