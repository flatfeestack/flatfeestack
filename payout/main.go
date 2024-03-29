package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/flatfeestack/go-lib/auth"
	env "github.com/flatfeestack/go-lib/environment"
	prom "github.com/flatfeestack/go-lib/prometheus"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	neo "github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"payout/metrics"
	"strconv"
	"strings"
	"time"
)

type Timewarp struct {
	Offset int `json:"offset"`
}

type Blockchain struct {
	Contract   string
	PrivateKey string
	Url        string
	Deploy     bool
}

type DaoAddresses struct {
	Dao        string `json:"dao"`
	Membership string `json:"membership"`
	Wallet     string `json:"wallet"`
}

type Opts struct {
	Port           int
	Env            string
	HS256          string
	Ethereum       Blockchain
	NEO            Blockchain
	Usdc           Blockchain
	Admins         string
	Dao            DaoAddresses
	PayoutUsername string
	PayoutPassword string
}

var (
	opts       *Opts
	jwtKey     []byte
	ethClient  *ClientETH
	neoClient  *neo.Client
	debug      bool
	secondsAdd int
	admins     []string
)

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
	flag.IntVar(&o.Port, "port", env.LookupEnvInt("PORT",
		9084), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", env.LookupEnv("HS256"), "HS256 key")

	flag.StringVar(&o.Ethereum.PrivateKey, "eth-private-key", env.LookupEnv("ETH_PRIVATE_KEY"), "Ethereum private key")
	flag.StringVar(&o.Ethereum.Contract, "eth-contract", env.LookupEnv("ETH_CONTRACT"), "Ethereum contract address")
	flag.StringVar(&o.Ethereum.Url, "eth-url", env.LookupEnv("ETH_URL"), "Ethereum URL")

	flag.StringVar(&o.Usdc.PrivateKey, "usdc-private-key", env.LookupEnv("USDC_PRIVATE_KEY"), "USDC private key")
	flag.StringVar(&o.Usdc.Contract, "usdc-contract", env.LookupEnv("USDC_CONTRACT"), "USDC contract address")
	flag.StringVar(&o.Usdc.Url, "usdc-url", env.LookupEnv("USDC_URL"), "USDC URL")

	flag.StringVar(&o.NEO.PrivateKey, "neo-private-key", env.LookupEnv("NEO_PRIVATE_KEY"), "NEO private key")
	flag.StringVar(&o.NEO.Contract, "neo-contract", env.LookupEnv("NEO_CONTRACT"), "NEO contract address")
	flag.StringVar(&o.NEO.Url, "neo-url", env.LookupEnv("NEO_URL"), "NEO URL")

	flag.StringVar(&o.Dao.Dao, "dao-dao-address", env.LookupEnv("DAO_DAO_CONTRACT"), "Address of the main DAO contract")
	flag.StringVar(&o.Dao.Membership, "dao-membership-address", env.LookupEnv("DAO_MEMBERSHIP_CONTRACT"), "Address of the membership contract")
	flag.StringVar(&o.Dao.Wallet, "dao-wallet-address", env.LookupEnv("DAO_WALLET_CONTRACT"), "Address of the Wallet contract")

	flag.StringVar(&o.PayoutUsername, "payout-username", env.LookupEnv("PAYOUT_USERNAME"), "Username to payout")
	flag.StringVar(&o.PayoutPassword, "payout-password", env.LookupEnv("PAYOUT_PASSWORD"), "Password to payout")

	flag.StringVar(&o.Admins, "admins", env.LookupEnv("ADMINS"), "Admins")

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

	if o.HS256 != "" {
		h := sha256.New()
		h.Write([]byte(o.HS256))
		jwtKey = h.Sum(nil)
		log.Debugf("jwtKey: %v", jwtKey)
	} else {
		log.Fatalf("HS256 seed is required, non was provided")
	}

	admins = strings.Split(o.Admins, ";")

	if strings.HasPrefix(o.Ethereum.PrivateKey, "0x") {
		o.Ethereum.PrivateKey = o.Ethereum.PrivateKey[2:]
	}

	return o
}

func ethInit() *ClientETH {
	now := time.Now()
	ethClient, err := getEthClient(opts.Ethereum.Url, opts.Ethereum.PrivateKey)
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(time.Second)
		ethClient, err = getEthClient(opts.Ethereum.Url, opts.Ethereum.PrivateKey)
	}
	if err != nil {
		log.Fatal("Could not initialize ETH network", err)
	}
	return ethClient
}

func neoInit() *neo.Client {
	now := time.Now()
	neoClient, err := getNeoClient(opts.NEO.Url)
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(time.Second)
		neoClient, err = getNeoClient(opts.NEO.Url)
	}
	if err != nil {
		//log.Fatal("Could not initialize NEO network", err)
		log.Debugf("Could not initialize NEO network %v", err)
	}
	return neoClient
}

func timeNow() time.Time {
	if debug {
		return time.Now().Add(time.Duration(secondsAdd) * time.Second).UTC()
	} else {
		return time.Now().UTC()
	}
}

func main() {
	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	opts = NewOpts()

	ethClient = ethInit()
	neoClient = neoInit()

	credentials := auth.Credentials{
		Username: opts.PayoutUsername,
		Password: opts.PayoutPassword,
	}

	// only internal routes, not accessible through caddy server
	router := mux.NewRouter()

	router.Use(prom.PrometheusMiddleware)
	registry := prom.CreateRegistry()

	metrics.InitMetricsGauges(registry)
	metrics.InitMetricsCron(ethClient.c, opts.Usdc.Contract, opts.Ethereum.Contract)

	router.Path("/metrics").Handler(promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			Registry: registry,
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))

	//this can only be called by an internal server
	router.HandleFunc("/admin/sign/eth", auth.BasicAuth(credentials, signEth)).Methods(http.MethodPost)
	router.HandleFunc("/admin/sign/neo", auth.BasicAuth(credentials, signNeo)).Methods(http.MethodPost)
	router.HandleFunc("/admin/sign/usdc", auth.BasicAuth(credentials, signUsdc)).Methods(http.MethodPost)

	//this can be called from frontend, but only the admin
	if debug {
		router.HandleFunc("/admin/time", jwtAuth(jwtAuthAdmin(serverTime, admins))).Methods(http.MethodGet)
		router.HandleFunc("/admin/time/eth", jwtAuth(jwtAuthAdmin(serverTimeEth, admins))).Methods(http.MethodGet)
		router.HandleFunc("/admin/timewarp/{hours}", jwtAuth(jwtAuthAdmin(timeWarp, admins))).Methods(http.MethodPost)
	}
	//available for the public
	router.HandleFunc("/config/dao", daoConfig).Methods(http.MethodGet)
	router.HandleFunc("/config/payout", payoutConfig).Methods(http.MethodGet)

	log.Printf("listing on port %v", opts.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}
