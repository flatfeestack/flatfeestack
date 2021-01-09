package main

import (
	_ "api/docs"
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

const (
	SPONSOR = iota + 1
	UNSPONSOR
)

var (
	db        *sql.DB
	opts      *Opts
	jwtKey    []byte
	privRSA   *rsa.PrivateKey
	privEdDSA *ed25519.PrivateKey
	debug     bool
)

type Opts struct {
	Port         int
	HS256        string
	Env          string
	StripeSecret string
	DBPath       string
	DBDriver     string
	AnalysisUrl  string
}

func NewOpts() *Opts {
	o := &Opts{}

	flag.StringVar(&o.Env, "env", lookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT"), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", lookupEnv("HS256"), "HS256 key")
	flag.StringVar(&o.StripeSecret, "string", lookupEnv("STRIPE_SECRET"), "Stripe secret")
	flag.StringVar(&o.DBPath, "db-path", lookupEnv("DB_PATH"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", lookupEnv("DB_DRIVER"), "DB driver")
	flag.StringVar(&o.AnalysisUrl, "analysis-url", lookupEnv("ANALYSIS-URL"), "Analysis Url")

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
	if o.Env == "local" || o.Env == "dev" {
		debug = true
	}

	o.Port = setDefaultInt(o.Port, lookupEnvInt("PORT"), 9082)
	o.HS256 = setDefault(o.HS256, lookupEnv("HS256"), "ORSXG5A=")
	o.DBDriver = setDefault(o.DBDriver, lookupEnv("DB_DRIVER"), "postgres")
	o.DBPath = setDefault(o.DBPath, lookupEnv("DB_PATH"), "postgresql://postgres:password@db:5432/flatfeestack?sslmode=disable")
	o.AnalysisUrl = setDefault(o.AnalysisUrl, lookupEnv("ANALYSIS-URL"), "http://analysis-engine:8080/webhook")

	var err error
	jwtKey, err = base32.StdEncoding.DecodeString(o.HS256)
	if err != nil {
		log.Fatalf("cannot decode %v", o.HS256)
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

func lookupEnv(key string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return ""
}

func lookupEnvInt(key string) int {
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
	db = initDb()

	stripe.Key = opts.StripeSecret

	// Routes
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	//user
	apiRouter.HandleFunc("/users/me", jwtAuth(getMyUser)).Methods("GET")
	apiRouter.HandleFunc("/users/me/connectedEmails", jwtAuth(getMyConnectedEmails)).Methods("GET")
	apiRouter.HandleFunc("/users/me/connectedEmails", jwtAuth(addGitEmail)).Methods("POST")
	apiRouter.HandleFunc("/users/me/connectedEmails/{email}", jwtAuth(removeGitEmail)).Methods("DELETE")
	apiRouter.HandleFunc("/users/me/payout/{address}", jwtAuth(updatePayout)).Methods("PUT")
	apiRouter.HandleFunc("/users/me/sponsored", jwtAuth(getSponsoredRepos)).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", getUserByID).Methods("GET")
	//repo github
	apiRouter.HandleFunc("/repos/search/github/{query}", jwtAuth(searchRepoGitHub)).Methods("GET")
	apiRouter.HandleFunc("/repos/sponsor/github/{id}", jwtAuth(sponsorRepoGitHub)).Methods("POST")
	//repo
	apiRouter.HandleFunc("/repos/{id}", jwtAuth(getRepoByID)).Methods("GET")
	apiRouter.HandleFunc("/repos/{id}/sponsor", jwtAuth(sponsorRepo)).Methods("POST")
	apiRouter.HandleFunc("/repos/{id}/unsponsor", jwtAuth(unsponsorRepo)).Methods("POST")
	apiRouter.HandleFunc("/exchanges", jwtAuth(getExchanges)).Methods("GET")
	//payment
	apiRouter.HandleFunc("/payments/subscriptions", jwtAuth(postSubscription)).Methods("POST")
	apiRouter.HandleFunc("/hooks/stripe", stripeWebhook).Methods("POST")

	if opts.Env == "local" || opts.Env == "dev" {
		router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	}

	log.Println("Starting server on port " + strconv.Itoa(opts.Port))
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

func jwtAuth(next func(w http.ResponseWriter, r *http.Request, user *User)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/hooks") {
				// webhook routes are not protected by middleware, but verified on an individual level
				next(w, r, nil)
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
				writeErr(w, http.StatusTeapot, "ERR-06, expired: %v", claims.Expiry.Time())
				return
			}

			if claims.Subject == "" {
				writeErr(w, http.StatusBadRequest, "ERR-07, no subject: %v", claims)
				return
			}

			// Fetch user from DB
			user, err := findUserByEmail(claims.Subject)
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

			log.Printf("Authenticated user %s\n", *user.Email)
			next(w, r, user)
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

	err = saveUser(&user)
	if err != nil {
		return nil, err
	}
	log.Printf("user created")
	return &user, nil
}
