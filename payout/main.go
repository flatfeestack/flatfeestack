package main

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

var (
	client     *ethclient.Client
	contract   *Flatfeestack
	privateKey *ecdsa.PrivateKey
	chainId    *big.Int
)

func main() {
	var err error

	// Ganache NewtorkID / URL
	chainId = big.NewInt(5777)
	client, err = ethclient.Dial("http://127.0.0.1:7545")
	if err != nil {
		log.Fatalf("Could not connect to ethereum client %v", err)
	}

	// if no .key file with the private key (Hex) is present, the method will generate a new keypair and store it in the file.
	// Notice that it's not possible to deploy the contract with a freshly generated keypair as their are costs for gas
	initializeWallet()

	// load existing or deploy new
	initializeContract()

	err = fillContract()
	if err != nil {
		log.Printf("could not fill contract %v", err)
	}
}

func initializeWallet() {
	if _, err := os.Stat(".key"); err == nil {
		var loadErr error
		privateKey, loadErr = crypto.LoadECDSA(".key")
		check(loadErr)

	} else {
		var genErr error
		privateKey, genErr = crypto.GenerateKey()
		check(genErr)

		genErr = crypto.SaveECDSA(".key", privateKey)
		check(genErr)
	}
}

// check if a contract already exists (.contract file with address) or otherwise deploy the contract
func initializeContract() {
	if _, err := os.Stat(".contract"); err == nil {
		file, err := ioutil.ReadFile(".contract")
		check(err)
		contract, err = NewFlatfeestack(common.HexToAddress(string(file)), client)
		check(err)
		log.Printf("contract retrieved from %v", string(file))
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
	f, err := os.Create(".contract")
	check(err)
	defer f.Close()
	_, err = f.WriteString(address.Hex())
	check(err)
	// set global contract instance
	contract = instance

	log.Printf("Created smart contract at (%v)", address.Hex())
}

func fillContract() error {
	ctx := context.Background()
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	nonce, err := client.PendingNonceAt(ctx, auth.From)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	addresses := []common.Address{common.HexToAddress("0xf2C7FFc096CCaA214b91da8240aF3f77C6a0Ee82"), common.HexToAddress("0x3093eEad2860c8B3b6dAeF946373C1E5EC338061")}
	balances := []*big.Int{big.NewInt(1000000000000), big.NewInt(2000000000000000)}
	total := big.NewInt(0)
	for _, value := range balances {
		total.Add(total, value)
	}

	auth.Value = total

	tx, err := contract.Fill(auth, addresses, balances)
	if err != nil {
		log.Fatalf("Failed to create transaction %v", err)
	}
	log.Printf("filled contract TX %v", tx.Hash().Hex())

	return nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
