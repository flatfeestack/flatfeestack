package main

import (
	api2 "backend/internal/api"
	"backend/internal/app"
	"backend/internal/client"
	"backend/internal/cron"
	"backend/pkg/config"
	util2 "backend/pkg/middleware"
	"backend/pkg/util"
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/flatfeestack/go-lib/auth"
	dbLib "github.com/flatfeestack/go-lib/database"
	env "github.com/flatfeestack/go-lib/environment"
	prom "github.com/flatfeestack/go-lib/prometheus"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
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
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		util.SetDebug(false)
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	if cfg.HS256 != "" {
		var err error
		cfg.JwtKey, err = base32.StdEncoding.DecodeString(cfg.HS256)
		if err != nil {
			h := sha256.New()
			h.Write([]byte(cfg.HS256))
			cfg.JwtKey = h.Sum(nil)
			slog.Debug("jwtKey: %v", jwtKey)
		}
	} else {
		slog.Error("HS256 seed is required, non was provided")
	}

	cfg.AdminsParsed = strings.Split(cfg.Admins, ";")

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
		slog.Info("Could not find .env file, using defaults")
	}
	//this will set the default ENVs
	parseFlags()

	ac := client.NewAnalysisClient(cfg.AnalyzerUrl, cfg.AnalyzerPassword, cfg.AnalyzerUsername)
	gc := client.NewGithubClient()
	ec := client.NewEmailClient(cfg.EmailUrl, cfg.EmailFromName, cfg.EmailFrom, cfg.EmailToken, cfg.Env, cfg.EmailMarketing, cfg.EmailLinkPrefix)

	ah := api2.NewApiHandler(cfg.StripeAPIPublicKey, cfg.Env)
	nh := api2.NewPaymentNowHandler(ec, cfg.NowpaymentsApiUrl, cfg.NowpaymentsToken, cfg.NowpaymentsIpnCallbackUrl, cfg.NowpaymentsIpnKey)
	sh := api2.NewPaymentHandler(ec, cfg.StripeAPISecretKey, cfg.StripeWebhookSecretKey)
	rh := api2.NewRepoHandler(ac, gc)
	eh := api2.NewEmailHandler(ec)
	rr := api2.NewResourceHandler(cfg)

	f, err := os.Open("cmd/banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		slog.Info("could not display banner...")
	}

	err = dbLib.InitDb(cfg.DBDriver, cfg.DBPath, cfg.DBScripts)
	if err != nil {
		slog.Error("DB not initialized",
			slog.Any("error", err))
		os.Exit(1)
	}

	//stripe.Key = cfg.StripeAPISecretKey

	credentials := auth.Credentials{
		Username: cfg.BackendUsername,
		Password: cfg.BackendPassword,
	}

	// Routes
	router := mux.NewRouter()
	router.Use(prom.PrometheusMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return util2.LogRequestHandler(next)
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

	jwt := util2.NewJwtHandler(cfg)
	jwtUser := util2.NewJwtUserHandler(cfg)

	router.HandleFunc("/users/me", jwt.JwtAuth(jwtUser.JwtUser(api2.GetMyUser))).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwt.JwtAuth(jwtUser.JwtUser(api2.GetMyConnectedEmails))).Methods(http.MethodGet)
	router.HandleFunc("/users/me/git-email", jwt.JwtAuth(jwtUser.JwtUser(eh.AddGitEmail))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/confirm", jwt.JwtAuth(jwtUser.JwtUser(api2.ConfirmConnectedEmails))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/git-email/{email}", jwt.JwtAuth(jwtUser.JwtUser(api2.RemoveGitEmail))).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/method/{method}", jwt.JwtAuth(jwtUser.JwtUser(api2.UpdateMethod))).Methods(http.MethodPut)
	router.HandleFunc("/users/me/method", jwt.JwtAuth(jwtUser.JwtUser(api2.DeleteMethod))).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/sponsored", jwt.JwtAuth(jwtUser.JwtUser(api2.GetSponsoredRepos))).Methods(http.MethodGet)
	router.HandleFunc("/users/me/name/{name}", jwt.JwtAuth(jwtUser.JwtUser(api2.UpdateName))).Methods(http.MethodPut)
	router.HandleFunc("/users/me/clear/name", jwt.JwtAuth(jwtUser.JwtUser(api2.ClearName))).Methods(http.MethodPut)
	router.HandleFunc("/users/me/image", util2.MaxBytes(jwt.JwtAuth(jwtUser.JwtUser(api2.UpdateImage)), 200*1024)).Methods(http.MethodPost)
	router.HandleFunc("/users/me/image", jwt.JwtAuth(jwtUser.JwtUser(api2.DeleteImage))).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe", jwt.JwtAuth(jwtUser.JwtUser(sh.SetupStripe))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwt.JwtAuth(jwtUser.JwtUser(api2.CancelSub))).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwt.JwtAuth(jwtUser.JwtUser(api2.StripePaymentInitial))).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwt.JwtAuth(jwtUser.JwtUser(api2.NowPayment))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwt.JwtAuth(jwtUser.JwtUser(api2.StatusSponsoredUsers))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/request-payout/{targetCurrency}", jwt.JwtAuth(jwtUser.JwtUser(rr.RequestPayout))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/balance", jwt.JwtAuth(jwtUser.JwtUser(api2.UserBalance))).Methods(http.MethodGet)
	router.HandleFunc("/users/contrib-snd", jwt.JwtAuth(jwtUser.JwtUser(api2.ContributionsSend))).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwt.JwtAuth(jwtUser.JwtUser(api2.ContributionsRcv))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwt.JwtAuth(jwtUser.JwtUser(api2.ContributionsSum))).Methods(http.MethodPost)
	router.HandleFunc("/users/summary/{uuid}", api2.UserSummary2).Methods(http.MethodGet)
	router.HandleFunc("/users/by/{email}", auth.BasicAuth(credentials, api2.GetUserByEmail)).Methods(http.MethodGet)

	//payment
	router.HandleFunc("/users/me/stripe", jwt.JwtAuth(jwtUser.JwtUser(sh.SetupStripe))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/stripe", jwt.JwtAuth(jwtUser.JwtUser(api2.CancelSub))).Methods(http.MethodDelete)
	router.HandleFunc("/users/me/stripe/{freq}/{seats}", jwt.JwtAuth(jwtUser.JwtUser(api2.StripePaymentInitial))).Methods(http.MethodPut)
	router.HandleFunc("/users/me/nowPayment/{freq}/{seats}", jwt.JwtAuth(jwtUser.JwtUser(api2.NowPayment))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/sponsored-users", jwt.JwtAuth(jwtUser.JwtUser(api2.StatusSponsoredUsers))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/payment", jwt.JwtAuth(jwtUser.JwtUser(api2.PaymentEvent))).Methods(http.MethodGet)

	// get public user
	router.HandleFunc("/users/{id}", api2.GetUserById).Methods(http.MethodGet)

	//contributions
	router.HandleFunc("/users/contrib-snd", jwt.JwtAuth(jwtUser.JwtUser(api2.ContributionsSend))).Methods(http.MethodPost)
	router.HandleFunc("/users/contrib-rcv", jwt.JwtAuth(jwtUser.JwtUser(api2.ContributionsRcv))).Methods(http.MethodPost)
	router.HandleFunc("/users/me/contributions-summary", jwt.JwtAuth(jwtUser.JwtUser(api2.ContributionsSum))).Methods(http.MethodPost)
	router.HandleFunc("/users/contributions-summary/{uuid}", api2.ContributionsSum2).Methods(http.MethodGet)

	//github
	router.HandleFunc("/repos/search", jwt.JwtAuth(jwtUser.JwtUser(rh.SearchRepoGitHub))).Methods(http.MethodGet)
	//repo
	router.HandleFunc("/repos/{id}", jwt.JwtAuth(jwtUser.JwtUser(api2.GetRepoByID))).Methods(http.MethodGet)
	router.HandleFunc("/repos/{id}/tag", jwt.JwtAuth(jwtUser.JwtUser(rh.TagRepo))).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/untag", jwt.JwtAuth(jwtUser.JwtUser(rh.UnTagRepo))).Methods(http.MethodPost)
	router.HandleFunc("/repos/{id}/{offset}/graph", jwt.JwtAuth(jwtUser.JwtUser(api2.Graph))).Methods(http.MethodGet)
	//payment

	//hooks
	router.HandleFunc("/hooks/stripe", util2.MaxBytes(sh.StripeWebhook, 65536)).Methods(http.MethodPost)
	router.HandleFunc("/hooks/nowpayments", nh.NowWebhook).Methods(http.MethodPost)
	router.HandleFunc("/hooks/analyzer", auth.BasicAuth(credentials, api2.AnalysisEngineHook)).Methods(http.MethodPost)

	//admin
	router.HandleFunc("/admin/time", jwt.JwtAuth(jwtUser.JwtUser(util2.JwtAdmin(api2.ServerTime)))).Methods(http.MethodGet)
	router.HandleFunc("/admin/users", jwt.JwtAuth(jwtUser.JwtUser(util2.JwtAdmin(api2.Users)))).Methods(http.MethodPost)

	router.HandleFunc("/config", ah.Config).Methods(http.MethodGet)

	//dev settings
	if debug {
		router.HandleFunc("/admin/fake/user/{email}", jwt.JwtAuth(jwtUser.JwtUser(util2.JwtAdmin(api2.FakeUser)))).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/payment/{email}/{seats}", jwt.JwtAuth(jwtUser.JwtUser(util2.JwtAdmin(api2.FakePayment)))).Methods(http.MethodPost)
		router.HandleFunc("/admin/fake/contribution", jwt.JwtAuth(jwtUser.JwtUser(util2.JwtAdmin(api2.FakeContribution)))).Methods(http.MethodPost)
		router.HandleFunc("/admin/timewarp/{hours}", jwt.JwtAuth(jwtUser.JwtUser(util2.JwtAdmin(api2.TimeWarp)))).Methods(http.MethodPost)
	}

	//invite
	router.HandleFunc("/confirm/invite/{email}", jwt.JwtAuth(jwtUser.JwtUser(api2.ConfirmInvite))).Methods(http.MethodPost)
	router.HandleFunc("/invite", jwt.JwtAuth(jwtUser.JwtUser(api2.Invitations))).Methods(http.MethodGet)
	router.HandleFunc("/invite/by/{email}", jwt.JwtAuth(jwtUser.JwtUser(api2.InviteByDelete))).Methods(http.MethodDelete)
	router.HandleFunc("/invite/my/{email}", jwt.JwtAuth(jwtUser.JwtUser(api2.InviteMyDelete))).Methods(http.MethodDelete)
	router.HandleFunc("/invite/{email}", jwt.JwtAuth(jwtUser.JwtUser(api2.InviteOther))).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("[404] no route matched for", "url", r.URL, "method", r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	c := app.NewCalcHandler(ac, ec)
	//scheduler
	cron.CronJobDay(c.DailyRunner, util.TimeNow())
	cron.CronJobHour(c.HourlyRunner, util.TimeNow())

	slog.Info("Starting FlatFeeStack backend", "port", cfg.Port)
	err = http.ListenAndServe(":"+strconv.Itoa(cfg.Port), router)
	if err != nil {
		slog.Error("Server stopped",
			slog.Any("error", err))
	}
	cron.CronStop()
}
