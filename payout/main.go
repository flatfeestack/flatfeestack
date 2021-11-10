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

// USD
type PayoutUsd struct {
	Address      string `json:"address"`
	Balance      int64  `json:"balance_micro_USD"`
	ExchangeRate string `json:"exchange_rate_USD_ETH"`
}

type PayoutWei struct {
	Address string  `json:"address"`
	Balance big.Int `json:"balance_wei"`
}

type PayoutResponse struct {
	TxHash     string      `json:"tx_hash"`
	PayoutWeis []PayoutWei `json:"payout_weis"`
}

//Crypto

type PayoutCryptoRequest struct {
	Address string `json:"address"`
	Balance int64  `json:"nano_tea"`
}

type PayoutCrypto struct {
	Address string  `json:"address"`
	Balance big.Int `json:"balance"`
}

type PayoutCryptoResponse struct {
	TxHash        string         `json:"tx_hash"`
	PayoutCryptos []PayoutCrypto `json:"payout_cryptos"`
}

type Opts struct {
	Port          int
	Env           string
	EthContract   string
	EthPrivateKey string
	EthUrl        string
	Deploy        bool
}

var (
	opts         *Opts
	EthWei       = big.NewFloat(0)
	MicroUsd     = big.NewFloat(0)
	UsdWei       = big.NewFloat(0)
	CryptoFactor = big.NewFloat(0)
	ethClient    *ClientETH
	debug        bool
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
	flag.StringVar(&o.EthPrivateKey, "eth-private-key", lookupEnv("ETH_PRIVATE_KEY",
		"4d5db4107d237df6a3d58ee5f70ae63d73d7658d4026f2eefd2f204c81682cb7"), "Ethereum private key")
	flag.StringVar(&o.EthContract, "eth-contract", lookupEnv("ETH_CONTRACT",
		"0x731a10897d267e19b34503ad902d0a29173ba4b1"), "Ethereum contract address")
	flag.StringVar(&o.EthUrl, "eth-url", lookupEnv("ETH_URL",
		"http://172.17.0.1:8545"), "Ethereum URL")
	flag.BoolVar(&o.Deploy, "deploy", lookupEnv("DEPLOY") == "true", "Set to true to deploy contract")

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
	UsdWei.SetString("1000000000000")    //EthWei/MicroUsd
	CryptoFactor.SetString("1000000000") // Fixed factor for the moment

	opts = NewOpts()

	ethClient, err = getEthClient(opts.EthUrl, opts.EthPrivateKey, opts.Deploy, opts.EthContract)
	if err != nil {
		log.Fatal(err)
	}

	// only internal routes, not accessible through caddy server
	router := mux.NewRouter()
	router.HandleFunc("/pay", PaymentRequestHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/pay-crypto/{currency}", PaymentCryptoRequestHandler).Methods("POST", "OPTIONS")

	log.Printf("listing on port %v", opts.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}

func PaymentRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data []PayoutUsd
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Could not decode Webhook Body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var amountWei []*big.Int
	var addresses []string
	var payoutWei []PayoutWei

	for _, v := range data {
		var flt *big.Float
		flt, _, err = big.ParseFloat(data[0].ExchangeRate, 10, 128, big.ToZero)
		if flt.Cmp(big.NewFloat(0)) == 0 {
			writeErr(w, http.StatusBadRequest, "exchange rate is zero, cannot calculate")
			return
		}
		balance := new(big.Float)
		balance.SetInt64(v.Balance)
		balance = balance.Mul(balance, UsdWei)
		balance = balance.Quo(balance, flt)
		i, _ := balance.Int(nil)
		amountWei = append(amountWei, i)
		addresses = append(addresses, v.Address)
		payoutWei = append(payoutWei, PayoutWei{
			Address: v.Address,
			Balance: *i,
		})
	}

	log.Printf("received payout request for %v addresses", len(data))

	if len(data) == 0 {
		log.Printf("no data received, don't write on the chain")
		return
	}

	if opts.Env == "local" || opts.Env == "dev" {
		for k := range addresses {
			log.Printf("sending %v wei to %s", amountWei[k], addresses[k])
		}
	}

	txHash, err := payoutEth(ethClient, addresses, amountWei)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "authorization header not set")
		return
	}

	p := PayoutResponse{TxHash: txHash.Hash().String(), PayoutWeis: payoutWei}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
}

func PaymentCryptoRequestHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	cur := params["currency"]
	w.Header().Set("Content-Type", "application/json")
	var data []PayoutCryptoRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Could not decode Webhook Body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var amount []*big.Int
	var addresses []string
	var payoutCrypto []PayoutCrypto

	for _, v := range data {
		balance := new(big.Int)
		balance.SetInt64(v.Balance)
		amount = append(amount, balance)
		addresses = append(addresses, v.Address)
		payoutCrypto = append(payoutCrypto, PayoutCrypto{
			Address: v.Address,
			Balance: *balance,
		})
	}

	txHash, err := "", nil
	switch cur {
	case "eth":
		transaction, err := payoutEth(ethClient, addresses, amount)
		if err != nil {
			log.Fatal(err)
		}
		txHash = transaction.Hash().String()
		break
	case "neo":
		break
	case "xtc":
		break
	default:
		log.Printf("Currency isn't supported %v", err)
		w.Header().Set("Content-Type", "application/json")
		writeErr(w, http.StatusBadRequest, "Currency isn't supported %v", err)
	}

	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
	p := PayoutCryptoResponse{TxHash: txHash, PayoutCryptos: payoutCrypto}
	log.Printf(p.TxHash)
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		return
	}
	return
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
