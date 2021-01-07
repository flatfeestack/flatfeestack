package main

import (
	_ "api/docs"
	"bytes"
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
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
	SPONSOR = iota
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

	flag.StringVar(&o.Env, "env", LookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", LookupEnvInt("PORT"), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", LookupEnv("HS256"), "HS256 key")
	flag.StringVar(&o.StripeSecret, "string", LookupEnv("STRIPE_SECRET"), "Stripe secret")
	flag.StringVar(&o.DBPath, "db-path", LookupEnv("DB_PATH"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", LookupEnv("DB_DRIVER"), "DB driver")
	flag.StringVar(&o.AnalysisUrl, "analysis-url", LookupEnv("ANALYSIS-URL"), "Analysis Url")

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

	o.Port = setDefaultInt(o.Port, LookupEnvInt("PORT"), 9082)
	o.HS256 = setDefault(o.HS256, LookupEnv("HS256"), "ORSXG5A=")
	o.DBDriver = setDefault(o.DBDriver, LookupEnv("DB_DRIVER"), "postgres")
	o.DBPath = setDefault(o.DBDriver, LookupEnv("DB_PATH"), "postgresql://postgres:password@db:5432/flatfeestack?sslmode=disable")
	o.AnalysisUrl = setDefault(o.AnalysisUrl, LookupEnv("ANALYSIS-URL"), "http://analysis-engine:8080/webhook")

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
	db = initDb()

	stripe.Key = opts.StripeSecret

	// Routes
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(AuthMiddleware())
	apiRouter.HandleFunc("/users/me", GetMyUser).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/me/connectedEmails", GetMyConnectedEmails).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/me/connectedEmails", AddGitEmail).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/users/me/connectedEmails/{email}", RemoveGitEmail).Methods("DELETE", "OPTIONS")
	apiRouter.HandleFunc("/users/me/payout/{address}", UpdatePayout).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/users/sponsored", GetSponsoredRepos).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{id}", GetUserByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users", CreateUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos", CreateRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/search", searchRepo).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}", GetRepoByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/sponsor", SponsorRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}/unsponsor", UnsponsorRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/payments/subscriptions", PostSubscription).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/exchanges", GetExchanges).Methods("GET", "OPTIONS")

	//webhooks
	apiRouter.HandleFunc("/hooks/stripe", StripeWebhook).Methods("POST", "OPTIONS")

	// Swagger
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

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

			ctx := r.Context()
			ctx = context.WithValue(ctx, "user", user)
			log.Printf("Authenticated user %s\n", *user.Email)
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

	err = saveUser(&user)
	if err != nil {
		return nil, err
	}
	log.Printf("user created")
	return &user, nil
}

type ExchangeRate struct {
	Ethereum struct {
		Usd decimal.Decimal `json:"usd"`
	} `json:"ethereum"`
}

//https://www.coingecko.com/en/api
func getPriceETH() (decimal.Decimal, error) {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	//curl -X GET "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd" -H  "accept: application/json"
	r, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")
	if err != nil {
		return decimal.Zero, err
	}
	defer r.Body.Close()
	rate := ExchangeRate{}
	err = json.NewDecoder(r.Body).Decode(&rate)
	if err != nil {
		return decimal.Zero, err
	}
	return rate.Ethereum.Usd, nil
}

type AnalysisRequest struct {
	RepositoryUrl       string    `json:"repository_url"`
	DateFrom            time.Time `json:"since"`
	DateTo              time.Time `json:"until"`
	PlatformInformation bool      `json:"platform_information"`
	Branch              string    `json:"branch"`
}

type AnalysisResponse struct {
	RequestId uuid.UUID `json:"request_id"`
}

func analysisRequest(repoId uuid.UUID, repoUrl string) error {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	now := time.Now()
	req := AnalysisRequest{
		RepositoryUrl:       repoUrl,
		DateFrom:            now.AddDate(0, -3, 0),
		DateTo:              now,
		PlatformInformation: false,
		Branch:              "master",
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r, err := client.Post(opts.AnalysisUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	var resp AnalysisResponse
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return err
	}

	return saveAnalysisRequest(resp.RequestId, repoId, req.DateFrom, req.DateTo, req.Branch)
}
