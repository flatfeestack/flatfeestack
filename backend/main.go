package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/base32"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kjk/dailyrotate"
	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/ed25519"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Active = iota + 1
	Inactive
)

const (
	mUSDPerHour = 13750 //1.375 cents - x 10'000
	mUSDPerDay  = mUSDPerHour * 24
)

var (
	db        *sql.DB
	opts      *Opts
	jwtKey    []byte
	privRSA   *rsa.PrivateKey
	privEdDSA *ed25519.PrivateKey
	debug     bool
	admins    []string
	hoursAdd  int
	km        = KeyedMutex{}
	logFile   *dailyrotate.File
)

type Opts struct {
	Port                   int
	HS256                  string
	Env                    string
	StripeAPISecretKey     string
	StripeAPIPublicKey     string
	StripeWebhookSecretKey string
	DBPath                 string
	DBDriver               string
	AnalysisUrl            string
	PayoutUrl              string
	Admins                 string
	EmailLinkPrefix        string
	EmailFrom              string
	EmailFromName          string
	EmailUrl               string
	EmailToken             string
	WebSocketBaseUrl       string
	RestTimeout            int
	LogPath                string
	ContractAddr           string
}

type TokenClaims struct {
	Meta         *string  `json:"meta,omitempty"`
	Scope        string   `json:"scope,omitempty"`
	InviteToken  string   `json:"inviteToken,omitempty"`
	InviteEmails []string `json:"inviteEmails,omitempty"`
	InviteMeta   []string `json:"inviteMeta,omitempty"`
	jwt.Claims
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
	flag.StringVar(&o.StripeAPISecretKey, "stripe-secret-api", lookupEnv("STRIPE_SECRET_API"), "Stripe API secret")
	flag.StringVar(&o.StripeAPIPublicKey, "stripe-public-api", lookupEnv("STRIPE_PUBLIC_API"), "Public Key for stripe")
	flag.StringVar(&o.StripeWebhookSecretKey, "stripe-secret-webhook", lookupEnv("STRIPE_SECRET_WEBHOOK"), "Stripe webhook secret")
	flag.StringVar(&o.DBPath, "db-path", lookupEnv("DB_PATH",
		"postgresql://postgres:password@db:5432/flatfeestack?sslmode=disable"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", lookupEnv("DB_DRIVER",
		"postgres"), "DB driver")
	flag.StringVar(&o.AnalysisUrl, "analysis-url", lookupEnv("ANALYSIS-URL",
		"http://analysis-engine:9083"), "Analysis Url")
	flag.StringVar(&o.PayoutUrl, "payout-url", lookupEnv("PAYOUT-URL",
		"http://payout:9084"), "Payout Url")
	flag.StringVar(&o.Admins, "admins", lookupEnv("ADMINS"), "Admins")
	flag.StringVar(&o.EmailFrom, "email-from", lookupEnv("EMAIL_FROM"), "Email from, default is info@flatfeestack.io")
	flag.StringVar(&o.EmailFromName, "email-from-name", lookupEnv("EMAIL_FROM_NAME"), "Email from name, default is a empty string")
	flag.StringVar(&o.EmailUrl, "email-url", lookupEnv("EMAIL_URL",
		"http://localhost"), "Email service URL")
	flag.StringVar(&o.EmailToken, "email-token", lookupEnv("EMAIL_TOKEN"), "Email service token")
	flag.StringVar(&o.EmailLinkPrefix, "email-prefix", lookupEnv("EMAIL_PREFIX",
		"http://localhost/"), "Email link prefix")
	flag.StringVar(&o.WebSocketBaseUrl, "ws-base-url", lookupEnv("WS_BASE_URL",
		"ws://localhost"), "Websocket base URL")
	flag.IntVar(&o.RestTimeout, "rest-timeout", lookupEnvInt("REST_TIMEOUT",
		5000), "Rest timeout, default 5s")
	flag.StringVar(&o.LogPath, "log", lookupEnv("LOG",
		os.TempDir()+"/ffs/"), "Log directory, default is /tmp/ffs/")
	flag.StringVar(&o.ContractAddr, "contract-addr", lookupEnv("CONTRACT_ADDR",
		"0x731a10897d267e19b34503ad902d0a29173ba4b1"), "Default Ethereum Address")

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

	admins = strings.Split(o.Admins, ";")

	if o.EmailFrom == "" {
		o.EmailFrom = "info@flatfeestack.io"
	}

	if o.StripeWebhookSecretKey == "" {
		o.StripeWebhookSecretKey = "whsec_9HJx5EoyhE1K3UFBnTxpOSr0lscZMHJL"
	}

	pathFormat := filepath.Join(o.LogPath, "backend_2006-01-02.txt")
	w, err := dailyrotate.NewFile(pathFormat, func(string, bool) {})
	if err != nil {
		log.Fatalf("cannot log")
	}
	logFile = w

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

// @title To run locally, set these ENV vars=LD_PRELOAD=/usr/local/lib/faketime/libfaketime.so.1;FAKETIME_NO_CACHE=1
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

	stripe.Key = opts.StripeAPISecretKey

	// Routes
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return logRequestHandler(next)
	})
	//apiRouter := router.PathPrefix("/backend").Subrouter()
	//user
	router.HandleFunc("/users/me", jwtAuthUser(getMyUser)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwtAuthUser(getMyConnectedEmails)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwtAuthUser(addGitEmail)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/{email}", jwtAuthUser(removeGitEmail)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/payout/{address}", jwtAuthUser(updatePayout)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/method/{method}", jwtAuthUser(updateMethod)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/method", jwtAuthUser(deleteMethod)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/sponsored", jwtAuthUser(getSponsoredRepos)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/name/{name}", jwtAuthUser(updateName)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/image", maxBytes(jwtAuthUser(updateImage), 200*1024)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/mode/{mode}", jwtAuthUser(updateMode)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(setupStripe)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(cancelSub)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwtAuthUser(stripePaymentInitial)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/payment", jwtAuthUser(ws)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/topup", jwtAuthUser(topup)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/payment-cycle", jwtAuthUser(paymentCycle)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/seats/{seats}", jwtAuthUser(updateSeats)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwtAuthUser(statusSponsoredUsers)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-send", jwtAuthUser(contributionsSend)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-receive", jwtAuthUser(contributionsRcv)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwtAuthUser(contributionsSum)).Methods(http.MethodPost)
	router.HandleFunc("/users/contributions-summary/{uuid}", contributionsSum2).Methods(http.MethodPost)
	router.HandleFunc("/users/summary/{uuid}", userSummary2).Methods(http.MethodPost)
	router.HandleFunc("/users/me/payout-pending", jwtAuthUser(pendingDailyUserPayouts)).Methods(http.MethodPost)
	//
	router.HandleFunc("/users/git-email", confirmConnectedEmails).Methods(http.MethodPost)
	//repo github
	router.HandleFunc("/repos/search", jwtAuthUser(searchRepoGitHub)).Methods(http.MethodGet)
	router.HandleFunc("/repos/{id}", jwtAuthUser(getRepoByID)).Methods(http.MethodGet)
	router.HandleFunc("/repos/tag", jwtAuthUser(tagRepo)).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/untag", jwtAuthUser(unTagRepo)).Methods(http.MethodPost)
	//payment

	router.HandleFunc("/hooks/stripe", maxBytes(stripeWebhook, 65536)).Methods(http.MethodPost)
	router.HandleFunc("/hooks/analysis-engine", jwtAuthAdmin(analysisEngineHook, []string{"analysis-engine@flatfeestack.io"})).Methods(http.MethodPost)
	router.HandleFunc("/admin/pending-payout/{type}", jwtAuthAdmin(getPayouts, admins)).Methods(http.MethodPost)
	router.HandleFunc("/admin/payout/{exchangeRate}", jwtAuthAdmin(payout, admins)).Methods(http.MethodPost)
	router.HandleFunc("/admin/time", jwtAuthAdmin(serverTime, admins)).Methods(http.MethodGet)
	router.HandleFunc("/admin/users", jwtAuthAdmin(users, admins)).Methods(http.MethodPost)

	router.HandleFunc("/config", config).Methods(http.MethodGet)

	router.HandleFunc("/hooks/nowpayments", nowpaymentsWebhook).Methods(http.MethodPost)
	router.HandleFunc("/users/me/nowpayments", jwtAuthUser(nowpaymentPayment)).Methods(http.MethodPost)

	//dev settings
	if opts.Env == "local" || opts.Env == "dev" {
		router.HandleFunc("/admin/fake/user/{email}", jwtAuthAdmin(fakeUser, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/payment/{email}/{seats}", jwtAuthAdmin(fakePayment, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/contribution", jwtAuthAdmin(fakeContribution, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/timewarp/{hours}", jwtAuthAdmin(timeWarp, admins)).Methods(http.MethodPost)
	}

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[404] no route matched for: %s, %s", r.URL, r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	//scheduler
	cronJobDay(dailyRunner, timeNow())

	log.Println("Starting backend on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
	cronStop()
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

func maxBytes(next func(w http.ResponseWriter, r *http.Request), size int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, size)
		next(w, r)
	}
}

func jwtAuth(r *http.Request) (*TokenClaims, error) {
	authHeader := r.Header.Get("Authorization")
	var bearerToken = ""
	if authHeader == "" {
		authHeader = r.Header.Get("Sec-WebSocket-Protocol")
		if authHeader == "" {
			return nil, fmt.Errorf("ERR-01, authorization header not set for %v", r.URL)
		}
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

	claims := &TokenClaims{}

	if tok.Headers[0].Algorithm == string(jose.RS256) {
		err = tok.Claims(privRSA.Public(), claims)
	} else if tok.Headers[0].Algorithm == string(jose.HS256) {
		err = tok.Claims(jwtKey, claims)
	} else if tok.Headers[0].Algorithm == string(jose.EdDSA) {
		err = tok.Claims(privEdDSA.Public(), claims)
	} else {
		return nil, fmt.Errorf("ERR-04, unknown algorithm: %v", tok.Headers[0].Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("ERR-05, could not parse claims: %v", bearerToken)
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(timeNow()) {
		return claims, fmt.Errorf("ERR-06, unauthorized: %v", bearerToken)
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("ERR-07, no subject: %v", claims)
	}
	return claims, nil
}

func jwtAuthAdmin(next func(w http.ResponseWriter, r *http.Request, email string), emails []string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwtAuth(r)
		if claims != nil && err != nil {
			writeErr(w, http.StatusUnauthorized, "Token expired: %v, available: %v", claims.Subject, emails)
			return
		} else if claims == nil && err != nil {
			writeErr(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
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
		claims, err := jwtAuth(r)

		if claims != nil && err != nil {
			if r.Header.Get("Sec-WebSocket-Protocol") == "" {
				//no websocket
				writeErr(w, http.StatusUnauthorized, "Token expired: %v", claims.Subject)
			} else {
				//we use websocket
				wsNoAuth(w, r)
			}
			return
		} else if claims == nil && err != nil {
			writeErr(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
			return
		}

		unlock := km.Lock(claims.Subject)
		defer unlock()

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

		found := true
		if len(claims.InviteEmails) > 0 {
			found = false
		} else {
			if user.InvitedEmail != nil {
				err = updateInvitedEmail(nil, user.Id)
				if err != nil {
					writeErr(w, http.StatusBadRequest, "ERR-09, user find error: %v", err)
					return
				}
			}
		}
		for _, v := range claims.InviteEmails {
			if user.InvitedEmail != nil && v == *user.InvitedEmail {
				found = true
				break
			}
		}
		if !found {
			err = updateInvitedEmail(&claims.InviteEmails[0], user.Id)
			if err != nil {
				writeErr(w, http.StatusBadRequest, "ERR-09, user find error: %v", err)
				return
			}
		}

		user.Claims = claims
		log.Printf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method)
		next(w, r, user)
	}
}

func createUser(email string) (*User, error) {
	var user User
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user.Id = uid
	user.Email = email
	user.CreatedAt = timeNow()
	user.Role = stringPointer("USR")

	rnd, err := genRnd(18)
	if err != nil {
		return nil, err
	}
	token := base32.StdEncoding.EncodeToString(rnd)

	err = insertUser(&user, token)
	if err != nil {
		return nil, err
	}
	log.Printf("user %v created", user)
	return &user, nil
}

func stringPointer(s string) *string {
	return &s
}

func timeNow() time.Time {
	if opts != nil && (opts.Env == "local" || opts.Env == "dev") {
		return time.Now().Add(time.Duration(hoursAdd) * time.Hour).UTC()
	} else {
		return time.Now().UTC()
	}
}

type KeyedMutex struct {
	mutexes sync.Map // Zero value is empty and ready for use
}

func (m *KeyedMutex) Lock(key string) func() {
	value, _ := m.mutexes.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}
