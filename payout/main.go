package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/callflag"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/manifest"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/nef"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"
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
	NanoTea int64  `json:"nano_tea"`
}

type PayoutCrypto struct {
	Address          string  `json:"address"`
	NanoTea          int64   `json:"nano_tea"`
	SmartContractTea big.Int `json:"smart_contract_tea"`
}

type PayoutCryptoResponse struct {
	TxHash        string         `json:"tx_hash"`
	PayoutCryptos []PayoutCrypto `json:"payout_cryptos"`
}

type Opts struct {
	Port            int
	Env             string
	EthContract     string
	EthPrivateKey   string
	EthUrl          string
	Deploy          bool
	PayoutNodejsUrl string
}

var (
	opts                *Opts
	EthWei              = big.NewFloat(0)
	MicroUsd            = big.NewFloat(0)
	UsdWei              = big.NewFloat(0)
	defaultCryptoFactor = big.NewFloat(0)
	cryptoFactor        = map[string]*big.Float{
		"eth": big.NewFloat(1000000000000000000),
		"neo": big.NewFloat(100000000),
		"xtz": big.NewFloat(1000000),
	}
	ethClient *ClientETH
	debug     bool
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
	flag.StringVar(&o.PayoutNodejsUrl, "payout-nodejs-url", lookupEnv("PAYOUT_NODEJS_URL",
		"http://localhost:9086"), "Payout Nodejs Url")

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
	UsdWei.SetString("1000000000000")           //EthWei/MicroUsd
	defaultCryptoFactor.SetString("1000000000") // Fixed factor for the moment (Nano)

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
		balance := new(big.Float)
		balance.SetInt64(v.NanoTea)
		balance = balance.Mul(balance, cryptoFactor[cur])
		balance = balance.Quo(balance, defaultCryptoFactor)
		i, _ := balance.Int(nil)
		amount = append(amount, i)
		addresses = append(addresses, v.Address)
		payoutCrypto = append(payoutCrypto, PayoutCrypto{
			Address:          v.Address,
			NanoTea:          v.NanoTea,
			SmartContractTea: *i,
		})
	}

	txHash, err := "", nil
	var p *PayoutCryptoResponse
	switch cur {
	case "eth":
		transaction, err := payoutEth(ethClient, addresses, amount)
		if err != nil {
			log.Fatal(err)
		}
		txHash = transaction.Hash().String()
		p = &PayoutCryptoResponse{TxHash: txHash, PayoutCryptos: payoutCrypto}
		break
	case "neo":
		txHash = payoutNEO(addresses, amount)
		p = &PayoutCryptoResponse{TxHash: txHash, PayoutCryptos: payoutCrypto}
		break
	case "xtz":
		txHash = payoutNodejsRequest(payoutCrypto, "xtz")
		p = &PayoutCryptoResponse{TxHash: txHash, PayoutCryptos: payoutCrypto}
		break
	default:
		log.Printf("Currency isn't supported %v", err)
		w.Header().Set("Content-Type", "application/json")
		writeErr(w, http.StatusBadRequest, "Currency isn't supported %v", err)
	}
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

func payoutNodejsRequest(payoutCrypto []PayoutCrypto, currency string) string {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	body, err := json.Marshal(payoutCrypto)
	if err != nil {
		log.Printf("Couldn't decode JSON %v", err)
		return ""
	}

	fmt.Println("sending request to: " + opts.PayoutNodejsUrl + "/payout/" + currency)
	r, err := client.Post(opts.PayoutNodejsUrl+"/payout/"+currency, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Couldn't POST request to the NodeJs %v", err)
		return ""
	}
	defer r.Body.Close()

	var resp PayoutCryptoResponse
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		log.Printf("Couldnt  %v", err)
		return ""
	}
	if resp.TxHash == "" {
		log.Printf("tx hash is empty contract call failed %v", err)
		return ""
	}
	return resp.TxHash
}

func payoutNEO(addressValues []string, teas []*big.Int) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Could not load env file.")
	}
	contractOwnerPrivateKey, _ := keys.NewPrivateKeyFromWIF(os.Getenv("CONTRACT_OWNER_WIF"))
	// signatureBytes := signature_provider.NewSignatureNeo(dev, tea, contractOwnerPrivateKey)

	// Following the steps on the developer's side after receiving the signature bytes:
	// Create and initialize client
	// Developer received the signature bytes and can now create the transaction to withdraw funds
	owner := wallet.NewAccountFromPrivateKey(contractOwnerPrivateKey)
	c, _ := client.New(context.TODO(), "http://seed1t4.neo.org:20332", client.Options{})

	err = c.Init()
	if err != nil {
		log.Fatalf("Could not initialize network.")
	}
	// Contract hash of deployed contract on testnet

	// var payoutNeoHash, _ = util.Uint160DecodeStringLE("76856f89cbb61e9fce7f6c1a76d79a1a3ee69ef4")  //Own 2
	var payoutNeoHash, _ = util.Uint160DecodeStringLE("80b4f117c6c882f0dd7c58cc6f6112e64b0f37b7") //Own
	// var payoutNeoHash, _ = util.Uint160DecodeStringLE("38f6215e40769c27fee742d7af1a9062e962158f") // Michael

	if false {
		h, err := deploy(c, owner)
		if err != nil {
			log.Fatalf("Could not initialize network.")
		} else {
			payoutNeoHash = h
		}
	}

	h := CreateBatchPayoutTx(c, payoutNeoHash, owner, 0, owner.PrivateKey().GetScriptHash(), addressValues, teas)
	return h
}

//CreateWithdrawTx creates a transaction to withdraw funds for the provided dev, tea and the signature bytes.
func CreateBatchPayoutTx(c *client.Client, contractHash util.Uint160, acc *wallet.Account, additionalNetworkFee int64,
	dev util.Uint160, addressValues []string, teas []*big.Int) string {
	return packParams(c, contractHash, acc, addressValues, teas)
}

//SignTransaction Signs the transaction with the provided signer account.
func SignTransaction(c *client.Client, signer *wallet.Account, transaction *transaction.Transaction) error {
	return signer.SignTx(c.GetNetwork(), transaction)
}

func readNEFFile(filename string) (*nef.File, []byte, error) {
	if len(filename) == 0 {
		return nil, nil, errors.New("no nef file was provided")
	}

	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	nefFile, err := nef.FileFromBytes(f)
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse NEF file: %w", err)
	}

	return &nefFile, f, nil
}

func readManifest(filename string) (*manifest.Manifest, []byte, error) {
	if len(filename) == 0 {
		return nil, nil, errors.New("no manifest file was provided")
	}

	manifestBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	m := new(manifest.Manifest)
	err = json.Unmarshal(manifestBytes, m)
	if err != nil {
		return nil, nil, err
	}
	return m, manifestBytes, nil
}

func deploy(c *client.Client, acc *wallet.Account) (util.Uint160, error) {
	nativeManagementContractHash, err := c.GetNativeContractHash(nativenames.Management)
	if err != nil {
		log.Fatalf("Couldn't get native management contract hash")
	}
	ne, nefB, err := readNEFFile("./PayoutNeo.nef")
	_, mfB, err := readManifest("./PayoutNeo.manifest.json")
	sender := acc.PrivateKey().GetScriptHash()
	pk := acc.PrivateKey().PublicKey().Bytes()
	appCallParams := []smartcontract.Parameter{
		{
			Type:  smartcontract.ByteArrayType,
			Value: nefB,
		},
		{
			Type:  smartcontract.ByteArrayType,
			Value: mfB,
		},
		{
			Type:  smartcontract.PublicKeyType,
			Value: pk,
		},
	}

	contractHash := state.CreateContractHash(sender, ne.Checksum, "PayoutNeo")
	signer := transaction.Signer{
		Account: sender,
		Scopes:  transaction.Global,
		// CustomContracts do not work with neo-go, if that scope is used for the sender when using the method CreateTxFromScript.
		// Same holds for CustomGroups...
		//Scopes:           transaction.CustomContracts,
		//AllowedContracts: []util.Uint160{contractHash},
	}
	resp, _ := c.InvokeFunction(nativeManagementContractHash, "deploy", appCallParams, []transaction.Signer{signer})
	tx, err := c.CreateTxFromScript(resp.Script, acc, -1, 0, []client.SignerAccount{{Signer: signer}})
	if err != nil {
		log.Fatalf(err.Error())
	}
	txHash, err := c.SignAndPushTx(tx, acc, nil)
	if err != nil {
		fmt.Errorf("failed to sign and push transaction: %w", err)
	}
	fmt.Println(txHash.StringLE())
	return contractHash, err
}

func packParams(c *client.Client, payoutNeoHash util.Uint160, acc *wallet.Account, addressValues []string, teas []*big.Int) string {
	var devP []interface{}
	for _, v := range addressValues {
		add, _ := address.StringToUint160(v)
		devP = append(devP, add)
	}
	var teaP []interface{}
	for _, v := range teas {
		teaP = append(teaP, v)
	}

	w := io.NewBufBinWriter()
	emit.AppCall(w.BinWriter, payoutNeoHash, "batchPayout", callflag.All, devP, teaP)
	script := w.Bytes()
	log.Printf(hex.EncodeToString(script))
	tx, err := c.CreateTxFromScript(script, acc, -1, 0, []client.SignerAccount{{
		Signer: transaction.Signer{
			Account: acc.PrivateKey().GetScriptHash(),
			Scopes:  transaction.CalledByEntry,
		},
	}})
	if err != nil {
		log.Fatalf(err.Error())
	}
	acc.SignTx(c.GetNetwork(), tx)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hash, err := c.SendRawTransaction(tx)
	if err != nil {
		fmt.Errorf("send raw transaction err: %v", err)
	}
	return hash.StringLE()
}
