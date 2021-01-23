package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
)

type Payout struct {
	Address      string `json:"address"`
	Amount       int64  `json:"amount_micro_USD"`
	ExchangeRate string `json:"exchange_rate_USD_ETH"`
}

type PayoutResponse struct {
	TxHash string `json:"tx_hash"`
}

type Opts struct {
	Port          int
	Env           string
	EthContract   string
	EthPrivateKey string
	EthUrl        string
}

var (
	opts     *Opts
	EthWei   = big.NewFloat(0)
	MicroUsd = big.NewFloat(0)
	UsdWei   = big.NewFloat(0)
	client   *ClientETH
	debug    bool
)

func NewOpts() *Opts {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &Opts{}
	flag.StringVar(&o.Env, "env", lookupEnv("ENV",
		"local"), "ENV variable")
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT",
		9084), "listening HTTP port")
	flag.StringVar(&o.EthPrivateKey, "eth-private-key", lookupEnv("ETH_PRIVATE_KEY",
		"4d5db4107d237df6a3d58ee5f70ae63d73d7658d4026f2eefd2f204c81682cb7"), "Ethereum private key")
	flag.StringVar(&o.EthContract, "eth-contract", lookupEnv("ETH_CONTRACT",
		"0x731a10897d267e19b34503ad902d0a29173ba4b"), "Ethereum contract address")
	flag.StringVar(&o.EthUrl, "eth-url", lookupEnv("ETH_URL",
		"http://172.17.0.1:8545"), "Ethereum URL")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

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

func main() {
	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	EthWei.SetString("1000000000000000000")
	MicroUsd.SetString("1000000")
	UsdWei.SetString("1000000000000") //EthWei/MicroUsd

	opts = NewOpts()

	client, err = NewClientETH(opts.EthUrl, opts.EthPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	// only internal routes, not accessible through caddy server
	router := mux.NewRouter()
	router.HandleFunc("/pay", PaymentRequestHandler).Methods("POST", "OPTIONS")

	log.Printf("listing on port %v", opts.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}

func PaymentRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data []Payout
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Could not decode Webhook Body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var amountWei []*big.Int
	var addresses []string

	for _, v := range data {
		var flt *big.Float
		flt, _, err = big.ParseFloat(data[0].ExchangeRate, 10, 128, big.ToZero)
		amount := new(big.Float)
		amount.SetInt64(v.Amount)
		amount = amount.Mul(amount, UsdWei)
		amount = amount.Quo(amount, flt)
		i, _ := amount.Int(nil)
		amountWei = append(amountWei, i)
		addresses = append(addresses, v.Address)
	}

	log.Printf("received payout request for %v addresses", len(data))

	txHash, err := client.fill(addresses, amountWei)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "authorization header not set")
		return
	}

	p := PayoutResponse{TxHash: txHash}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
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
