package main

import (
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"forum/api"
	"forum/dao"
	"forum/globals"
	"forum/jwt"
	"forum/types"
	"forum/utils"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/dimiro1/banner"
	dbLib "github.com/flatfeestack/go-lib/database"
	env "github.com/flatfeestack/go-lib/environment"
	prom "github.com/flatfeestack/go-lib/prometheus"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func NewOpts() *types.Opts {
	o := &types.Opts{}
	flag.StringVar(&o.Env, "env", env.LookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", env.LookupEnvInt("PORT", 9086), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", env.LookupEnv("HS256", "test-seed"), "HS256 key")
	flag.StringVar(&o.DBPath, "db-path", env.LookupEnv("DB_PATH",
		"postgresql://postgres:password@localhost:5432/flatfeestack?sslmode=disable"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", env.LookupEnv("DB_DRIVER",
		"postgres"), "DB driver")
	flag.StringVar(&o.DBScripts, "db-scripts", env.LookupEnv("DB_SCRIPTS"), "DB scripts to run at startup")
	flag.StringVar(&o.Admins, "admins", env.LookupEnv("ADMINS"), "Admins")

	flag.StringVar(&o.BackendUrl, "backend-url", env.LookupEnv("BACKEND_URL"), "Backend URL")
	flag.StringVar(&o.BackendUsername, "backend-username", env.LookupEnv("BACKEND_USERNAME"), "Username for accessing backend API")
	flag.StringVar(&o.BackendPassword, "backend-password", env.LookupEnv("BACKEND_PASSWORD"), "Password for accessing backend API")

	flag.StringVar(&o.EthWsUrl, "eth-ws-url", env.LookupEnv("ETH_WS_URL"), "Websocket URL for ETH connection")
	flag.StringVar(&o.DaoContractAddress, "dao-contract-address", env.LookupEnv("DAO_CONTRACT_ADDRESS"), "DAO contract address")

	//set defaults, be explicit
	if o.Env == "local" || o.Env == "dev" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if o.HS256 != "" {
		var err error
		globals.JwtKey, err = base32.StdEncoding.DecodeString(o.HS256)
		if err != nil {
			h := sha256.New()
			h.Write([]byte(o.HS256))
			globals.JwtKey = h.Sum(nil)
			log.Debugf("jwtKey: %v", globals.JwtKey)
		}
	} else {
		log.Fatalf("HS256 seed is required, non was provided")
	}

	globals.ADMINS = strings.Split(o.Admins, ";")

	return o
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}
	globals.OPTS = NewOpts()

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	err = dbLib.InitDb(globals.OPTS.DBDriver, globals.OPTS.DBPath, globals.OPTS.DBScripts)
	if err != nil {
		log.Fatal(err)
	}

	dao.RunEventListener()

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal(err)
	}
	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	server := api.NewStrictServerImpl()
	options := api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			utils.WriteErrorf(w, http.StatusBadRequest, "Bad Request: %v", err.Error())
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			// Custom implementation for handling response errors
			utils.WriteErrorf(w, http.StatusInternalServerError, "Internal Server Error: %v", err.Error())
		},
	}
	serverInterface := api.NewStrictHandlerWithOptions(server, nil, options)

	validator := middleware.OapiRequestValidatorWithOptions(swagger, utils.EmptyOptions())

	registry := prom.CreateRegistry()

	router := mux.NewRouter()
	router.Use(prom.PrometheusMiddleware)
	router.Use(validator)
	router.Path("/metrics").Handler(promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			Registry: registry,
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))

	serverOptions := api.GorillaServerOptions{
		BaseURL:    "",
		BaseRouter: router,
		Middlewares: []api.MiddlewareFunc{
			jwt.AuthMiddleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			utils.WriteErrorf(w, http.StatusInternalServerError, "Internal Server Error: %v", err.Error())
			return
		},
	}

	handler := api.HandlerWithOptions(serverInterface, serverOptions)

	log.Println("Starting forum on port " + strconv.Itoa(globals.OPTS.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(globals.OPTS.Port), handler))

}
