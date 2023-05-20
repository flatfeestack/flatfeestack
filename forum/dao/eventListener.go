package dao

import (
	"context"
	"fmt"
	database "forum/db"
	"forum/globals"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
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

func LinkOrCreateDiscussion(event ContractDAOProposalCreated) {
	// if the user selects a discussion in our "Create proposal" mask in the frontend
	// a line like "Original discussion: http://localhost:8080/dao/discussion/21a3c381-4bcf-4f4b-a341-a28365518af1" is added to the discussion
	linkPattern := regexp.MustCompile(`Original discussion\: [a-zA-Z\:\/\.\d]+\/dao\/discussion\/([0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12})$`)
	matches := linkPattern.FindStringSubmatch(event.Description)

	if len(matches) == 0 {
		description := fmt.Sprintf(
			`A new proposal has been created without any linked discussion.

Proposer creator: %s
Proposal description: %s`, event.Proposer, event.Description)

		_, err := database.InsertPost(
			uuid.Nil,
			fmt.Sprintf("Discussion for proposal %s", event.ProposalId),
			description,
		)
		if err != nil {
			log.Errorf("Unable to insert new post: %s", err)
		}
	} else {
		// link to existing discussion
	}
}
