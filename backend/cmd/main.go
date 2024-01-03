package main

import (
	"backend/api"
	"backend/internal/cron"
	"backend/pkg/util"
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

	"backend/pkg/config"
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
	cfg    *config.Config
	jwtKey []byte
	debug  bool
	admins []string
)

type Subsystem struct {
	Url      string
	Username string
	Password string
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func parseFlags() {
	cfg = &config.Config{}

	flag.StringVar(&cfg.Env, "env", env.LookupEnv("ENV"), "ENV variable")
	flag.IntVar(&cfg.Port, "port", env.LookupEnvInt("PORT",
		9082), "listening HTTP port")
	flag.StringVar(&cfg.HS256, "hs256", env.LookupEnv("HS256"), "HS256 key")
	flag.StringVar(&cfg.StripeAPISecretKey, "stripe-secret-api", env.LookupEnv("STRIPE_SECRET_API"), "Stripe API secret")
	flag.StringVar(&cfg.StripeAPIPublicKey, "stripe-public-api", env.LookupEnv("STRIPE_PUBLIC_API"), "Public Key for stripe")
	flag.StringVar(&cfg.StripeWebhookSecretKey, "stripe-secret-webhook", env.LookupEnv("STRIPE_SECRET_WEBHOOK"), "Stripe webhook secret")
	flag.StringVar(&cfg.DBPath, "db-path", env.LookupEnv("DB_PATH",
		"postgresql://postgres:password@db:5432/flatfeestack?sslmode=disable"), "DB path")
	flag.StringVar(&cfg.DBDriver, "db-driver", env.LookupEnv("DB_DRIVER",
		"postgres"), "DB driver")
	flag.StringVar(&cfg.DBScripts, "db-scripts", env.LookupEnv("DB_SCRIPTS"), "DB scripts to run at startup")
	flag.StringVar(&cfg.Admins, "admins", env.LookupEnv("ADMINS"), "Admins")
	flag.StringVar(&cfg.EmailFrom, "email-from", env.LookupEnv("EMAIL_FROM"), "Email from, default is info@flatfeestack.io")
	flag.StringVar(&cfg.EmailFromName, "email-from-name", env.LookupEnv("EMAIL_FROM_NAME"), "Email from name, default is a empty string")
	flag.StringVar(&cfg.EmailUrl, "email-url", env.LookupEnv("EMAIL_URL",
		"http://localhost"), "Email service URL")
	flag.StringVar(&cfg.EmailToken, "email-token", env.LookupEnv("EMAIL_TOKEN"), "Email service token")
	flag.StringVar(&cfg.EmailLinkPrefix, "email-prefix", env.LookupEnv("EMAIL_PREFIX",
		"http://localhost/"), "Email link prefix")
	flag.StringVar(&cfg.EmailMarketing, "email-marketing", env.LookupEnv("EMAIL_MARKETING",
		"tom.marketing@bocek.ch"), "Email marketing email. Set the value to 'live' to send out real emails")
	flag.StringVar(&cfg.NowpaymentsToken, "nowpayments-token", env.LookupEnv("NOWPAYMENTS_TOKEN"), "Token for NOWPayments access")
	flag.StringVar(&cfg.NowpaymentsIpnKey, "nowpayments-ipn-key", env.LookupEnv("NOWPAYMENTS_IPN_KEY"), "Key for NOWPayments IPN")
	flag.StringVar(&cfg.NowpaymentsApiUrl, "nowpayments-api-url", env.LookupEnv("NOWPAYMENTS_API_URL",
		"https://api.sandbox.nowpayments.io/v1"), "NOWPayments API URL")
	flag.StringVar(&cfg.NowpaymentsIpnCallbackUrl, "nowpayments-ipn-callback-url", env.LookupEnv("NOWPAYMENTS_IPN_CALLBACK_URL"), "Callback URL for NOWPayments IPN")

	flag.StringVar(&cfg.BackendUsername, "backend-username", env.LookupEnv("BACKEND_USERNAME"), "Username for accessing backend API")
	flag.StringVar(&cfg.BackendPassword, "backend-password", env.LookupEnv("BACKEND_PASSWORD"), "Password for accessing backend API")

	flag.StringVar(&cfg.AnalyzerUrl, "analyzer-url", env.LookupEnv("ANALYZER_URL"), "URL to analysis engine")
	flag.StringVar(&cfg.AnalyzerUsername, "analyzer-username", env.LookupEnv("ANALYZER_USERNAME"), "Username to analysis engine")
	flag.StringVar(&cfg.AnalyzerPassword, "analyzer-password", env.LookupEnv("ANALYZER_PASSWORD"), "Password to analysis engine")

	flag.StringVar(&cfg.NEOPrivateKey, "neo-private-key", env.LookupEnv("NEO_PRIVATE_KEY"), "NEO private key")
	flag.StringVar(&cfg.ETHPrivateKey, "eth-private-key", env.LookupEnv("ETH_PRIVATE_KEY"), "Ethereum private key")
	flag.StringVar(&cfg.ETHContractAddress, "eth-contract-address", env.LookupEnv("ETH_CONTRACT_ADDRESS"), "Ethereum contract address")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults, be explicit
	if cfg.Env == "local" || cfg.Env == "dev" {
		debug = true
		util.SetDebug(true)
		log.SetLevel(log.DebugLevel)
	} else {
		util.SetDebug(false)
		log.SetLevel(log.InfoLevel)
	}

	if cfg.HS256 != "" {
		var err error
		cfg.JwtKey, err = base32.StdEncoding.DecodeString(cfg.HS256)
		if err != nil {
			h := sha256.New()
			h.Write([]byte(cfg.HS256))
			cfg.JwtKey = h.Sum(nil)
			log.Debugf("jwtKey: %v", jwtKey)
		}
	} else {
		log.Fatalf("HS256 seed is required, non was provided")
	}

	admins = strings.Split(cfg.Admins, ";")

	if cfg.EmailFrom == "" {
		cfg.EmailFrom = "info@flatfeestack.io"
	}
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
	parseFlags()

	clients.InitAnalyzer(cfg.AnalyzerUrl, cfg.AnalyzerPassword, cfg.AnalyzerUsername)
	clients.InitEmail(cfg.EmailUrl, cfg.EmailFromName, cfg.EmailFrom, cfg.EmailToken, cfg.Env, cfg.EmailMarketing, cfg.EmailLinkPrefix)
	api.InitStripe(cfg.StripeAPISecretKey, cfg.StripeWebhookSecretKey)
	api.InitNow(cfg.NowpaymentsApiUrl, cfg.NowpaymentsToken, cfg.NowpaymentsIpnCallbackUrl, cfg.NowpaymentsIpnKey)
	api.InitApi(cfg.StripeAPIPublicKey, cfg.Env)

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	err = dbLib.InitDb(cfg.DBDriver, cfg.DBPath, cfg.DBScripts)
	if err != nil {
		log.Fatal(err)
	}

	stripe.Key = cfg.StripeAPISecretKey

	credentials := auth.Credentials{
		Username: cfg.BackendUsername,
		Password: cfg.BackendPassword,
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

	jwt := util.NewJwtHandler(cfg)

	router.HandleFunc("/users/me", jwt.JwtAuthUser(api.GetMyUser)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwt.JwtAuthUser(api.GetMyConnectedEmails)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwt.JwtAuthUser(api.AddGitEmail)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/confirm", jwt.JwtAuthUser(api.ConfirmConnectedEmails)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/{email}", jwt.JwtAuthUser(api.RemoveGitEmail)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/method/{method}", jwt.JwtAuthUser(api.UpdateMethod)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/method", jwt.JwtAuthUser(api.DeleteMethod)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/sponsored", jwt.JwtAuthUser(api.GetSponsoredRepos)).Methods(http.MethodGet)
	router.HandleFunc("/users/me/name/{name}", jwt.JwtAuthUser(api.UpdateName)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/clear/name", jwt.JwtAuthUser(api.ClearName)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/image", util.MaxBytes(jwt.JwtAuthUser(api.UpdateImage), 200*1024)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/image", jwt.JwtAuthUser(api.DeleteImage)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe", jwt.JwtAuthUser(api.SetupStripe)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwt.JwtAuthUser(api.CancelSub)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwt.JwtAuthUser(api.StripePaymentInitial)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwt.JwtAuthUser(api.NowPayment)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwt.JwtAuthUser(api.StatusSponsoredUsers)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/request-payout/{targetCurrency}", jwt.JwtAuthUser(api.RequestPayout)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/balance", jwt.JwtAuthUser(api.UserBalance)).Methods(http.MethodGet)
	router.HandleFunc("/users/contrib-snd", jwt.JwtAuthUser(api.ContributionsSend)).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwt.JwtAuthUser(api.ContributionsRcv)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwt.JwtAuthUser(api.ContributionsSum)).Methods(http.MethodPost)
	router.HandleFunc("/users/summary/{uuid}", api.UserSummary2).Methods(http.MethodGet)
	router.HandleFunc("/users/by/{email}", auth.BasicAuth(credentials, api.GetUserByEmail)).Methods(http.MethodGet)

	//payment
	router.HandleFunc("/users/me/stripe", jwt.JwtAuthUser(api.SetupStripe)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwt.JwtAuthUser(api.CancelSub)).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwt.JwtAuthUser(api.StripePaymentInitial)).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwt.JwtAuthUser(api.NowPayment)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwt.JwtAuthUser(api.StatusSponsoredUsers)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/payment", jwt.JwtAuthUser(api.PaymentEvent)).Methods(http.MethodGet)

	// get public user
	router.HandleFunc("/users/{id}", api.GetUserById).Methods(http.MethodGet)

	//contributions
	router.HandleFunc("/users/contrib-snd", jwt.JwtAuthUser(api.ContributionsSend)).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwt.JwtAuthUser(api.ContributionsRcv)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwt.JwtAuthUser(api.ContributionsSum)).Methods(http.MethodPost)
	router.HandleFunc("/users/contributions-summary/{uuid}", api.ContributionsSum2).Methods(http.MethodGet)

	//github
	router.HandleFunc("/repos/search", jwt.JwtAuthUser(api.SearchRepoGitHub)).Methods(http.MethodGet)
	//repo
	router.HandleFunc("/repos/{id}", jwt.JwtAuthUser(api.GetRepoByID)).Methods(http.MethodGet)
	router.HandleFunc("/repos/{id}/tag", jwt.JwtAuthUser(api.TagRepo)).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/untag", jwt.JwtAuthUser(api.UnTagRepo)).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/{offset}/graph", jwt.JwtAuthUser(api.Graph)).Methods(http.MethodGet)
	//payment

	//hooks
	router.HandleFunc("/hooks/stripe", util.MaxBytes(api.StripeWebhook, 65536)).Methods(http.MethodPost)
	router.HandleFunc("/hooks/nowpayments", api.NowWebhook).Methods(http.MethodPost)
	router.HandleFunc("/hooks/analyzer", auth.BasicAuth(credentials, api.AnalysisEngineHook)).Methods(http.MethodPost)

	//admin
	router.HandleFunc("/admin/time", jwt.JwtAuthAdmin(api.ServerTime, admins)).Methods(http.MethodGet)
	router.HandleFunc("/admin/users", jwt.JwtAuthAdmin(api.Users, admins)).Methods(http.MethodPost)

	router.HandleFunc("/config", api.Config).Methods(http.MethodGet)

	//dev settings
	if debug {
		router.HandleFunc("/admin/fake/user/{email}", jwt.JwtAuthAdmin(api.FakeUser, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/payment/{email}/{seats}", jwt.JwtAuthAdmin(api.FakePayment, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/contribution", jwt.JwtAuthAdmin(api.FakeContribution, admins)).Methods(http.MethodPost)
		router.HandleFunc("/admin/timewarp/{hours}", jwt.JwtAuthAdmin(api.TimeWarp, admins)).Methods(http.MethodPost)
	}

	//invite
	router.HandleFunc("/confirm/invite/{email}", jwt.JwtAuthUser(api.ConfirmInvite)).Methods(http.MethodPost)
	router.HandleFunc("/invite", jwt.JwtAuthUser(api.Invitations)).Methods(http.MethodGet)
	router.HandleFunc("/invite/by/{email}", jwt.JwtAuthUser(api.InviteByDelete)).Methods(http.MethodDelete)
	router.HandleFunc("/invite/my/{email}", jwt.JwtAuthUser(api.InviteMyDelete)).Methods(http.MethodDelete)
	router.HandleFunc("/invite/{email}", jwt.JwtAuthUser(api.InviteOther)).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[404] no route matched for: %s, %s", r.URL, r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	//scheduler
	cron.CronJobDay(dailyRunner, util.TimeNow())
	cron.CronJobHour(hourlyRunner, util.TimeNow())

	log.Println("Starting backend on port " + strconv.Itoa(cfg.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), router))
	cron.CronStop()
}
