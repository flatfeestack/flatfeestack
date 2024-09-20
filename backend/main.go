package main

import (
	api2 "backend/internal/api"
	"backend/internal/app"
	"backend/internal/client"
	"backend/internal/cron"
	"backend/internal/db"
	"backend/pkg/config"
	util2 "backend/pkg/middleware"
	"backend/pkg/util"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/dusted-go/logging/prettylog"
	"github.com/flatfeestack/go-lib/auth"
	dbLib "github.com/flatfeestack/go-lib/database"
	env "github.com/flatfeestack/go-lib/environment"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	cfg    *config.Config
	jwtKey []byte
	debug  bool
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
		prettyHandler := prettylog.NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   false,
			ReplaceAttr: nil,
		})
		slog.SetDefault(slog.New(prettyHandler))
	} else {
		util.SetDebug(false)
	}

	if cfg.HS256 != "" {
		var err error
		cfg.JwtKey, err = base32.StdEncoding.DecodeString(cfg.HS256)
		if err != nil {
			h := sha256.New()
			h.Write([]byte(cfg.HS256))
			cfg.JwtKey = h.Sum(nil)
			slog.Debug("jwtKey", slog.String("key", hex.EncodeToString(jwtKey)))
		}
	} else {
		slog.Error("HS256 seed is required, non was provided")
	}

	cfg.AdminsParsed = strings.Split(cfg.Admins, ";")

	if cfg.EmailFrom == "" {
		cfg.EmailFrom = "info@flatfeestack.io"
	}
}

func middlewareJwtAuthUserLog(handlerFunc func(http.ResponseWriter, *http.Request, *db.UserDetail)) func(w http.ResponseWriter, r *http.Request) {
	jwt := util2.NewJwtHandler(cfg)
	jwtUser := util2.NewJwtUserHandler(cfg)

	// Apply the jwtUser and jwtAuth middleware
	jwtHandler := jwt.JwtAuth(jwtUser.JwtUser(handlerFunc))

	// Apply the LogRequestHandler middleware
	loggedHandler := util2.LogRequestHandler(http.HandlerFunc(jwtHandler))

	return loggedHandler
}

func middlewareJwtAuthAdminLog(handlerFunc func(http.ResponseWriter, *http.Request, *db.UserDetail)) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request, u *db.UserDetail) {
		if u.Role != "admin" {
			slog.Error("not admin",
				slog.String("email", u.Email))
			util.WriteErrorf(w, http.StatusBadRequest, api2.GenericErrorMessage)
			return
		}
		handlerFunc(w, r, u)
	}
	return middlewareJwtAuthUserLog(fn)
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

	f, err := os.Open("banner.txt")
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
	router := http.NewServeMux()
	//router := mux.NewRouter()

	router.HandleFunc("GET /users/me", middlewareJwtAuthUserLog(api2.GetMyUser))
	router.HandleFunc("GET /users/me/git-email", middlewareJwtAuthUserLog(api2.GetMyConnectedEmails))
	router.HandleFunc("POST /users/me/git-email", middlewareJwtAuthUserLog(eh.AddGitEmail))
	router.HandleFunc("POST /users/me/git-email/confirm", middlewareJwtAuthUserLog(api2.ConfirmConnectedEmails))
	router.HandleFunc("DELETE /users/me/git-email/{email}", middlewareJwtAuthUserLog(api2.RemoveGitEmail))
	router.HandleFunc("PUT /users/me/method/{method}", middlewareJwtAuthUserLog(api2.UpdateMethod))
	router.HandleFunc("DELETE /users/me/method", middlewareJwtAuthUserLog(api2.DeleteMethod))
	router.HandleFunc("GET /users/me/sponsored", middlewareJwtAuthUserLog(api2.GetSponsoredRepos))
	router.HandleFunc("PUT /users/me/name/{name}", middlewareJwtAuthUserLog(api2.UpdateName))
	router.HandleFunc("PUT /users/me/clear/name", middlewareJwtAuthUserLog(api2.ClearName))
	router.HandleFunc("POST /users/me/image", util2.MaxBytes(middlewareJwtAuthUserLog(api2.UpdateImage), 200*1024))
	router.HandleFunc("DELETE /users/me/image", middlewareJwtAuthUserLog(api2.DeleteImage))
	router.HandleFunc("POST /users/me/request-payout/{targetCurrency}", middlewareJwtAuthUserLog(rr.RequestPayout))
	router.HandleFunc("GET /users/me/balance", middlewareJwtAuthUserLog(api2.UserBalance))
	router.HandleFunc("GET /users/summary/{uuid}", api2.UserSummary2)
	router.HandleFunc("GET /users/by/{email}", auth.BasicAuth(credentials, api2.GetUserByEmail))

	//payment
	router.HandleFunc("POST /users/me/stripe", middlewareJwtAuthUserLog(sh.SetupStripe))
	router.HandleFunc("DELETE /users/me/stripe", middlewareJwtAuthUserLog(api2.CancelSub))
	router.HandleFunc("PUT /user s/me/stripe/{freq}/{seats}", middlewareJwtAuthUserLog(api2.StripePaymentInitial))
	router.HandleFunc("POST /users/me/nowPayment/{freq}/{seats}", middlewareJwtAuthUserLog(api2.NowPayment))
	router.HandleFunc("POST /users/me/sponsored-users", middlewareJwtAuthUserLog(api2.StatusSponsoredUsers))
	router.HandleFunc("GET /users/me/payment", middlewareJwtAuthUserLog(api2.PaymentEvent))

	// get public user
	router.HandleFunc("GET /users/{id}", api2.GetUserById)

	//contributions
	router.HandleFunc("POST /users/contrib-snd", middlewareJwtAuthUserLog(api2.ContributionsSend))
	router.HandleFunc("POST /users/contrib-rcv", middlewareJwtAuthUserLog(api2.ContributionsRcv))
	router.HandleFunc("POST /users/me/contributions-summary", middlewareJwtAuthUserLog(api2.ContributionsSum))
	router.HandleFunc("GET /users/contributions-summary/{uuid}", api2.ContributionsSum2)

	//github
	router.HandleFunc("GET /repos/search", middlewareJwtAuthUserLog(rh.SearchRepoGitHub))
	//repo
	router.HandleFunc("GET /repos/{id}", middlewareJwtAuthUserLog(api2.GetRepoByID))
	router.HandleFunc("POST /repos/{id}/tag", middlewareJwtAuthUserLog(rh.TagRepo))
	router.HandleFunc("POST /repos/{id}/untag", middlewareJwtAuthUserLog(rh.UnTagRepo))
	router.HandleFunc("GET /repos/{id}/{offset}/graph", middlewareJwtAuthUserLog(api2.Graph))
	//payment

	//hooks
	router.HandleFunc("POST /hooks/stripe", util2.MaxBytes(sh.StripeWebhook, 65536))
	router.HandleFunc("POST /hooks/nowpayments", nh.NowWebhook)
	router.HandleFunc("POST /hooks/analyzer", auth.BasicAuth(credentials, api2.AnalysisEngineHook))

	//admin
	router.HandleFunc("GET /admin/time", middlewareJwtAuthAdminLog(api2.ServerTime))
	router.HandleFunc("POST /admin/users", middlewareJwtAuthAdminLog(api2.Users))

	router.HandleFunc("GET /config", ah.Config)

	//dev settings
	if debug {
		router.HandleFunc("POST /admin/fake/user/{email}", middlewareJwtAuthAdminLog(api2.FakeUser))
		router.HandleFunc("POST /admin/fake/payment/{email}/{seats}", middlewareJwtAuthAdminLog(api2.FakePayment))
		router.HandleFunc("POST /admin/fake/contribution", middlewareJwtAuthAdminLog(api2.FakeContribution))
		router.HandleFunc("POST /admin/timewarp/{hours}", middlewareJwtAuthAdminLog(api2.TimeWarp))
	}

	//invite
	router.HandleFunc("POST /confirm/invite/{email}", middlewareJwtAuthUserLog(api2.ConfirmInvite))
	router.HandleFunc("GET /invite", middlewareJwtAuthUserLog(api2.Invitations))
	router.HandleFunc("DELETE /invite/by/{email}", middlewareJwtAuthUserLog(api2.InviteByDelete))
	router.HandleFunc("DELETE /invite/my/{email}", middlewareJwtAuthUserLog(api2.InviteMyDelete))
	router.HandleFunc("POST /invite/{email}", middlewareJwtAuthUserLog(api2.InviteOther))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
