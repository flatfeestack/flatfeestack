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
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/ed25519"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Active = iota + 1
	Inactive
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
)

type Opts struct {
	Port                      int
	HS256                     string
	Env                       string
	StripeAPISecretKey        string
	StripeAPIPublicKey        string
	StripeWebhookSecretKey    string
	DBPath                    string
	DBDriver                  string
	AnalysisUrl               string
	PayoutUrl                 string
	Admins                    string
	EmailLinkPrefix           string
	EmailFrom                 string
	EmailFromName             string
	EmailUrl                  string
	EmailToken                string
	WebSocketBaseUrl          string
	ContractAddr              string
	NowpaymentsToken          string
	NowpaymentsIpnKey         string
	NowpaymentsApiUrl         string
	NowpaymentsIpnCallbackUrl string
}

type TokenClaims struct {
	Scope            string                 `json:"scope,omitempty"`
	InviteMetaSystem map[string]interface{} `json:"inviteMetaSystem,omitempty"`
	InviteMetaUser   map[string]interface{} `json:"inviteMetaUser,omitempty"`
	jwt.Claims
}

func NewOpts() *Opts {
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
	flag.StringVar(&o.AnalysisUrl, "analysis-url", lookupEnv("ANALYSIS_URL",
		"http://analysis-engine:9083"), "Analysis Url")
	flag.StringVar(&o.PayoutUrl, "payout-url", lookupEnv("PAYOUT_URL",
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
	flag.StringVar(&o.ContractAddr, "contract-addr", lookupEnv("CONTRACT_ADDR",
		"0x731a10897d267e19b34503ad902d0a29173ba4b1"), "Default Ethereum Address")
	flag.StringVar(&o.NowpaymentsToken, "nowpayments-token", lookupEnv("NOWPAYMENTS_TOKEN"), "Token for NOWPayments access")
	flag.StringVar(&o.NowpaymentsIpnKey, "nowpayments-ipn-key", lookupEnv("NOWPAYMENTS_IPN_KEY"), "Key for NOWPayments IPN")
	flag.StringVar(&o.NowpaymentsApiUrl, "nowpayments-api-url", lookupEnv("NOWPAYMENTS_API_URL",
		"https://api.sandbox.nowpayments.io/v1"), "NOWPayments API URL")
	flag.StringVar(&o.NowpaymentsIpnCallbackUrl, "nowpayments-ipn-callback-url", lookupEnv("NOWPAYMENTS_IPN_CALLBACK_URL"), "Callback URL for NOWPayments IPN")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults
	if o.Env == "local" || o.Env == "dev" {
		debug = true
	}

	var err error
	jwtKey, err = base32.StdEncoding.DecodeString(o.HS256)
	if err != nil {
		log.Fatalf("cannot decode %v", o.HS256)
	}

	admins = strings.Split(o.Admins, ";")

	if o.EmailFrom == "" {
		o.EmailFrom = "info@flatfeestack.io"
	}

	if o.StripeWebhookSecretKey == "" {
		o.StripeWebhookSecretKey = "whsec_BlO0hcHIJb82nUM9v8fpq0WP55FxKF2U"
	}

	return o
}

func lookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			os.Setenv(key, v)
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
			os.Setenv(key, strconv.Itoa(v))
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
	//the .env should be loaded before showing the banner, as the banner shows also the ENVs
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}
	//this will set the default ENVs
	opts = NewOpts()

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

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
	router.HandleFunc("/users/me/method/{method}", jwtAuthUser(updateMethod)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/method", jwtAuthUser(deleteMethod)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/sponsored", jwtAuthUser(getSponsoredRepos)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/name/{name}", jwtAuthUser(updateName)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/image", maxBytes(jwtAuthUser(updateImage), 200*1024)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(setupStripe)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(cancelSub)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwtAuthUser(stripePaymentInitial)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwtAuthUser(nowPayment)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/payment", jwtAuthUser(ws)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/payment-cycle", jwtAuthUser(paymentCycle)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwtAuthUser(statusSponsoredUsers)).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-snd", jwtAuthUser(contributionsSend)).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwtAuthUser(contributionsRcv)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwtAuthUser(contributionsSum)).Methods(http.MethodPost)
	router.HandleFunc("/users/contributions-summary/{uuid}", contributionsSum2).Methods(http.MethodPost)
	router.HandleFunc("/users/summary/{uuid}", userSummary2).Methods(http.MethodPost)
	router.HandleFunc("/users/me/wallets", jwtAuthUser(getUserWallets)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/wallets", jwtAuthUser(addUserWallet)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/wallets/{uuid}", jwtAuthUser(deleteUserWallet)).Methods(http.MethodDelete)

	//
	router.HandleFunc("/users/git-email", confirmConnectedEmails).Methods(http.MethodPost)
	//repo github
	router.HandleFunc("/repos/search", jwtAuthUser(searchRepoGitHub)).Methods(http.MethodGet)
	router.HandleFunc("/repos/name", jwtAuthUser(searchRepoNames)).Methods(http.MethodGet)
	router.HandleFunc("/repos/link/{repoId}", jwtAuthUser(linkGitUrl)).Methods(http.MethodPost)
	router.HandleFunc("/repos/root/{repoId}/{rootUuid}", jwtAuthUser(makeRoot)).Methods(http.MethodGet)
	router.HandleFunc("/repos/{id}", jwtAuthUser(getRepoByID)).Methods(http.MethodGet)
	router.HandleFunc("/repos/{id}/tag", jwtAuthUser(tagRepo)).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/untag", jwtAuthUser(unTagRepo)).Methods(http.MethodPost)
	//payment

	//hooks
	router.HandleFunc("/hooks/stripe", maxBytes(stripeWebhook, 65536)).Methods(http.MethodPost)
	router.HandleFunc("/hooks/nowpayments", nowWebhook).Methods(http.MethodPost)
	router.HandleFunc("/hooks/analysis-engine", jwtAuthAdmin(analysisEngineHook, []string{"analysis-engine@flatfeestack.io"})).Methods(http.MethodPost)

	//admin
	router.HandleFunc("/admin/payout", jwtAuthAdmin(getPayoutInfos, admins)).Methods(http.MethodGet)
	router.HandleFunc("/admin/payout/{exchangeRate}", jwtAuthAdmin(monthlyPayout, admins)).Methods(http.MethodPost)
	router.HandleFunc("/admin/time", jwtAuthAdmin(serverTime, admins)).Methods(http.MethodGet)
	router.HandleFunc("/admin/users", jwtAuthAdmin(users, admins)).Methods(http.MethodPost)

	router.HandleFunc("/config", config).Methods(http.MethodGet)

	//dev settings
	if opts.Env == "local" || opts.Env == "dev" {
		router.HandleFunc("/admin/fake/user/{email}", jwtAuthAdmin(fakeUser, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/payment/{email}/{seats}", jwtAuthAdmin(fakePayment, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/contribution", jwtAuthAdmin(fakeContribution, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/timewarp/{hours}", jwtAuthAdmin(timeWarp, admins)).Methods(http.MethodPost)
		router.HandleFunc("/nowpayments/crontester", jwtAuthAdmin(crontester, admins)).Methods(http.MethodPost)
	}

	router.HandleFunc("/confirm/invite/{email}", jwtAuthUser(confirmInvite)).Methods(http.MethodPost)
	router.HandleFunc("/invite", jwtAuthUser(invitations)).Methods(http.MethodGet)
	router.HandleFunc("/invite/by/{email}", jwtAuthUser(inviteByDelete)).Methods(http.MethodDelete)
	router.HandleFunc("/invite/my/{email}", jwtAuthUser(inviteMyDelete)).Methods(http.MethodDelete)
	router.HandleFunc("/invite/{email}", jwtAuthUser(inviteOther)).Methods(http.MethodPost)
	/**
	//invites
	router.HandleFunc("/confirm/invite-new", confirmInviteNew).Methods(http.MethodPost)

	//invites
	router.HandleFunc("/invite", jwtAuth(inviteOther)).Methods(http.MethodPost)


	//TODO: not yet in the frontend
	router.HandleFunc("/invite", jwtAuth(inviteResetMyToken)).Methods(http.MethodPatch)
	*/

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[404] no route matched for: %s, %s", r.URL, r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	//scheduler
	cronJobDay(dailyRunner, timeNow())
	cronJobHour(hourlyRunner, timeNow())

	log.Println("Starting backend on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
	cronStop()
}

func writeErrorf(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Error(msg)
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
			writeErrorf(w, http.StatusUnauthorized, "Token expired: %v, available: %v", claims.Subject, emails)
			return
		} else if claims == nil && err != nil {
			writeErrorf(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
			return
		}
		for _, email := range emails {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				next(w, r, email)
				return
			}
		}
		writeErrorf(w, http.StatusBadRequest, "ERR-01,jwtAuthAdmin error: %v != %v", claims.Subject, emails)
	}
}

func jwtAuthUser(next func(w http.ResponseWriter, r *http.Request, user *User)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwtAuth(r)

		if claims != nil && err != nil {
			if r.Header.Get("Sec-WebSocket-Protocol") == "" {
				//no websocket
				writeErrorf(w, http.StatusUnauthorized, "Token expired: %v", claims.Subject)
			} else {
				//we use websocket
				wsNoAuth(w, r)
			}
			return
		} else if claims == nil && err != nil {
			writeErrorf(w, http.StatusBadRequest, "jwtAuthAdmin error: %v", err)
			return
		}

		unlock := km.Lock(claims.Subject)
		defer unlock()

		// Fetch user from DB
		user, err := findUserByEmail(claims.Subject)
		if err != nil {
			writeErrorf(w, http.StatusBadRequest, "ERR-08, user find error: %v", err)
			return
		}

		if user == nil {
			user, err = createUser(claims.Subject)
			if err != nil {
				writeErrorf(w, http.StatusBadRequest, "ERR-09, user update error: %v", err)
				return
			}
		}

		//User exists now, check if we are admin
		for _, email := range admins {
			if claims.Subject == email {
				log.Printf("Authenticated admin %s\n", email)
				user.Role = stringPointer("admin")
			}
		}

		user.Claims = claims
		log.Printf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method)
		next(w, r, user)
	}
}

func createUser(email string) (*User, error) {
	payOutId := uuid.New()

	user := User{
		Id:                uuid.New(),
		PaymentCycleOutId: &payOutId,
		Email:             email,
		CreatedAt:         timeNow(),
	}

	err := insertUser(&user)
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

func validateEmail(email string) error {
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(email) > 254 || !rxEmail.MatchString(email) {
		return fmt.Errorf("[%s] is not a valid email address", email)
	}
	return nil
}
