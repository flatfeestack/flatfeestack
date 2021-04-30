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
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/ed25519"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"net/http"
	"os"
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
}

type TokenClaims struct {
	Meta        *string   `json:"meta,omitempty"`
	Scope       string    `json:"scope,omitempty"`
	InviteToken string    `json:"invite_token,omitempty"`
	InviteEmail *string   `json:"invite_email,omitempty"`
	InvitedAt   time.Time `json:"invited_at,omitempty"`
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
	flag.StringVar(&o.StripeAPIPublicKey, "stripe-public-api", lookupEnv("STRIPE_PUBLIC_API",
		"pk_test_51ITqIGItjdVuh2paNpnIUSWtsHJCLwY9fBYtiH2leQh2BvaMWB4de40Ea0ntC14nnmYcUyBD21LKO9ldlaXL6DJJ00Qm1toLdb"), "Public Key for localhost")
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
		"ws://localhost/"), "Websocket base URL")

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

	if o.EmailFrom == "" {
		o.EmailFrom = "info@flatfeestack.io"
	}

	if o.EmailFromName == "" {
		o.EmailFromName = ""
	}

	if o.StripeWebhookSecretKey == "" {
		o.StripeWebhookSecretKey = "whsec_9HJx5EoyhE1K3UFBnTxpOSr0lscZMHJL"
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
	apiRouter := mux.NewRouter()
	//apiRouter := router.PathPrefix("/backend").Subrouter()
	//user
	apiRouter.HandleFunc("/users/me", jwtAuthUser(getMyUser)).Methods("GET")
	apiRouter.HandleFunc("/users/me/git-email", jwtAuthUser(getMyConnectedEmails)).Methods("GET")
	apiRouter.HandleFunc("/users/me/git-email", jwtAuthUser(addGitEmail)).Methods("POST")
	apiRouter.HandleFunc("/users/me/git-email/{email}", jwtAuthUser(removeGitEmail)).Methods("DELETE")
	apiRouter.HandleFunc("/users/me/payout/{address}", jwtAuthUser(updatePayout)).Methods("PUT")
	apiRouter.HandleFunc("/users/me/method/{method}", jwtAuthUser(updateMethod)).Methods("PUT")
	apiRouter.HandleFunc("/users/me/method", jwtAuthUser(deleteMethod)).Methods(http.MethodDelete)
	apiRouter.HandleFunc("/users/me/sponsored", jwtAuthUser(getSponsoredRepos)).Methods("GET")
	apiRouter.HandleFunc("/users/me/name/{name}", jwtAuthUser(updateName)).Methods("PUT")
	apiRouter.HandleFunc("/users/me/image", maxBytesMiddleware(jwtAuthUser(updateImage), 200*1024)).Methods("POST")
	apiRouter.HandleFunc("/users/me/mode/{mode}", jwtAuthUser(updateMode)).Methods("PUT")
	apiRouter.HandleFunc("/users/me/stripe", jwtAuthUser(setupStripe)).Methods(http.MethodPost)
	apiRouter.HandleFunc("/users/me/stripe", jwtAuthUser(cancelSub)).Methods(http.MethodDelete)
	apiRouter.HandleFunc("/users/me/stripe/{freq}/{seats}", jwtAuthUser(stripePaymentInitial)).Methods("PUT")
	apiRouter.HandleFunc("/users/me/payment", jwtAuthUser(ws)).Methods(http.MethodGet)
	//
	apiRouter.HandleFunc("/users/git-email", confirmConnectedEmails).Methods("POST")
	//repo github
	apiRouter.HandleFunc("/repos/search", jwtAuthUser(searchRepoGitHub)).Methods("GET")
	apiRouter.HandleFunc("/repos/{id}", jwtAuthUser(getRepoByID)).Methods("GET")
	apiRouter.HandleFunc("/repos/tag", jwtAuthUser(tagRepo)).Methods("POST")
	apiRouter.HandleFunc("/repos/{id}/untag", jwtAuthUser(unTagRepo)).Methods("POST")
	//payment

	apiRouter.HandleFunc("/hooks/stripe", maxBytesMiddleware(stripeWebhook, 65536)).Methods("POST")
	apiRouter.HandleFunc("/hooks/analysis-engine", jwtAuthAdmin(analysisEngineHook, []string{"analysis-engine@flatfeestack.io"})).Methods("POST")
	apiRouter.HandleFunc("/admin/pending-payout/{type}", jwtAuthAdmin(getPayouts, admins)).Methods("POST")
	apiRouter.HandleFunc("/admin/payout/{exchangeRate}", jwtAuthAdmin(payout, admins)).Methods("POST")
	apiRouter.HandleFunc("/admin/time", jwtAuthAdmin(serverTime, admins)).Methods("GET")

	apiRouter.HandleFunc("/config", jwtAuthUser(config)).Methods(http.MethodGet)

	//dev settings
	if opts.Env == "local" || opts.Env == "dev" {
		apiRouter.HandleFunc("/admin/fake-user", jwtAuthAdmin(fakeUser, admins)).Methods("POST")
		apiRouter.HandleFunc("/admin/timewarp/{hours}", jwtAuthAdmin(timeWarp, admins)).Methods("POST")
	}

	apiRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("ERROR: Unknown path [%s]:%s", r.URL, r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	//scheduler
	cronJobDay(dailyRunner, timeNow())

	log.Println("Starting backend on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), apiRouter))
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

func maxBytesMiddleware(next func(w http.ResponseWriter, r *http.Request), size int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, size)
		next(w, r)
	}
}

func jwtAuth(w http.ResponseWriter, r *http.Request) *TokenClaims {
	authHeader := r.Header.Get("Authorization")
	var bearerToken = ""
	if authHeader == "" {
		authHeader = r.Header.Get("Sec-WebSocket-Protocol")
		if authHeader == "" {
			writeErr(w, http.StatusBadRequest, "ERR-01, authorization header not set for %v", r.URL)
			return nil
		}
		w.Header().Set("Sec-WebSocket-Protocol", "access_token")
	}
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		writeErr(w, http.StatusBadRequest, "ERR-02, could not split token: %v", bearerToken)
		return nil
	}
	bearerToken = split[1]

	tok, err := jwt.ParseSigned(bearerToken)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "ERR-03, could not parse token: %v", bearerToken[1])
		return nil
	}

	claims := &TokenClaims{}

	if tok.Headers[0].Algorithm == string(jose.RS256) {
		err = tok.Claims(privRSA.Public(), claims)
	} else if tok.Headers[0].Algorithm == string(jose.HS256) {
		err = tok.Claims(jwtKey, claims)
	} else if tok.Headers[0].Algorithm == string(jose.EdDSA) {
		err = tok.Claims(privEdDSA.Public(), claims)
	} else {
		writeErr(w, http.StatusBadRequest, "ERR-04, unknown algorithm: %v", tok.Headers[0].Algorithm)
		return nil
	}

	if err != nil {
		writeErr(w, http.StatusBadRequest, "ERR-05, could not parse claims: %v", bearerToken)
		return nil
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(timeNow()) {
		writeErr(w, http.StatusUnauthorized, "ERR-06, expired: %v", claims.Expiry.Time())
		return nil
	}

	if claims.Subject == "" {
		writeErr(w, http.StatusBadRequest, "ERR-07, no subject: %v", claims)
		return nil
	}
	return claims
}

func jwtAuthAdmin(next func(w http.ResponseWriter, r *http.Request, email string), emails []string) func(http.ResponseWriter, *http.Request) {
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
			writeErr(w, http.StatusBadRequest, "ERR-08a, no claims provided: %v", r.URL)
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
			if claims.InviteEmail != nil {
				//we have a new sponsor!
				err := takeSponsorship(*claims.InviteEmail, user.Id)
				if err != nil {
					writeErr(w, http.StatusBadRequest, "ERR-08, taking sponsorship failed: %v", err)
					return
				}
			}
		}

		log.Printf("User [%s] request [%s]:%s\n", claims.Subject, r.URL, r.Method)
		next(w, r, user)
	}
}

func takeSponsorship(inviteEmail string, userId uuid.UUID) error {
	//check if we have an inviteEmail that changed
	sponsor, err := findUserByEmail(inviteEmail)
	if err != nil {
		return err
	}

	pc, err := findPaymentCycle(sponsor.PaymentCycleId)
	if err != nil {
		return err
	}

	updateSponsor(userId, inviteEmail, sponsor.Id)
	currentSeats, err := findSeats(sponsor.Id)
	if err != nil {
		return err
	}
	if int64(pc.Seats) > currentSeats {
		//parent has enough seats go for it!
		sum, err := findSumUserBalance(sponsor.Id)
		if err != nil {
			return err
		}
		if sum > mUSDPerDay {
			transferBalance(sponsor.PaymentCycleId, userId, sponsor.Id, pc.Freq*mUSDPerDay, "SPON", timeNow())
		} else {
			log.Printf("parent has not enough funding")
		}
	} else {
		log.Printf("parent has not enough seats")
	}
	return nil
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
