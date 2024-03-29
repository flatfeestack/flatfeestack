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
		log.Errorf("Unable to initialise dial context: %s", err)
		return
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
			LinkOrCreateDiscussion(*proposalCreatedEvent)
		}
	}
}

func LinkOrCreateDiscussion(event ContractDAOProposalCreated) {
	// if the user selects a discussion in our "Create proposal" mask in the frontend
	// a line like "Original discussion: http://localhost:8080/dao/discussion/21a3c381-4bcf-4f4b-a341-a28365518af1" is added to the discussion
	linkPattern := regexp.MustCompile(`Original discussion\: [a-zA-Z\:\/\.\d]+\/dao\/discussion\/([0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12})$`)
	matches := linkPattern.FindStringSubmatch(event.Description)

	if len(matches) == 0 {
		_, err := createPostForProposal(event)
		if err != nil {
			log.Errorf("Unabe to create post for proposal: %s", err)
		}
	} else {
		// discussion is linked, check if reference is valid.
		updateExistingDiscussion(event, matches)
	}
}

func updateExistingDiscussion(event ContractDAOProposalCreated, matches []string) {
	postUuid, err := uuid.Parse(matches[1])
	if err != nil {
		log.Errorf("Unable to parse UUID from event: %s", err)
	}

	exists, err := database.CheckIfPostExists(postUuid)
	if err != nil {
		log.Errorf("Unable to check if post exists: %s", err)
	}

	if exists {
		post, err := database.GetPostById(postUuid)
		if err != nil {
			log.Errorf("Unable to get post from database: %s", err)
		}

		post, err = database.AddProposalIdToPost(post.Id, event.ProposalId)
		if err != nil {
			log.Errorf("Unable to add proposal id to post: %s", err)
		}

		comment := "This discussion has been linked to proposal " + event.ProposalId.String()
		_, err = database.InsertComment(post.Id, uuid.Nil, comment)
		if err != nil {
			log.Errorf("Unable to add a new comment to existing discussion: %s", err)
		}

	} else {
		_, err = createPostForProposal(event)
		if err != nil {
			log.Errorf("Unabe to create post for proposal: %s", err)
		}
	}
}

func createPostForProposal(event ContractDAOProposalCreated) (*database.DbPost, error) {
	description := fmt.Sprintf(
		`A new proposal has been created without any linked discussion.

Proposal id: %s
Proposer creator: %s
Proposal description: %s`, event.ProposalId, event.Proposer, event.Description)

	post, err := database.InsertPost(
		uuid.Nil,
		fmt.Sprintf("Discussion for proposal by %s", event.Proposer),
		description,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	post, err = database.AddProposalIdToPost(post.Id, event.ProposalId)
	if err != nil {
		log.Errorf("Unable to add proposal id to post: %s", err)
	}

	return post, nil
}
