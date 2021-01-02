package main

import (
	_ "api/docs"
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/base32"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go/v72"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/crypto/ed25519"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	db        *sql.DB
	opts      *Opts
	jwtKey    []byte
	debug     = true
	privRSA   *rsa.PrivateKey
	privEdDSA *ed25519.PrivateKey
)

type Opts struct {
	Host         string
	Port         int
	HS256        string
	Env          string
	StripeSecret string
}

func NewOpts() *Opts {
	o := &Opts{}

	flag.StringVar(&o.Env, "env", LookupEnv("ENV"), "ENV variable")
	flag.StringVar(&o.Host, "host", LookupEnv("HOST"), "Host")
	flag.IntVar(&o.Port, "port", LookupEnvInt("PORT"), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", LookupEnv("HS256"), "HS256 key")
	flag.StringVar(&o.StripeSecret, "string", LookupEnv("STRIPE_SECRET"), "Stripe secret")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	//set defaults
	if o.Env == "local" {
		err := godotenv.Load()
		if err != nil {
			err = godotenv.Load("../.env")
			if err != nil {
				log.Printf("could not find env file in this or in the parent dir: %v", err)
			}
		}
	}
	o.Host = setDefault(o.Host, LookupEnv("HOST"), "db")
	o.Port = setDefaultInt(o.Port, LookupEnvInt("PORT"), 9082)
	o.HS256 = setDefault(o.HS256, LookupEnv("HS256"), "ORSXG5A=")

	if o.HS256 != "" {
		var err error
		jwtKey, err = base32.StdEncoding.DecodeString(o.HS256)
		if err != nil {
			log.Fatalf("cannot decode %v", o.HS256)
		}
	}

	return o
}

func setDefault(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func setDefaultInt(values ...int) int {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

func LookupEnv(key string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return ""
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

// @title Flatfeestack API
// @version 0.0.1
// @host localhost:8080
// @BasePath /
func main() {
	opts = NewOpts()

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

	fmt.Println("Starting server on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}

func writeErr(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Printf(msg)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(code)
	if debug {
		w.Write([]byte(`{"error":"` + msg + `"}`))
	}
}

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/hooks") {
				// webhook routes are not protected by middleware, but verified on an individual level
				next.ServeHTTP(w, r)
				return
			}
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErr(w, http.StatusBadRequest, "ERR-01, authorization header not set")
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 {
				writeErr(w, http.StatusBadRequest, "ERR-02, could not split token: %v", bearerToken)
				return
			}

			tok, err := jwt.ParseSigned(bearerToken[1])
			if err != nil {
				writeErr(w, http.StatusBadRequest, "ERR-03, could not parse token: %v", bearerToken[1])
				return
			}

			claims := &jwt.Claims{}

			if tok.Headers[0].Algorithm == string(jose.RS256) {
				err = tok.Claims(privRSA.Public(), claims)
			} else if tok.Headers[0].Algorithm == string(jose.HS256) {
				err = tok.Claims(jwtKey, claims)
			} else if tok.Headers[0].Algorithm == string(jose.EdDSA) {
				err = tok.Claims(privEdDSA.Public(), claims)
			} else {
				writeErr(w, http.StatusUnauthorized, "ERR-04, unknown algorithm: %v", tok.Headers[0].Algorithm)
				return
			}

			if err != nil {
				writeErr(w, http.StatusUnauthorized, "ERR-05, could not parse claims: %v", bearerToken[1])
				return
			}

			if claims.Expiry != nil && !claims.Expiry.Time().After(time.Now()) {
				writeErr(w, http.StatusBadRequest, "ERR-06, expired: %v", claims.Expiry.Time())
				return
			}

			if claims.Subject == "" {
				writeErr(w, http.StatusBadRequest, "ERR-07, no subject: %v", claims)
				return
			}

			// Fetch user from DB
			user, err := FindUserByEmail(claims.Subject)
			if err != nil {
				writeErr(w, http.StatusBadRequest, "ERR-08, user find error: %v", err)
				return
			}

			if user == nil {
				user, err = createUser(claims.Subject)
				if err != nil {
					writeErr(w, http.StatusBadRequest, "ERR-09, user update error: %v", err)
					return
				}
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "user", user)
			log.Printf("Authenticated user %s\n", user.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func createUser(email string) (*User, error) {
	log.Printf("need to create a user")
	var user User
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user.Id = uid
	user.Email = &email
	sid := "local"
	if opts.Env != "local" {
		sid, err = createStripeCustomer(&user)
	}
	user.StripeId = &sid

	err = SaveUser(&user)
	if err != nil {
		return nil, err
	}
	log.Printf("user created")
	return &user, nil
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
