package main

import (
	"backend/api"
	"backend/clients"
	"backend/utils"
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"fmt"
	"github.com/flatfeestack/go-lib/auth"
	"github.com/stripe/stripe-go/v74"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dimiro1/banner"
	dbLib "github.com/flatfeestack/go-lib/database"
	env "github.com/flatfeestack/go-lib/environment"
	prom "github.com/flatfeestack/go-lib/prometheus"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	Active = iota + 1
	Inactive
)

var (
	opts   *Opts
	jwtKey []byte
	debug  bool
	admins []string
)

type Subsystem struct {
	Url      string
	Username string
	Password string
}

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
	Admins                    string
	EmailLinkPrefix           string
	EmailFrom                 string
	EmailFromName             string
	EmailUrl                  string
	EmailToken                string
	EmailMarketing            string
	NowpaymentsToken          string
	NowpaymentsIpnKey         string
	NowpaymentsApiUrl         string
	NowpaymentsIpnCallbackUrl string
	BackendUsername           string
	BackendPassword           string
	Payout                    Subsystem
	Analyzer                  Subsystem
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func NewOpts() *Opts {
	o := &Opts{}

	flag.StringVar(&o.Env, "env", env.LookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", env.LookupEnvInt("PORT",
		9082), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", env.LookupEnv("HS256"), "HS256 key")
	flag.StringVar(&o.StripeAPISecretKey, "stripe-secret-api", env.LookupEnv("STRIPE_SECRET_API"), "Stripe API secret")
	flag.StringVar(&o.StripeAPIPublicKey, "stripe-public-api", env.LookupEnv("STRIPE_PUBLIC_API"), "Public Key for stripe")
	flag.StringVar(&o.StripeWebhookSecretKey, "stripe-secret-webhook", env.LookupEnv("STRIPE_SECRET_WEBHOOK"), "Stripe webhook secret")
	flag.StringVar(&o.DBPath, "db-path", env.LookupEnv("DB_PATH",
		"postgresql://postgres:password@db:5432/flatfeestack?sslmode=disable"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", env.LookupEnv("DB_DRIVER",
		"postgres"), "DB driver")
	flag.StringVar(&o.DBScripts, "db-scripts", env.LookupEnv("DB_SCRIPTS"), "DB scripts to run at startup")
	flag.StringVar(&o.Admins, "admins", env.LookupEnv("ADMINS"), "Admins")
	flag.StringVar(&o.EmailFrom, "email-from", env.LookupEnv("EMAIL_FROM"), "Email from, default is info@flatfeestack.io")
	flag.StringVar(&o.EmailFromName, "email-from-name", env.LookupEnv("EMAIL_FROM_NAME"), "Email from name, default is a empty string")
	flag.StringVar(&o.EmailUrl, "email-url", env.LookupEnv("EMAIL_URL",
		"http://localhost"), "Email service URL")
	flag.StringVar(&o.EmailToken, "email-token", env.LookupEnv("EMAIL_TOKEN"), "Email service token")
	flag.StringVar(&o.EmailLinkPrefix, "email-prefix", env.LookupEnv("EMAIL_PREFIX",
		"http://localhost/"), "Email link prefix")
	flag.StringVar(&o.EmailMarketing, "email-marketing", env.LookupEnv("EMAIL_MARKETING",
		"tom.marketing@bocek.ch"), "Email marketing email. Set the value to 'live' to send out real emails")
	flag.StringVar(&o.NowpaymentsToken, "nowpayments-token", env.LookupEnv("NOWPAYMENTS_TOKEN"), "Token for NOWPayments access")
	flag.StringVar(&o.NowpaymentsIpnKey, "nowpayments-ipn-key", env.LookupEnv("NOWPAYMENTS_IPN_KEY"), "Key for NOWPayments IPN")
	flag.StringVar(&o.NowpaymentsApiUrl, "nowpayments-api-url", env.LookupEnv("NOWPAYMENTS_API_URL",
		"https://api.sandbox.nowpayments.io/v1"), "NOWPayments API URL")
	flag.StringVar(&o.NowpaymentsIpnCallbackUrl, "nowpayments-ipn-callback-url", env.LookupEnv("NOWPAYMENTS_IPN_CALLBACK_URL"), "Callback URL for NOWPayments IPN")

	flag.StringVar(&o.BackendUsername, "backend-username", env.LookupEnv("BACKEND_USERNAME"), "Username for accessing backend API")
	flag.StringVar(&o.BackendPassword, "backend-password", env.LookupEnv("BACKEND_PASSWORD"), "Password for accessing backend API")

	flag.StringVar(&o.Analyzer.Url, "analyzer-url", env.LookupEnv("ANALYZER_URL"), "URL to analysis engine")
	flag.StringVar(&o.Analyzer.Username, "analyzer-username", env.LookupEnv("ANALYZER_USERNAME"), "Username to analysis engine")
	flag.StringVar(&o.Analyzer.Password, "analyzer-password", env.LookupEnv("ANALYZER_PASSWORD"), "Password to analysis engine")

	flag.StringVar(&o.Payout.Url, "payout-url", env.LookupEnv("PAYOUT_URL"), "URL to payout")
	flag.StringVar(&o.Payout.Username, "payout-username", env.LookupEnv("PAYOUT_USERNAME"), "Username to payout")
	flag.StringVar(&o.Payout.Password, "payout-password", env.LookupEnv("PAYOUT_PASSWORD"), "Password to payout")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults, be explicit
	if o.Env == "local" || o.Env == "dev" {
		debug = true
		utils.SetDebug(true)
		log.SetLevel(log.DebugLevel)
	} else {
		utils.SetDebug(false)
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

	return o
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

	clients.InitAnalyzer(opts.Analyzer.Url, opts.Analyzer.Password, opts.Analyzer.Username)
	clients.InitPayout(opts.Payout.Url, opts.Payout.Password, opts.Payout.Username)
	clients.InitEmail(opts.EmailUrl, opts.EmailFromName, opts.EmailFrom, opts.EmailToken, opts.Env, opts.EmailMarketing, opts.EmailLinkPrefix)
	api.InitStripe(opts.StripeAPISecretKey, opts.StripeWebhookSecretKey)
	api.InitNow(opts.NowpaymentsApiUrl, opts.NowpaymentsToken, opts.NowpaymentsIpnCallbackUrl, opts.NowpaymentsIpnKey)
	api.InitApi(opts.StripeAPIPublicKey, opts.Env)

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	err = dbLib.InitDb(opts.DBDriver, opts.DBPath, opts.DBScripts)
	if err != nil {
		log.Fatal(err)
	}

	stripe.Key = opts.StripeAPISecretKey

	credentials := auth.Credentials{
		Username: opts.BackendUsername,
		Password: opts.BackendPassword,
	}

	// Routes
	router := mux.NewRouter()
	router.Use(prom.PrometheusMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return logRequestHandler(next)
	})
	//apiRouter := router.PathPrefix("/backend").Subrouter()
	registry := prom.CreateRegistry()
	router.Path("/metrics").Handler(promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			Registry: registry,
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	//user
	router.HandleFunc("/users/me", jwtAuthUser(api.GetMyUser)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwtAuthUser(api.GetMyConnectedEmails)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwtAuthUser(api.AddGitEmail)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/confirm", jwtAuthUser(api.ConfirmConnectedEmails)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/{email}", jwtAuthUser(api.RemoveGitEmail)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/method/{method}", jwtAuthUser(api.UpdateMethod)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/method", jwtAuthUser(api.DeleteMethod)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/sponsored", jwtAuthUser(api.GetSponsoredRepos)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/name/{name}", jwtAuthUser(api.UpdateName)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/clear/name", jwtAuthUser(api.ClearName)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/image", maxBytes(jwtAuthUser(api.UpdateImage), 200*1024)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/image", jwtAuthUser(api.DeleteImage)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(api.SetupStripe)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(api.CancelSub)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwtAuthUser(api.StripePaymentInitial)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwtAuthUser(api.NowPayment)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwtAuthUser(api.StatusSponsoredUsers)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/request-payout/{targetCurrency}", jwtAuthUser(api.RequestPayout)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/balance", jwtAuthUser(api.UserBalance)).Methods(http.MethodGet)
	router.HandleFunc("/users/contrib-snd", jwtAuthUser(api.ContributionsSend)).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwtAuthUser(api.ContributionsRcv)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwtAuthUser(api.ContributionsSum)).Methods(http.MethodPost)
	router.HandleFunc("/users/summary/{uuid}", api.UserSummary2).Methods(http.MethodGet)
	router.HandleFunc("/users/by/{email}", auth.BasicAuth(credentials, api.GetUserByEmail)).Methods(http.MethodGet)

	//payment
	router.HandleFunc("/users/me/stripe", jwtAuthUser(api.SetupStripe)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwtAuthUser(api.CancelSub)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwtAuthUser(api.StripePaymentInitial)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwtAuthUser(api.NowPayment)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwtAuthUser(api.StatusSponsoredUsers)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/payment", jwtAuthUser(api.PaymentEvent)).Methods(http.MethodGet)

	// get public user
	router.HandleFunc("/users/{id}", api.GetUserById).Methods(http.MethodGet)

	//contributions
	router.HandleFunc("/users/contrib-snd", jwtAuthUser(api.ContributionsSend)).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwtAuthUser(api.ContributionsRcv)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwtAuthUser(api.ContributionsSum)).Methods(http.MethodPost)
	router.HandleFunc("/users/contributions-summary/{uuid}", api.ContributionsSum2).Methods(http.MethodGet)

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
	router.HandleFunc("/hooks/analyzer", auth.BasicAuth(credentials, api.AnalysisEngineHook)).Methods(http.MethodPost)

	//admin
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
