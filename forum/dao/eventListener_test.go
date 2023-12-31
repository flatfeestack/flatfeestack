package dao

import (
	"context"
	"fmt"
	database "forum/db"
	"forum/utils"
	"github.com/ethereum/go-ethereum/common"
	dbLib "github.com/flatfeestack/go-lib/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	container, err := utils.InitDatabase()
	if err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	err = dbLib.DB.Close()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer container.Terminate(ctx)

	os.Exit(code)
}

func createPost(t *testing.T) uuid.UUID {
	author := uuid.New()
	title := "Test Post"
	content := "This is a test post."

	post, err := database.InsertPost(author, title, content)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	return post.Id
}

func TestLinkOrCreateDiscussion(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("creates a new post if original link is missing in description", func(t *testing.T) {

		proposalId := big.NewInt(1234)
		proposer := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
		description := "couple of things mentioned but nothing that relates to original discussion"
		postDescription := fmt.Sprintf(
			`A new proposal has been created without any linked discussion.

Proposal id: %s
Proposer creator: %s
Proposal description: %s`, proposalId, proposer, description)

		event := ContractDAOProposalCreated{
			ProposalId:  proposalId,
			Proposer:    common.HexToAddress(proposer),
			Description: description,
		}

		LinkOrCreateDiscussion(event)

		post, err := database.GetPostByProposalId(proposalId.String())
		assert.NoError(t, err, "Can't get post")
		assert.Equal(t, postDescription, post.Content, "Incorrect post description")

	})

	t.Run("creates new post if link is found, but post is missing", func(t *testing.T) {
		proposalId := big.NewInt(4321)
		proposer := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
		discussionId := uuid.New()
		description := "Original discussion: http://localhost/dao/discussion/" + discussionId.String()

		postDescription := fmt.Sprintf(
			`A new proposal has been created without any linked discussion.

Proposal id: %s
Proposer creator: %s
Proposal description: %s`, proposalId, proposer, "some text and "+description)

		event := ContractDAOProposalCreated{
			ProposalId:  proposalId,
			Proposer:    common.HexToAddress(proposer),
			Description: "some text and " + description,
		}

		LinkOrCreateDiscussion(event)
		post, err := database.GetPostByProposalId(proposalId.String())
		assert.NoError(t, err, "Can't get post")
		assert.Equal(t, postDescription, post.Content, "Incorrect post description")

	})

	t.Run("adds proposal id to existing post", func(t *testing.T) {
		postId := createPost(t)
		proposalId := big.NewInt(5678)
		proposer := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
		description := "Original discussion: http://localhost/dao/discussion/" + postId.String()

		comment := "This discussion has been linked to proposal " + proposalId.String()

		event := ContractDAOProposalCreated{
			ProposalId:  proposalId,
			Proposer:    common.HexToAddress(proposer),
			Description: "some text and " + description,
		}

		LinkOrCreateDiscussion(event)

		post, err := database.GetPostByProposalId(proposalId.String())
		assert.NoError(t, err, "Failed to delete post")

		comments, err := database.GetAllComments(post.Id)
		assert.NoError(t, err, "Failed to get comments")
		assert.Equal(t, comment, comments[0].Content, "Incorrect post description")

	})
}
