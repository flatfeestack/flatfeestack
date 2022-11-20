package main

import (
	"encoding/base32"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	neo "github.com/nspcc-dev/neo-go/pkg/rpcclient"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Signature struct {
	Raw  []byte   `json:"raw"`
	Hash [32]byte `json:"hash"`
	R    [32]byte `json:"r"`
	S    [32]byte `json:"s"`
	V    uint8    `json:"v"`
}

type Timewarp struct {
	Offset int `json:"offset"`
}

type Blockchain struct {
	Contract   string
	PrivateKey string
	Url        string
	Deploy     bool
}

type Opts struct {
	Port     int
	Env      string
	HS256    string
	Ethereum Blockchain
	NEO      Blockchain
}

var (
	opts       *Opts
	jwtKey     []byte
	ethClient  *ClientETH
	neoClient  *neo.Client
	debug      bool
	secondsAdd int
)

func NewOpts() *Opts {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &Opts{}

	flag.StringVar(&o.Env, "env", lookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT",
		9084), "listening HTTP port")
	flag.StringVar(&o.Ethereum.PrivateKey, "eth-private-key", lookupEnv("ETH_PRIVATE_KEY",
		"4d5db4107d237df6a3d58ee5f70ae63d73d7658d4026f2eefd2f204c81682cb7"), "Ethereum private key")
	flag.StringVar(&o.Ethereum.Contract, "eth-contract", lookupEnv("ETH_CONTRACT",
		"0x731a10897d267e19b34503ad902d0a29173ba4b1"), "Ethereum contract address")
	flag.StringVar(&o.Ethereum.Url, "eth-url", lookupEnv("ETH_URL",
		"http://openethereum:8545"), "Ethereum URL")
	flag.BoolVar(&o.Ethereum.Deploy, "eth-deploy", lookupEnv("ETH_DEPLOY") == "true", "Set to true to deploy ETH contract")
	flag.StringVar(&o.NEO.PrivateKey, "neo-private-key", lookupEnv("NEO_PRIVATE_KEY",
		"L3WX5hiSstmFZBbr5Yyyvce1DoBZcQDgKn4xLeTdJHxsx7XcF3mp"), "NEO private key")
	flag.StringVar(&o.NEO.Contract, "neo-contract", lookupEnv("NEO_CONTRACT",
		"0x731a10897d267e19b34503ad902d0a29173ba4b1"), "NEO contract address")
	flag.StringVar(&o.NEO.Url, "neo-url", lookupEnv("NEO_URL",
		"http://172.17.0.1:8545"), "NEO URL")
	flag.BoolVar(&o.NEO.Deploy, "neo-deploy", lookupEnv("NEO_DEPLOY") == "true", "Set to true to deploy NEO contract")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if o.HS256 != "" {
		jwtKey, err = base32.StdEncoding.DecodeString(opts.HS256)
		if err != nil {
			log.Fatalf("cannot decode %v", opts.HS256)
		}
	}

	if strings.HasPrefix(o.Ethereum.PrivateKey, "0x") {
		o.Ethereum.PrivateKey = o.Ethereum.PrivateKey[2:]
	}

	//set defaults
	if o.Env == "local" || o.Env == "dev" {
		debug = true
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

func ethInit() *ClientETH {
	now := time.Now()
	ethClient, err := getEthClient(opts.Ethereum.Url, opts.Ethereum.PrivateKey, opts.Ethereum.Deploy, opts.Ethereum.Contract)
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(time.Second)
		ethClient, err = getEthClient(opts.Ethereum.Url, opts.Ethereum.PrivateKey, opts.Ethereum.Deploy, opts.Ethereum.Contract)
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
		log.Fatal("Could not initialize NEO network", err)
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

	// only internal routes, not accessible through caddy server
	router := mux.NewRouter()
	router.HandleFunc("/admin/sign/{userId}/{totalPayedOut}", jwtAuth(sign)).Methods(http.MethodPost)
	router.HandleFunc("/admin/timewarp/{hours}", jwtAuth(timeWarp)).Methods(http.MethodPost)
	router.HandleFunc("/admin/timewarp", jwtAuth(timeWarpOffset)).Methods(http.MethodGet)

	log.Printf("listing on port %v", opts.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}
