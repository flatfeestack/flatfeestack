package main

import (
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	env "github.com/flatfeestack/go-lib/environment"
	prom "github.com/flatfeestack/go-lib/prometheus"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

var (
	opts  *Opts
	debug bool
)

type Opts struct {
	Port               int
	Env                string
	HS256              string
	BackendToken       string
	BackendCallbackUrl string
	GitBasePath        string
	AnalyzerUsername   string
	AnalyzerPassword   string
	BackendUsername    string
	BackendPassword    string
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func NewOpts() *Opts {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &Opts{}

	flag.StringVar(&o.Env, "env", env.LookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", env.LookupEnvInt("PORT", 9083), "listening HTTP port")
	flag.StringVar(&o.GitBasePath, "git-base", env.LookupEnv("GIT_BASE", "/tmp"), "Git base storage path")

	flag.StringVar(&o.AnalyzerUsername, "analyzer-username", env.LookupEnv("ANALYZER_USERNAME"), "Username for accessing API")
	flag.StringVar(&o.AnalyzerPassword, "analyzer-password", env.LookupEnv("ANALYZER_PASSWORD"), "Password for accessing API")

	flag.StringVar(&o.BackendCallbackUrl, "callback", env.LookupEnv("BACKEND_CALLBACK_URL"), "Callback URL")
	flag.StringVar(&o.BackendUsername, "backend-username", env.LookupEnv("BACKEND_USERNAME"), "Username for accessing backend API")
	flag.StringVar(&o.BackendPassword, "backend-password", env.LookupEnv("BACKEND_PASSWORD"), "Password for accessing backend API")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
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

	return o
}

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

	router := mux.NewRouter().StrictSlash(true)
	registry := prom.CreateRegistry()
	router.Path("/metrics").Handler(promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			Registry: registry,
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	router.HandleFunc("/analyze", basicAuth(analyze)).Methods("POST")
	log.Println("Starting analysis on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}
