package main

import (
	_ "api/docs"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go/v72"
	httpSwagger "github.com/swaggo/http-swagger"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	db   *sql.DB
	opts *Opts
)

type Opts struct {
	Host string
	Port int
}

func NewOpts() *Opts {
	o := &Opts{}
	flag.StringVar(&o.Host, "host", LookupEnv("HOST"), "Host")
	flag.IntVar(&o.Port, "port", LookupEnvInt("PORT"), "listening HTTP port")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	return o
}

func setDefault(actualValue string, defaultValue string) string {
	if actualValue == "" {
		return defaultValue
	}
	return actualValue
}

func LookupEnv(key string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return ""
}

func setDefaultInt(actualValue int, defaultValue int) int {
	if actualValue == 0 {
		return defaultValue
	}
	return actualValue
}

func LookupEnvInt(key string) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("LookupEnvInt[%s]: %v", key, err)
			return 0
		}
		return v
	}
	return 0
}

func defaultOpts(o *Opts) {
	o.Host = setDefault(o.Host, "db")
	opts.Port = setDefaultInt(opts.Port, 9082)
}

// @title Flatfeestack API
// @version 0.0.1
// @host localhost:8080
// @BasePath /
func main() {
	opts = NewOpts()
	defaultOpts(opts)
	if os.Getenv("ENV") == "local" {
		//if run locally get environment file from above docker config file
		err := godotenv.Load("../.env")
		if err != nil {
			log.Fatalf("could not find env file. Please add an .env file if you want to run it without docker.", err)
		}
	}

	stripe.Key = os.Getenv("STRIPE_SECRET")

	db = createConnection()

	// Routes
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(AuthMiddleware())
	apiRouter.HandleFunc("/users/me", GetMyUser).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/me/connectedEmails", GetMyConnectedEmails).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/me/connectedEmails", AddGitEmail).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/users/me/payout", PutPayoutAddress).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/users/me/payout", GetPayoutAddress).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/me/connectedEmails/{email}", RemoveGitEmail).Methods("DELETE", "OPTIONS")
	apiRouter.HandleFunc("/users/sponsored", GetSponsoredRepos).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{id}", GetUserByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users", CreateUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos", CreateRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/search", SearchRepo).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}", GetRepoByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/sponsor", SponsorRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/unsponsor", UnsponsorRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/payments/subscriptions", PostSubscription).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/exchanges", GetExchanges).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/exchanges/{id}", PutExchange).Methods("PUT", "OPTIONS")

	//webhooks
	apiRouter.HandleFunc("/hooks/stripe", StripeWebhook).Methods("POST", "OPTIONS")

	// Swagger
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	fmt.Println("Starting server on port "+strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}

type authMiddlewareKey string

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/hooks") {
				// webhook routes are not protected by middleware, but verified on an individual level
				next.ServeHTTP(w, r)
				return
			}
			reqToken := r.Header.Get("Authorization")
			if !strings.HasPrefix(reqToken, "Bearer") {
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

	dbString = fmt.Sprintf("postgresql://%v:%v@%v:5432/%v?sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), opts.Host, os.Getenv("POSTGRES_DB"))
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
