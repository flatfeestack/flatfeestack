package main

import (
	"backend/api"
	"backend/clients"
	db2 "backend/db"
	"backend/utils"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v74"
	"golang.org/x/crypto/ed25519"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	opts      *Opts
	jwtKey    []byte
	privRSA   *rsa.PrivateKey
	privEdDSA *ed25519.PrivateKey
	debug     bool
	admins    []string
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
	DBScripts                 string
	AnalysisUrl               string
	PayoutUrl                 string
	Admins                    string
	EmailLinkPrefix           string
	EmailFrom                 string
	EmailFromName             string
	EmailUrl                  string
	EmailToken                string
	EmailMarketing            string
	WebSocketBaseUrl          string
	NowpaymentsToken          string
	NowpaymentsIpnKey         string
	NowpaymentsApiUrl         string
	NowpaymentsIpnCallbackUrl string
	ServerKey                 string
}

func NewOpts() *Opts {
	o := &Opts{}
	flag.StringVar(&o.Env, "env", lookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT",
		9082), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", lookupEnv("HS256"), "HS256 key")
	flag.StringVar(&o.StripeAPISecretKey, "stripe-secret-api", lookupEnv("STRIPE_SECRET_API"), "Stripe API secret")
	flag.StringVar(&o.StripeAPIPublicKey, "stripe-public-api", lookupEnv("STRIPE_PUBLIC_API"), "Public Key for stripe")
	flag.StringVar(&o.StripeWebhookSecretKey, "stripe-secret-webhook", lookupEnv("STRIPE_SECRET_WEBHOOK"), "Stripe webhook secret")
	flag.StringVar(&o.DBPath, "db-path", lookupEnv("DB_PATH",
		"postgresql://postgres:password@db:5432/flatfeestack?sslmode=disable"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", lookupEnv("DB_DRIVER",
		"postgres"), "DB driver")
	flag.StringVar(&o.DBScripts, "db-scripts", lookupEnv("DB_SCRIPTS"), "DB scripts to run at startup")
	flag.StringVar(&o.AnalysisUrl, "analyzer-url", lookupEnv("ANALYZER_URL",
		"http://analyzer:9083"), "Analysis Url")
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
	flag.StringVar(&o.EmailMarketing, "email-marketing", lookupEnv("EMAIL_MARKETING",
		"tom.marketing@bocek.ch"), "Email marketing email. Set the value to 'live' to send out real emails")
	flag.StringVar(&o.WebSocketBaseUrl, "ws-base-url", lookupEnv("WS_BASE_URL",
		"ws://localhost"), "Websocket base URL")
	flag.StringVar(&o.NowpaymentsToken, "nowpayments-token", lookupEnv("NOWPAYMENTS_TOKEN"), "Token for NOWPayments access")
	flag.StringVar(&o.NowpaymentsIpnKey, "nowpayments-ipn-key", lookupEnv("NOWPAYMENTS_IPN_KEY"), "Key for NOWPayments IPN")
	flag.StringVar(&o.NowpaymentsApiUrl, "nowpayments-api-url", lookupEnv("NOWPAYMENTS_API_URL",
		"https://api.sandbox.nowpayments.io/v1"), "NOWPayments API URL")
	flag.StringVar(&o.NowpaymentsIpnCallbackUrl, "nowpayments-ipn-callback-url", lookupEnv("NOWPAYMENTS_IPN_CALLBACK_URL"), "Callback URL for NOWPayments IPN")
	flag.StringVar(&o.ServerKey, "server-key", lookupEnv("SERVER_KEY"), "make secure calls to the subsystems")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults, be explicit
	if o.Env == "local" || o.Env == "dev" {
		debug = true
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if o.HS256 != "" {
		var err error
		jwtKey, err = base32.StdEncoding.DecodeString(o.HS256)
		if err != nil {
			h := sha256.New()
			h.Write([]byte(o.HS256))
			jwtKey = h.Sum(nil)
			log.Debugf("jwtKey: %v", jwtKey)
		}
	} else {
		log.Fatalf("HS256 seed is required, non was provided")
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
			err := os.Setenv(key, v)
			if err != nil {
				log.Printf("LookupEnvInt[%s]: %v", key, err)
				return ""
			}
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
			err := os.Setenv(key, strconv.Itoa(v))
			if err != nil {
				log.Printf("LookupEnvInt[%s]: %v", key, err)
				return 0
			}
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

	clients.Init(opts.PayoutUrl, opts.ServerKey, opts.AnalysisUrl)
	clients.InitEmail(opts.EmailUrl, opts.EmailFromName, opts.EmailFrom, opts.EmailToken, opts.Env, opts.EmailMarketing, opts.EmailLinkPrefix)
	api.InitStripe(opts.StripeAPISecretKey, opts.StripeWebhookSecretKey)
	api.InitNow(opts.NowpaymentsApiUrl, opts.NowpaymentsToken, opts.NowpaymentsIpnCallbackUrl, opts.NowpaymentsIpnKey)
	api.InitApi(opts.StripeAPIPublicKey, opts.WebSocketBaseUrl, opts.Env)

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	err = db2.InitDb(opts.DBDriver, opts.DBPath, opts.DBScripts)
	if err != nil {
		log.Fatal(err)
	}

	stripe.Key = opts.StripeAPISecretKey

	// Routes
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return logRequestHandler(next)
	})
	//apiRouter := router.PathPrefix("/backend").Subrouter()
	//user
	router.HandleFunc("/users/me/git-email", jwtAuthUser(api.GetMyConnectedEmails)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwtAuthUser(api.AddGitEmail)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/{email}", jwtAuthUser(api.RemoveGitEmail)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me", jwtAuthUser(api.GetMyUser)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/method/{method}", jwtAuthUser(api.UpdateMethod)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/method", jwtAuthUser(api.DeleteMethod)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/name/{name}", jwtAuthUser(api.UpdateName)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/image", maxBytes(jwtAuthUser(api.UpdateImage), 200*1024)).Methods(http.MethodPost)
	//unauthenticated
	router.HandleFunc("/users/summary/{uuid}", api.UserSummary2).Methods(http.MethodPost)
	router.HandleFunc("/users/git-email", api.ConfirmConnectedEmails).Methods(http.MethodPost)
	//repo
	router.HandleFunc("/users/me/sponsored", jwtAuthUser(api.GetSponsoredRepos)).Methods(http.MethodGet)
	router.HandleFunc("/repos/name", jwtAuthUser(api.SearchRepoNames)).Methods(http.MethodGet)

	router.HandleFunc("/users/me/stripe", jwtAuthUser(api.SetupStripe)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(api.CancelSub)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwtAuthUser(api.StripePaymentInitial)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwtAuthUser(api.NowPayment)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/payment", jwtAuthUser(api.WebSocket)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/payment-cycle", jwtAuthUser(api.PaymentCycle)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwtAuthUser(api.StatusSponsoredUsers)).Methods(http.MethodPost)
	//contributions
	router.HandleFunc("/users/contrib-snd", jwtAuthUser(api.ContributionsSend)).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwtAuthUser(api.ContributionsRcv)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwtAuthUser(api.ContributionsSum)).Methods(http.MethodPost)
	router.HandleFunc("/users/contributions-summary/{uuid}", api.ContributionsSum2).Methods(http.MethodPost)
	//wallet
	router.HandleFunc("/users/me/wallets", jwtAuthUser(api.GetUserWallets)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/wallets", jwtAuthUser(api.AddUserWallet)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/wallets/{uuid}", jwtAuthUser(api.DeleteUserWallet)).Methods(http.MethodDelete)
	//github
	router.HandleFunc("/repos/search", jwtAuthUser(api.SearchRepoGitHub)).Methods(http.MethodGet)
	//repo
	router.HandleFunc("/repos/{id}", jwtAuthUser(api.GetRepoByID)).Methods(http.MethodGet)
	router.HandleFunc("/repos/{id}/tag", jwtAuthUser(api.TagRepo)).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/untag", jwtAuthUser(api.UnTagRepo)).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/{offset}/graph", jwtAuthUser(api.Graph)).Methods(http.MethodGet)
	//payment

	//hooks
	router.HandleFunc("/hooks/stripe", maxBytes(api.StripeWebhook, 65536)).Methods(http.MethodPost)
	router.HandleFunc("/hooks/nowpayments", api.NowWebhook).Methods(http.MethodPost)
	router.HandleFunc("/hooks/analyzer", jwtAuthAdmin(api.AnalysisEngineHook, []string{"ffs-server"})).Methods(http.MethodPost)

	//admin
	router.HandleFunc("/admin/payout/{exchangeRate}", jwtAuthAdmin(api.MonthlyPayout, admins)).Methods(http.MethodPost)
	router.HandleFunc("/admin/time", jwtAuthAdmin(api.ServerTime, admins)).Methods(http.MethodGet)
	router.HandleFunc("/admin/users", jwtAuthAdmin(api.Users, admins)).Methods(http.MethodPost)

	router.HandleFunc("/config", api.Config).Methods(http.MethodGet)

	//dev settings
	if debug {
		router.HandleFunc("/admin/fake/user/{email}", jwtAuthAdmin(api.FakeUser, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/payment/{email}/{seats}", jwtAuthAdmin(api.FakePayment, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/contribution", jwtAuthAdmin(api.FakeContribution, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/timewarp/{hours}", jwtAuthAdmin(api.TimeWarp, admins)).Methods(http.MethodPost)
	}

	//invite
	router.HandleFunc("/confirm/invite/{email}", jwtAuthUser(api.ConfirmInvite)).Methods(http.MethodPost)
	router.HandleFunc("/invite", jwtAuthUser(api.Invitations)).Methods(http.MethodGet)
	router.HandleFunc("/invite/by/{email}", jwtAuthUser(api.InviteByDelete)).Methods(http.MethodDelete)
	router.HandleFunc("/invite/my/{email}", jwtAuthUser(api.InviteMyDelete)).Methods(http.MethodDelete)
	router.HandleFunc("/invite/{email}", jwtAuthUser(api.InviteOther)).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[404] no route matched for: %s, %s", r.URL, r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	//scheduler
	cronJobDay(dailyRunner, utils.TimeNow())
	cronJobHour(hourlyRunner, utils.TimeNow())

	log.Println("Starting backend on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
	cronStop()
}

func maxBytes(next func(w http.ResponseWriter, r *http.Request), size int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, size)
		next(w, r)
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
