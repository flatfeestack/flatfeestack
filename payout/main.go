package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"
)

type PayoutMeta struct {
	Currency string
	Tea      int64
}

type PayoutCryptoRequest struct {
	Address string       `json:"address"`
	NanoTea int64        `json:"nano_tea"`
	Meta    []PayoutMeta `json:"meta"`
}

type PayoutCrypto struct {
	Address          string       `json:"address"`
	NanoTea          int64        `json:"nano_tea"`
	SmartContractTea big.Int      `json:"smart_contract_tea"`
	Meta             []PayoutMeta `json:"meta"`
}

type PayoutCryptoResponse struct {
	TxHash        string         `json:"tx_hash"`
	PayoutCryptos []PayoutCrypto `json:"payout_cryptos"`
}

type Blockchain struct {
	Contract   string
	PrivateKey string
	Url        string
	Deploy     bool
	Factor     *big.Float
}

type Opts struct {
	Port            int
	Env             string
	Blockchains     map[string]Blockchain
	PayoutNodejsUrl string
}

var (
	opts                *Opts
	EthWei              = big.NewFloat(0)
	MicroUsd            = big.NewFloat(0)
	UsdWei              = big.NewFloat(0)
	defaultCryptoFactor = big.NewFloat(0)
	ethClient           *ClientETH
	neoClient           *client.Client
	debug               bool
)

func NewOpts() *Opts {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	eth := Blockchain{Factor: big.NewFloat(1000000000000000000)}
	neo := Blockchain{Factor: big.NewFloat(100000000)}
	xtz := Blockchain{Factor: big.NewFloat(1000000)}

	o := &Opts{}
	flag.StringVar(&o.Env, "env", lookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT",
		9084), "listening HTTP port")
	flag.StringVar(&eth.PrivateKey, "eth-private-key", lookupEnv("ETH_PRIVATE_KEY",
		"4d5db4107d237df6a3d58ee5f70ae63d73d7658d4026f2eefd2f204c81682cb7"), "Ethereum private key")
	flag.StringVar(&eth.Contract, "eth-contract", lookupEnv("ETH_CONTRACT",
		"0x731a10897d267e19b34503ad902d0a29173ba4b1"), "Ethereum contract address")
	flag.StringVar(&eth.Url, "eth-url", lookupEnv("ETH_URL",
		"http://openethereum:8545"), "Ethereum URL")
	flag.BoolVar(&eth.Deploy, "eth-deploy", lookupEnv("ETH_DEPLOY") == "true", "Set to true to deploy ETH contract")
	flag.StringVar(&neo.PrivateKey, "neo-private-key", lookupEnv("NEO_PRIVATE_KEY",
		"L3WX5hiSstmFZBbr5Yyyvce1DoBZcQDgKn4xLeTdJHxsx7XcF3mp"), "NEO private key")
	flag.StringVar(&neo.Contract, "neo-contract", lookupEnv("NEO_CONTRACT",
		"0x731a10897d267e19b34503ad902d0a29173ba4b1"), "NEO contract address")
	flag.StringVar(&neo.Url, "neo-url", lookupEnv("NEO_URL",
		"http://172.17.0.1:8545"), "NEO URL")
	flag.BoolVar(&neo.Deploy, "neo-deploy", lookupEnv("NEO_DEPLOY") == "true", "Set to true to deploy NEO contract")
	flag.StringVar(&o.PayoutNodejsUrl, "payout-nodejs-url", lookupEnv("PAYOUT_NODEJS_URL",
		"http://localhost:9086"), "Payout Nodejs Url")

	o.Blockchains = make(map[string]Blockchain)
	o.Blockchains["eth"] = eth
	o.Blockchains["neo"] = neo
	o.Blockchains["xtz"] = xtz

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

	defaultCryptoFactor.SetString("1000000000") // Fixed factor for the moment (Nano)

	opts = NewOpts()

	var eth = opts.Blockchains["eth"]
	now := time.Now()
	ethClient, err = getEthClient(eth.Url, eth.PrivateKey, eth.Deploy, eth.Contract)
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(time.Second)
		ethClient, err = getEthClient(eth.Url, eth.PrivateKey, eth.Deploy, eth.Contract)
	}
	if err != nil {
		log.Fatal("Could not initialize ETH network", err)
	}

	var neo = opts.Blockchains["neo"]
	now = time.Now()
	neoClient, err = client.New(context.TODO(), neo.Url, client.Options{})
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(time.Second)
		neoClient, err = client.New(context.TODO(), neo.Url, client.Options{})
	}
	if err != nil {
		log.Fatalf("Could not create a new NEO client", err)
	}

	err = neoClient.Init()
	if err != nil {
		log.Fatalf("Could not initialize NEO network", err)
	}

	contractOwnerPrivateKey, err := keys.NewPrivateKeyFromWIF(neo.PrivateKey)
	if err != nil {
		log.Fatalf("Could not transform NEO private key %v", err)
	}
	// signatureBytes := signature_provider.NewSignatureNeo(dev, tea, contractOwnerPrivateKey)

	// Following the steps on the developer's side after receiving the signature bytes:
	// Create and initialize client
	// Developer received the signature bytes and can now create the transaction to withdraw funds
	owner := wallet.NewAccountFromPrivateKey(contractOwnerPrivateKey)

	if neo.Deploy {
		h, err := deploy(neoClient, owner)
		if err != nil {
			log.Fatalf("Could not initialize network.")
		} else {
			neo.Contract = h.StringLE()
			opts.Blockchains["neo"] = neo
		}
	}

	// only internal routes, not accessible through caddy server
	router := mux.NewRouter()
	router.HandleFunc("/pay-crypto/{currency}", PaymentCryptoRequestHandler).Methods("POST", "OPTIONS")

	log.Printf("listing on port %v", opts.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
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
		balance := new(big.Float)
		balance.SetInt64(v.NanoTea)
		balance = balance.Mul(balance, opts.Blockchains[cur].Factor)
		balance = balance.Quo(balance, defaultCryptoFactor)
		i, _ := balance.Int(nil)
		amount = append(amount, i)
		addresses = append(addresses, v.Address)
		payoutCrypto = append(payoutCrypto, PayoutCrypto{
			Address:          v.Address,
			NanoTea:          v.NanoTea,
			SmartContractTea: *i,
			Meta:             v.Meta,
		})
	}

	txHash, err := "", nil
	var p *PayoutCryptoResponse
	switch cur {
	case "eth":
		txHash, err = payoutEth(ethClient, addresses, amount)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Unable to payout eth")
			return
		}
		break
	case "neo":
		txHash, err = payoutNEO(addresses, amount)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Unable to payout neo")
			return
		}
		break
	case "xtz":
		txHash, err = payoutNodejsRequest(payoutCrypto, "xtz")
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Unable to payout xtz")
			return
		}
		break
	default:
		log.Printf("Currency isn't supported %v", err)
		w.Header().Set("Content-Type", "application/json")
		writeErr(w, http.StatusBadRequest, "Currency isn't supported %v", err)
	}
	if txHash == "" {
		log.Printf("tx hash is empty contract call failed %v", err)
		writeErr(w, http.StatusBadRequest, "Tx hash is empty contract call failed")
		return
	}
	p = &PayoutCryptoResponse{TxHash: txHash, PayoutCryptos: payoutCrypto}
	log.Printf("%v: Contract call succeeded. Transaction Hash is %v", cur, p.TxHash)
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

func payoutNodejsRequest(payoutCrypto []PayoutCrypto, currency string) (string, error) {
	nodejsClient := http.Client{
		Timeout: 10 * time.Second,
	}
	body, err := json.Marshal(payoutCrypto)
	if err != nil {
		log.Printf("Couldn't decode JSON %v", err)
		return "", err
	}

	fmt.Println("sending request to: " + opts.PayoutNodejsUrl + "/payout/" + currency)
	r, err := nodejsClient.Post(opts.PayoutNodejsUrl+"/payout/"+currency, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Couldn't POST request to the NodeJs %v", err)
		return "", err
	}
	defer r.Body.Close()

	var resp PayoutCryptoResponse
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		log.Printf("Couldnt  %v", err)
		return "", err
	}
	return resp.TxHash, nil
}
