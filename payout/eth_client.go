package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
)

type ClientETH struct {
	c           *ethclient.Client
	rpc         *rpc.Client
	privateKey  *ecdsa.PrivateKey
	publicKey   *ecdsa.PublicKey
	fromAddress common.Address
	chainId     *big.Int
	contract    *PayoutEth
}

func payoutEth(ethClient *ClientETH, addressValues []string, teas []*big.Int) (string, error) {
	var addresses []common.Address
	for i := range addressValues {
		addresses = append(addresses, common.HexToAddress(addressValues[i]))
	}
	transactor, err := bind.NewKeyedTransactorWithChainID(ethClient.privateKey, ethClient.chainId)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	tx, err := ethClient.contract.BatchPayout(transactor, addresses, teas)
	if err != nil {
		log.Fatalf("Failed transaction: %v", err)
	}
	return tx.Hash().String(), err
}

func getEthClient(ethUrl string, hexPrivateKey string, deploy bool, ethContract string) (*ClientETH, error) {
	rpc, err := rpc.DialContext(context.Background(), ethUrl)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(rpc)

	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.HexToECDSA(hexPrivateKey)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	c := &ClientETH{
		c:           client,
		rpc:         rpc,
		privateKey:  privateKey,
		publicKey:   publicKeyECDSA,
		fromAddress: fromAddress,
	}

	chainId, err := c.c.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}
	c.chainId = chainId

	if deploy {
		log.Printf("Start deploying ETH Contract...")
		c.contract = deployEthContract(c)
	} else {
		c.contract, err = NewPayoutEth(common.HexToAddress(ethContract), c.c)
		if err != nil {
			log.Fatal(err)
		}
	}

	return c, nil
}

func deployEthContract(ethClient *ClientETH) *PayoutEth {
	opts, err := bind.NewKeyedTransactorWithChainID(ethClient.privateKey, ethClient.chainId)
	address, tx, contract, err := DeployPayoutEth(opts, ethClient.c)
	if err != nil {
		log.Fatal(err)
	}
	_, err = bind.WaitDeployed(context.Background(), ethClient.c, tx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("---------------------------------")
	log.Printf("ETH Contract deployed at %v", address)
	fmt.Println("---------------------------------")
	return contract
}
