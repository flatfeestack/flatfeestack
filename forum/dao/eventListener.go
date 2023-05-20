package dao

import (
	"context"
	"fmt"
	"forum/globals"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
	"math/big"
)

func RunEventListener() {
	dialContext, err := rpc.DialContext(context.Background(), globals.OPTS.EthWsUrl)
	if err != nil {
		log.Fatalf("Unable to initialise dial context: %s", err)
	}
	ethClient := ethclient.NewClient(dialContext)
	daoContractAddress := common.HexToAddress(globals.OPTS.DaoContractAddress)
	daoContractInstance, err := NewContract(daoContractAddress, ethClient)
	if err != nil {
		log.Fatal(err)
	}

	watchOpts := bind.WatchOpts{Context: context.Background()}
	outputChannel := make(chan *ContractDAOProposalCreated)
	subscription, err := daoContractInstance.WatchDAOProposalCreated(
		&watchOpts, outputChannel, []*big.Int{}, []common.Address{}, []uint8{},
	)
	if err != nil {
		log.Fatalf("Unable to create subscription for proposal events: %s", err)
	}

	log.Printf("Successfully initialised connection to ETH chain!")

	go loop(subscription, outputChannel)
}

func loop(subscription event.Subscription, outputChannel chan *ContractDAOProposalCreated) {
	for {
		select {
		case err := <-subscription.Err():
			log.Fatal(err)
		case proposalCreatedEvent := <-outputChannel:
			fmt.Println(proposalCreatedEvent) // pointer to proposalCreatedEvent log
		}
	}
}
