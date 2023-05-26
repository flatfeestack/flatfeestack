package dao

import (
	"fmt"
	"forum/globals"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func TestLinkOrCreateDiscussion(t *testing.T) {
	t.Run("creates new discussion if original link is missing in description", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.Nil(t, err)
		globals.DB = db
		defer db.Close()

		require.Nil(t, err)
		proposalId := big.NewInt(1234)
		proposer := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
		description := "couple of things mentioned but nothing that relates to original discussion"
		postDescription := fmt.Sprintf(
			`A new proposal has been created without any linked discussion.

Proposer creator: %s
Proposal description: %s`, proposer, description)

		mock.ExpectPrepare("INSERT INTO post")
		mock.ExpectQuery("INSERT INTO post").WithArgs(uuid.Nil, postDescription, fmt.Sprintf("Discussion for proposal %s", proposalId)).WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "open", "updated_at"}).
				AddRow("8bef1c41-7625-482e-8589-25cfb31b14a4", time.Now(), true, nil),
		)

		event := ContractDAOProposalCreated{
			ProposalId:  proposalId,
			Proposer:    common.HexToAddress(proposer),
			Description: description,
		}

		LinkOrCreateDiscussion(event)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("creates new discussion if link is found, but discussion is missing", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.Nil(t, err)
		globals.DB = db
		defer db.Close()

		require.Nil(t, err)
		proposalId := big.NewInt(1234)
		proposer := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
		discussionId := uuid.New()
		description := "Original discussion: http://localhost/dao/discussion/" + discussionId.String()

		postDescription := fmt.Sprintf(
			`A new proposal has been created without any linked discussion.

Proposer creator: %s
Proposal description: %s`, proposer, "some text and "+description)

		mock.ExpectQuery("SELECT EXISTS").WithArgs(discussionId).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow("false"))
		mock.ExpectPrepare("INSERT INTO post")
		mock.ExpectQuery("INSERT INTO post").WithArgs(uuid.Nil, postDescription, fmt.Sprintf("Discussion for proposal %s", proposalId)).WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "open", "updated_at"}).
				AddRow("8bef1c41-7625-482e-8589-25cfb31b14a4", time.Now(), true, nil),
		)

		event := ContractDAOProposalCreated{
			ProposalId:  proposalId,
			Proposer:    common.HexToAddress(proposer),
			Description: "some text and " + description,
		}

		LinkOrCreateDiscussion(event)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
