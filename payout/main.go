package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"net/http"
	"os"
)

var (
	client     *ethclient.Client
	contract   *Flatfeestack
	privateKey *ecdsa.PrivateKey
	chainId    *big.Int
)

func main() {

	if os.Getenv("ENV") == "local" {
		//if run locally get environment file from above docker config file
		err := godotenv.Load("../.env")
		if err != nil {
			log.Fatalf("could not find env file. Please add an .env file if you want to run it without docker.", err)
		}
	}

	var err error

	// Ganache NewtorkID / URL
	chainId = big.NewInt(3)
	client, err = ethclient.Dial("https://ropsten.infura.io/v3/6d6c0e875d6c4becaec0e1b10d5bc3cc")
	if err != nil {
		log.Fatalf("Could not connect to ethereum client %v", err)
	}

	// if no .key file with the private key (Hex) is present, the method will generate a new keypair and store it in the file.
	// Notice that it's not possible to deploy the contract with a freshly generated keypair as their are costs for gas
	initializeWallet()

	// load existing or deploy new
	initializeContract()

	// only internal routes, not accessible through caddy server
	router := mux.NewRouter()
	router.HandleFunc("/pay", PaymentRequestHandler).Methods("POST", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func initializeWallet() {

	log.Printf("initializing wallet")
	key, exists := os.LookupEnv("ETH_PK")
	if exists{
		var loadErr error
		privateKey, loadErr = crypto.HexToECDSA(key)
		check(loadErr)
	} else {
		var genErr error
		privateKey, genErr = crypto.GenerateKey()
		check(genErr)
		genErr = crypto.SaveECDSA("secrets/.key", privateKey)
	}
}

// check if a contract already exists (.contract file with address) or otherwise deploy the contract
func initializeContract() {
	c, exists := os.LookupEnv("ETH_CONTRACT")
	if exists{
		var err error
		contract, err = NewFlatfeestack(common.HexToAddress(c), client)
		check(err)
		log.Printf("contract retrieved from %v", c)
	} else {
		deployContract()
	}
}

func deployContract() {
	ctx := context.Background()
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainId)

	nonce, err := client.PendingNonceAt(ctx, auth.From)
	check(err)

	gasPrice, err := client.SuggestGasPrice(ctx)
	check(err)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	address, _, instance, err := DeployFlatfeestack(auth, client)
	if err != nil {
		log.Fatalf("error deploying contract %v", err)
	}
	f, err := os.Create("secrets/.contract")
	check(err)
	defer f.Close()
	_, err = f.WriteString(address.Hex())
	check(err)
	// set global contract instance
	contract = instance

	log.Printf("Created smart contract at (%v)", address.Hex())
}

func fillContract(payouts []Payout) (common.Hash, error) {
	ctx := context.Background()
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	nonce, err := client.PendingNonceAt(ctx, auth.From)
	if err != nil {
		return common.Hash{}, err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	var addresses []common.Address
	var balances []*big.Int
	total := big.NewInt(0)

	for _, p := range payouts {
		addresses = append(addresses, common.HexToAddress(p.Address))
		balances = append(balances, big.NewInt(p.Amount))
		total.Add(total, big.NewInt(p.Amount))
	}

	auth.Value = total

	tx, err := contract.Fill(auth, addresses, balances)
	if err != nil {
		log.Fatalf("Failed to create transaction %v", err)
	}
	log.Printf("filled contract TX %v", tx.Hash().Hex())

	return tx.Hash(), nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
	log.Printf("received payout request for %v addresses", len(data))

	tx, err := fillContract(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(PayoutResponse{TxHash: tx.String()})
}
