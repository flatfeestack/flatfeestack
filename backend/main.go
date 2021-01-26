package main

import (
	_ "backend/docs"
	"crypto/rsa"
	"database/sql"
	"encoding/base32"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/go-co-op/gocron"
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
	admins    []string
)

type Opts struct {
	Port         int
	HS256        string
	Env          string
	StripeSecret string
	DBPath       string
	DBDriver     string
	AnalysisUrl  string
	PayoutUrl    string
	Admins       string
}

func NewOpts() *Opts {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &Opts{}
	flag.StringVar(&o.Env, "env", lookupEnv("ENV",
		"local"), "ENV variable")
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT",
		9082), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", lookupEnv("HS256",
		"ORSXG5A="), "HS256 key")
	flag.StringVar(&o.StripeSecret, "string", lookupEnv("STRIPE_SECRET"), "Stripe secret")
	flag.StringVar(&o.DBPath, "db-path", lookupEnv("DB_PATH",
		"postgresql://postgres:password@db:5432/flatfeestack?sslmode=disable"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", lookupEnv("DB_DRIVER",
		"postgres"), "DB driver")
	flag.StringVar(&o.AnalysisUrl, "analysis-url", lookupEnv("ANALYSIS-URL",
		"http://analysis-engine:9083"), "Analysis Url")
	flag.StringVar(&o.PayoutUrl, "payout-url", lookupEnv("PAYOUT-URL",
		"http://payout:9084"), "Payout Url")
	flag.StringVar(&o.Admins, "admins", lookupEnv("ADMINS"), "Admins")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults
	if o.Env == "local" || o.Env == "dev" {
		debug = true
	}

	jwtKey, err = base32.StdEncoding.DecodeString(o.HS256)
	if err != nil {
		log.Fatalf("cannot decode %v", o.HS256)
	}

	if o.Admins != "" {
		admins = strings.Split(o.Admins, ";")
	}

	return o
}

func lookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			return v
		}
	}
	return ""
}

func lookupEnvInt(key string, defaultValues ...int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("LookupEnvInt[%s]: %v", key, err)
			return 0
		}
		return v
	}
	for _, v := range defaultValues {
		if v != 0 {
			return v
		}
	}
	return 0
}

// @title Flatfeestack API
// @version 0.0.1
// @host localhost:8080
// @BasePath /
func main() {
	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	opts = NewOpts()
	db = initDb()

	stripe.Key = opts.StripeSecret

	// Routes
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/backend").Subrouter()
	//user
	apiRouter.HandleFunc("/users/me", jwtAuthUser(getMyUser)).Methods("GET")
	apiRouter.HandleFunc("/users/me/connectedEmails", jwtAuthUser(getMyConnectedEmails)).Methods("GET")
	apiRouter.HandleFunc("/users/me/connectedEmails", jwtAuthUser(addGitEmail)).Methods("POST")
	apiRouter.HandleFunc("/users/me/connectedEmails/{email}", jwtAuthUser(removeGitEmail)).Methods("DELETE")
	apiRouter.HandleFunc("/users/me/payout/{address}", jwtAuthUser(updatePayout)).Methods("PUT")
	apiRouter.HandleFunc("/users/me/sponsored", jwtAuthUser(getSponsoredRepos)).Methods("GET")
	//repo github
	apiRouter.HandleFunc("/repos/search", jwtAuthUser(searchRepoGitHub)).Methods("GET")
	apiRouter.HandleFunc("/repos/sponsor/github/{id}", jwtAuthUser(sponsorRepoGitHub)).Methods("POST")
	//repo
	apiRouter.HandleFunc("/repos/{id}", jwtAuthUser(getRepoByID)).Methods("GET")
	apiRouter.HandleFunc("/repos/{id}/sponsor", jwtAuthUser(sponsorRepo)).Methods("POST")
	apiRouter.HandleFunc("/repos/{id}/unsponsor", jwtAuthUser(unsponsorRepo)).Methods("POST")
	apiRouter.HandleFunc("/exchanges", jwtAuthUser(getExchanges)).Methods("GET")
	//payment
	apiRouter.HandleFunc("/payments/subscriptions", jwtAuthUser(postSubscription)).Methods("POST")
	apiRouter.HandleFunc("/hooks/stripe", stripeWebhook).Methods("POST")
	apiRouter.HandleFunc("/hooks/analysis-engine", jwtAuthAdmin(analysisEngineHook, "analysis-engine@flatfeestack.io")).Methods("POST")
	apiRouter.HandleFunc("/admin/pending-payout", jwtAuthAdmin(pendingPayouts, opts.Admins)).Methods("GET")
	apiRouter.HandleFunc("/admin/payout", jwtAuthAdmin(payout, opts.Admins)).Methods("POST")
	apiRouter.HandleFunc("/admin/time", jwtAuthAdmin(serverTime, opts.Admins)).Methods("GET")

	//dev settings
	if opts.Env == "local" || opts.Env == "dev" {
		router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
		apiRouter.HandleFunc("/admin/fake-user", jwtAuthAdmin(fakeUser, opts.Admins)).Methods("POST")
		apiRouter.HandleFunc("/admin/timewarp/{hours}", jwtAuthAdmin(timeWarp, opts.Admins)).Methods("POST")
	}

	//scheduler

	s1 := gocron.NewScheduler(time.Local)

	j, err := s1.Every(1).Day().At("00:01").Do(dailyRunner)
	if err != nil {
		log.Printf("error during job execution: %v, runcount: %v", err, j.RunCount())
	}

	s1.Every(1).Week().At("02:01").Do(weeklyRunner)

	s1.Every(1).Month(1).At("04:01").Do(monthlyRunner)

	log.Println("Starting backend on port " + strconv.Itoa(opts.Port))
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

func jwtAuth(w http.ResponseWriter, r *http.Request) *jwt.Claims {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeErr(w, http.StatusBadRequest, "ERR-01, authorization header not set")
		return nil
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		writeErr(w, http.StatusBadRequest, "ERR-02, could not split token: %v", bearerToken)
		return nil
	}

	tok, err := jwt.ParseSigned(bearerToken[1])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "ERR-03, could not parse token: %v", bearerToken[1])
		return nil
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
		return nil
	}

	if err != nil {
		writeErr(w, http.StatusUnauthorized, "ERR-05, could not parse claims: %v", bearerToken[1])
		return nil
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(time.Now()) {
		writeErr(w, http.StatusTeapot, "ERR-06, expired: %v", claims.Expiry.Time())
		return nil
	}

	if claims.Subject == "" {
		writeErr(w, http.StatusBadRequest, "ERR-07, no subject: %v", claims)
		return nil
	}
	return claims
}

func jwtAuthAdmin(next func(w http.ResponseWriter, r *http.Request, email string), emails ...string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := jwtAuth(w, r)
		if claims == nil {
			return
		}
		for _, email := range emails {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				next(w, r, email)
				return
			}
		}
		writeErr(w, http.StatusBadRequest, "ERR-01,jwtAuthAdmin error: %v != %v", claims.Subject, emails)
	}
}

func jwtAuthUser(next func(w http.ResponseWriter, r *http.Request, user *User)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := jwtAuth(w, r)
		if claims == nil {
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
	log.Printf("need to stringPointer a user")
	var user User
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user.Id = uid
	user.Email = &email
	if opts.Env != "local" {
		sid, err := createStripeCustomer(&user)
		if err != nil {
			return nil, err
		}
		user.StripeId = &sid
	} else {
		user.StripeId = &opts.Env
	}

	err = saveUser(&user)
	if err != nil {
		return nil, err
	}
	log.Printf("user created")
	return &user, nil
}
