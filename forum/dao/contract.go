// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package dao

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"Empty\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"newBylawsUrl\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"newBylawsHash\",\"type\":\"string\"}],\"name\":\"BylawsChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"targets\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"signatures\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"bytes[]\",\"name\":\"calldatas\",\"type\":\"bytes[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startBlock\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endBlock\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"enumGovernorUpgradeable.ProposalCategory\",\"name\":\"category\",\"type\":\"uint8\"}],\"name\":\"DAOProposalCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"targets\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"signatures\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"bytes[]\",\"name\":\"calldatas\",\"type\":\"bytes[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startBlock\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endBlock\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"name\":\"ExtraOrdinaryAssemblyRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeslot\",\"type\":\"uint256\"}],\"name\":\"NewTimeslotSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"}],\"name\":\"ProposalCanceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"targets\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"signatures\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"bytes[]\",\"name\":\"calldatas\",\"type\":\"bytes[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startBlock\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endBlock\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"name\":\"ProposalCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"}],\"name\":\"ProposalExecuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"eta\",\"type\":\"uint256\"}],\"name\":\"ProposalQueued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"oldTime\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"newTime\",\"type\":\"uint64\"}],\"name\":\"ProposalVotingTimeChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldQuorumNumerator\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newQuorumNumerator\",\"type\":\"uint256\"}],\"name\":\"QuorumNumeratorUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oldTimelock\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newTimelock\",\"type\":\"address\"}],\"name\":\"TimelockChange\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"support\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"weight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"VoteCast\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"support\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"weight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"}],\"name\":\"VoteCastWithParams\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"VotingSlotCancelled\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BALLOT_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"COUNTING_MODE\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"EXTENDED_BALLOT_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"associationDissolutionQuorumNominator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bylawsHash\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bylawsUrl\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"cancelVotingSlot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"support\",\"type\":\"uint8\"}],\"name\":\"castVote\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"support\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"castVoteBySig\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"support\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"castVoteWithReason\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"support\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"}],\"name\":\"castVoteWithReasonAndParams\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"support\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"castVoteWithReasonAndParamsBySig\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"daoActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dissolveDAO\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"targets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes[]\",\"name\":\"calldatas\",\"type\":\"bytes[]\"},{\"internalType\":\"bytes32\",\"name\":\"descriptionHash\",\"type\":\"bytes32\"}],\"name\":\"execute\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"extraOrdinaryAssemblyProposals\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"extraOrdinaryAssemblyVotingPeriod\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"extraordinaryVoteQuorumNominator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getExtraOrdinaryProposalsLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMinDelay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotNumber\",\"type\":\"uint256\"}],\"name\":\"getNumberOfProposalsInVotingSlot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSlotsLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"getVotes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"}],\"name\":\"getVotesWithParams\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasVoted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"targets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes[]\",\"name\":\"calldatas\",\"type\":\"bytes[]\"},{\"internalType\":\"bytes32\",\"name\":\"descriptionHash\",\"type\":\"bytes32\"}],\"name\":\"hashProposal\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractMembership\",\"name\":\"_membership\",\"type\":\"address\"},{\"internalType\":\"contractTimelockControllerUpgradeable\",\"name\":\"_timelock\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"bylawsHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"bylawsUrl\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"}],\"name\":\"proposalDeadline\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"}],\"name\":\"proposalEta\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"}],\"name\":\"proposalSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proposalThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"}],\"name\":\"proposalVotes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"againstVotes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"forVotes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"abstainVotes\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"targets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes[]\",\"name\":\"calldatas\",\"type\":\"bytes[]\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"name\":\"propose\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"targets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes[]\",\"name\":\"calldatas\",\"type\":\"bytes[]\"},{\"internalType\":\"bytes32\",\"name\":\"descriptionHash\",\"type\":\"bytes32\"}],\"name\":\"queue\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"quorum\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"quorumDenominator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"quorumNumerator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"quorumNumerator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"relay\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"setAssociationDissolutionQuorumNominator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"newExtraOrdinaryAssemblyVotingPeriod\",\"type\":\"uint64\"}],\"name\":\"setExtraOrdinaryAssemblyVotingPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"setExtraordinaryVoteQuorumNominator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"newBylawsHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"newBylawsUrl\",\"type\":\"string\"}],\"name\":\"setNewBylaws\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newSlotCloseTime\",\"type\":\"uint256\"}],\"name\":\"setSlotCloseTime\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"setVotingSlot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"newVotingSlotAnnouncementPeriod\",\"type\":\"uint64\"}],\"name\":\"setVotingSlotAnnouncementPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"slotCloseTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slots\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposalId\",\"type\":\"uint256\"}],\"name\":\"state\",\"outputs\":[{\"internalType\":\"enumIGovernorUpgradeable.ProposalState\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timelock\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIVotesUpgradeable\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newQuorumNumerator\",\"type\":\"uint256\"}],\"name\":\"updateQuorumNumerator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractTimelockControllerUpgradeable\",\"name\":\"newTimelock\",\"type\":\"address\"}],\"name\":\"updateTimelock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"votingDelay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"votingPeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"votingSlotAnnouncementPeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"votingSlots\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// BALLOTTYPEHASH is a free data retrieval call binding the contract method 0xdeaaa7cc.
//
// Solidity: function BALLOT_TYPEHASH() view returns(bytes32)
func (_Contract *ContractCaller) BALLOTTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "BALLOT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BALLOTTYPEHASH is a free data retrieval call binding the contract method 0xdeaaa7cc.
//
// Solidity: function BALLOT_TYPEHASH() view returns(bytes32)
func (_Contract *ContractSession) BALLOTTYPEHASH() ([32]byte, error) {
	return _Contract.Contract.BALLOTTYPEHASH(&_Contract.CallOpts)
}

// BALLOTTYPEHASH is a free data retrieval call binding the contract method 0xdeaaa7cc.
//
// Solidity: function BALLOT_TYPEHASH() view returns(bytes32)
func (_Contract *ContractCallerSession) BALLOTTYPEHASH() ([32]byte, error) {
	return _Contract.Contract.BALLOTTYPEHASH(&_Contract.CallOpts)
}

// COUNTINGMODE is a free data retrieval call binding the contract method 0xdd4e2ba5.
//
// Solidity: function COUNTING_MODE() pure returns(string)
func (_Contract *ContractCaller) COUNTINGMODE(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "COUNTING_MODE")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// COUNTINGMODE is a free data retrieval call binding the contract method 0xdd4e2ba5.
//
// Solidity: function COUNTING_MODE() pure returns(string)
func (_Contract *ContractSession) COUNTINGMODE() (string, error) {
	return _Contract.Contract.COUNTINGMODE(&_Contract.CallOpts)
}

// COUNTINGMODE is a free data retrieval call binding the contract method 0xdd4e2ba5.
//
// Solidity: function COUNTING_MODE() pure returns(string)
func (_Contract *ContractCallerSession) COUNTINGMODE() (string, error) {
	return _Contract.Contract.COUNTINGMODE(&_Contract.CallOpts)
}

// EXTENDEDBALLOTTYPEHASH is a free data retrieval call binding the contract method 0x2fe3e261.
//
// Solidity: function EXTENDED_BALLOT_TYPEHASH() view returns(bytes32)
func (_Contract *ContractCaller) EXTENDEDBALLOTTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "EXTENDED_BALLOT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// EXTENDEDBALLOTTYPEHASH is a free data retrieval call binding the contract method 0x2fe3e261.
//
// Solidity: function EXTENDED_BALLOT_TYPEHASH() view returns(bytes32)
func (_Contract *ContractSession) EXTENDEDBALLOTTYPEHASH() ([32]byte, error) {
	return _Contract.Contract.EXTENDEDBALLOTTYPEHASH(&_Contract.CallOpts)
}

// EXTENDEDBALLOTTYPEHASH is a free data retrieval call binding the contract method 0x2fe3e261.
//
// Solidity: function EXTENDED_BALLOT_TYPEHASH() view returns(bytes32)
func (_Contract *ContractCallerSession) EXTENDEDBALLOTTYPEHASH() ([32]byte, error) {
	return _Contract.Contract.EXTENDEDBALLOTTYPEHASH(&_Contract.CallOpts)
}

// AssociationDissolutionQuorumNominator is a free data retrieval call binding the contract method 0xdf748749.
//
// Solidity: function associationDissolutionQuorumNominator() view returns(uint256)
func (_Contract *ContractCaller) AssociationDissolutionQuorumNominator(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "associationDissolutionQuorumNominator")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AssociationDissolutionQuorumNominator is a free data retrieval call binding the contract method 0xdf748749.
//
// Solidity: function associationDissolutionQuorumNominator() view returns(uint256)
func (_Contract *ContractSession) AssociationDissolutionQuorumNominator() (*big.Int, error) {
	return _Contract.Contract.AssociationDissolutionQuorumNominator(&_Contract.CallOpts)
}

// AssociationDissolutionQuorumNominator is a free data retrieval call binding the contract method 0xdf748749.
//
// Solidity: function associationDissolutionQuorumNominator() view returns(uint256)
func (_Contract *ContractCallerSession) AssociationDissolutionQuorumNominator() (*big.Int, error) {
	return _Contract.Contract.AssociationDissolutionQuorumNominator(&_Contract.CallOpts)
}

// BylawsHash is a free data retrieval call binding the contract method 0xaeb275c8.
//
// Solidity: function bylawsHash() view returns(string)
func (_Contract *ContractCaller) BylawsHash(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "bylawsHash")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// BylawsHash is a free data retrieval call binding the contract method 0xaeb275c8.
//
// Solidity: function bylawsHash() view returns(string)
func (_Contract *ContractSession) BylawsHash() (string, error) {
	return _Contract.Contract.BylawsHash(&_Contract.CallOpts)
}

// BylawsHash is a free data retrieval call binding the contract method 0xaeb275c8.
//
// Solidity: function bylawsHash() view returns(string)
func (_Contract *ContractCallerSession) BylawsHash() (string, error) {
	return _Contract.Contract.BylawsHash(&_Contract.CallOpts)
}

// BylawsUrl is a free data retrieval call binding the contract method 0x7b9bf794.
//
// Solidity: function bylawsUrl() view returns(string)
func (_Contract *ContractCaller) BylawsUrl(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "bylawsUrl")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// BylawsUrl is a free data retrieval call binding the contract method 0x7b9bf794.
//
// Solidity: function bylawsUrl() view returns(string)
func (_Contract *ContractSession) BylawsUrl() (string, error) {
	return _Contract.Contract.BylawsUrl(&_Contract.CallOpts)
}

// BylawsUrl is a free data retrieval call binding the contract method 0x7b9bf794.
//
// Solidity: function bylawsUrl() view returns(string)
func (_Contract *ContractCallerSession) BylawsUrl() (string, error) {
	return _Contract.Contract.BylawsUrl(&_Contract.CallOpts)
}

// DaoActive is a free data retrieval call binding the contract method 0x2ee3e71d.
//
// Solidity: function daoActive() view returns(bool)
func (_Contract *ContractCaller) DaoActive(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "daoActive")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// DaoActive is a free data retrieval call binding the contract method 0x2ee3e71d.
//
// Solidity: function daoActive() view returns(bool)
func (_Contract *ContractSession) DaoActive() (bool, error) {
	return _Contract.Contract.DaoActive(&_Contract.CallOpts)
}

// DaoActive is a free data retrieval call binding the contract method 0x2ee3e71d.
//
// Solidity: function daoActive() view returns(bool)
func (_Contract *ContractCallerSession) DaoActive() (bool, error) {
	return _Contract.Contract.DaoActive(&_Contract.CallOpts)
}

// ExtraOrdinaryAssemblyProposals is a free data retrieval call binding the contract method 0xacabafb6.
//
// Solidity: function extraOrdinaryAssemblyProposals(uint256 ) view returns(uint256)
func (_Contract *ContractCaller) ExtraOrdinaryAssemblyProposals(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "extraOrdinaryAssemblyProposals", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExtraOrdinaryAssemblyProposals is a free data retrieval call binding the contract method 0xacabafb6.
//
// Solidity: function extraOrdinaryAssemblyProposals(uint256 ) view returns(uint256)
func (_Contract *ContractSession) ExtraOrdinaryAssemblyProposals(arg0 *big.Int) (*big.Int, error) {
	return _Contract.Contract.ExtraOrdinaryAssemblyProposals(&_Contract.CallOpts, arg0)
}

// ExtraOrdinaryAssemblyProposals is a free data retrieval call binding the contract method 0xacabafb6.
//
// Solidity: function extraOrdinaryAssemblyProposals(uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) ExtraOrdinaryAssemblyProposals(arg0 *big.Int) (*big.Int, error) {
	return _Contract.Contract.ExtraOrdinaryAssemblyProposals(&_Contract.CallOpts, arg0)
}

// ExtraOrdinaryAssemblyVotingPeriod is a free data retrieval call binding the contract method 0xa3e857da.
//
// Solidity: function extraOrdinaryAssemblyVotingPeriod() view returns(uint64)
func (_Contract *ContractCaller) ExtraOrdinaryAssemblyVotingPeriod(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "extraOrdinaryAssemblyVotingPeriod")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// ExtraOrdinaryAssemblyVotingPeriod is a free data retrieval call binding the contract method 0xa3e857da.
//
// Solidity: function extraOrdinaryAssemblyVotingPeriod() view returns(uint64)
func (_Contract *ContractSession) ExtraOrdinaryAssemblyVotingPeriod() (uint64, error) {
	return _Contract.Contract.ExtraOrdinaryAssemblyVotingPeriod(&_Contract.CallOpts)
}

// ExtraOrdinaryAssemblyVotingPeriod is a free data retrieval call binding the contract method 0xa3e857da.
//
// Solidity: function extraOrdinaryAssemblyVotingPeriod() view returns(uint64)
func (_Contract *ContractCallerSession) ExtraOrdinaryAssemblyVotingPeriod() (uint64, error) {
	return _Contract.Contract.ExtraOrdinaryAssemblyVotingPeriod(&_Contract.CallOpts)
}

// ExtraordinaryVoteQuorumNominator is a free data retrieval call binding the contract method 0xd87b33ec.
//
// Solidity: function extraordinaryVoteQuorumNominator() view returns(uint256)
func (_Contract *ContractCaller) ExtraordinaryVoteQuorumNominator(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "extraordinaryVoteQuorumNominator")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExtraordinaryVoteQuorumNominator is a free data retrieval call binding the contract method 0xd87b33ec.
//
// Solidity: function extraordinaryVoteQuorumNominator() view returns(uint256)
func (_Contract *ContractSession) ExtraordinaryVoteQuorumNominator() (*big.Int, error) {
	return _Contract.Contract.ExtraordinaryVoteQuorumNominator(&_Contract.CallOpts)
}

// ExtraordinaryVoteQuorumNominator is a free data retrieval call binding the contract method 0xd87b33ec.
//
// Solidity: function extraordinaryVoteQuorumNominator() view returns(uint256)
func (_Contract *ContractCallerSession) ExtraordinaryVoteQuorumNominator() (*big.Int, error) {
	return _Contract.Contract.ExtraordinaryVoteQuorumNominator(&_Contract.CallOpts)
}

// GetExtraOrdinaryProposalsLength is a free data retrieval call binding the contract method 0xaa10701a.
//
// Solidity: function getExtraOrdinaryProposalsLength() view returns(uint256)
func (_Contract *ContractCaller) GetExtraOrdinaryProposalsLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getExtraOrdinaryProposalsLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetExtraOrdinaryProposalsLength is a free data retrieval call binding the contract method 0xaa10701a.
//
// Solidity: function getExtraOrdinaryProposalsLength() view returns(uint256)
func (_Contract *ContractSession) GetExtraOrdinaryProposalsLength() (*big.Int, error) {
	return _Contract.Contract.GetExtraOrdinaryProposalsLength(&_Contract.CallOpts)
}

// GetExtraOrdinaryProposalsLength is a free data retrieval call binding the contract method 0xaa10701a.
//
// Solidity: function getExtraOrdinaryProposalsLength() view returns(uint256)
func (_Contract *ContractCallerSession) GetExtraOrdinaryProposalsLength() (*big.Int, error) {
	return _Contract.Contract.GetExtraOrdinaryProposalsLength(&_Contract.CallOpts)
}

// GetMinDelay is a free data retrieval call binding the contract method 0xf27a0c92.
//
// Solidity: function getMinDelay() view returns(uint256 duration)
func (_Contract *ContractCaller) GetMinDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getMinDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinDelay is a free data retrieval call binding the contract method 0xf27a0c92.
//
// Solidity: function getMinDelay() view returns(uint256 duration)
func (_Contract *ContractSession) GetMinDelay() (*big.Int, error) {
	return _Contract.Contract.GetMinDelay(&_Contract.CallOpts)
}

// GetMinDelay is a free data retrieval call binding the contract method 0xf27a0c92.
//
// Solidity: function getMinDelay() view returns(uint256 duration)
func (_Contract *ContractCallerSession) GetMinDelay() (*big.Int, error) {
	return _Contract.Contract.GetMinDelay(&_Contract.CallOpts)
}

// GetNumberOfProposalsInVotingSlot is a free data retrieval call binding the contract method 0x0ff3dc01.
//
// Solidity: function getNumberOfProposalsInVotingSlot(uint256 slotNumber) view returns(uint256)
func (_Contract *ContractCaller) GetNumberOfProposalsInVotingSlot(opts *bind.CallOpts, slotNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getNumberOfProposalsInVotingSlot", slotNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNumberOfProposalsInVotingSlot is a free data retrieval call binding the contract method 0x0ff3dc01.
//
// Solidity: function getNumberOfProposalsInVotingSlot(uint256 slotNumber) view returns(uint256)
func (_Contract *ContractSession) GetNumberOfProposalsInVotingSlot(slotNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetNumberOfProposalsInVotingSlot(&_Contract.CallOpts, slotNumber)
}

// GetNumberOfProposalsInVotingSlot is a free data retrieval call binding the contract method 0x0ff3dc01.
//
// Solidity: function getNumberOfProposalsInVotingSlot(uint256 slotNumber) view returns(uint256)
func (_Contract *ContractCallerSession) GetNumberOfProposalsInVotingSlot(slotNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetNumberOfProposalsInVotingSlot(&_Contract.CallOpts, slotNumber)
}

// GetSlotsLength is a free data retrieval call binding the contract method 0x2e559b99.
//
// Solidity: function getSlotsLength() view returns(uint256)
func (_Contract *ContractCaller) GetSlotsLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getSlotsLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlotsLength is a free data retrieval call binding the contract method 0x2e559b99.
//
// Solidity: function getSlotsLength() view returns(uint256)
func (_Contract *ContractSession) GetSlotsLength() (*big.Int, error) {
	return _Contract.Contract.GetSlotsLength(&_Contract.CallOpts)
}

// GetSlotsLength is a free data retrieval call binding the contract method 0x2e559b99.
//
// Solidity: function getSlotsLength() view returns(uint256)
func (_Contract *ContractCallerSession) GetSlotsLength() (*big.Int, error) {
	return _Contract.Contract.GetSlotsLength(&_Contract.CallOpts)
}

// GetVotes is a free data retrieval call binding the contract method 0xeb9019d4.
//
// Solidity: function getVotes(address account, uint256 blockNumber) view returns(uint256)
func (_Contract *ContractCaller) GetVotes(opts *bind.CallOpts, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getVotes", account, blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetVotes is a free data retrieval call binding the contract method 0xeb9019d4.
//
// Solidity: function getVotes(address account, uint256 blockNumber) view returns(uint256)
func (_Contract *ContractSession) GetVotes(account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetVotes(&_Contract.CallOpts, account, blockNumber)
}

// GetVotes is a free data retrieval call binding the contract method 0xeb9019d4.
//
// Solidity: function getVotes(address account, uint256 blockNumber) view returns(uint256)
func (_Contract *ContractCallerSession) GetVotes(account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetVotes(&_Contract.CallOpts, account, blockNumber)
}

// GetVotesWithParams is a free data retrieval call binding the contract method 0x9a802a6d.
//
// Solidity: function getVotesWithParams(address account, uint256 blockNumber, bytes params) view returns(uint256)
func (_Contract *ContractCaller) GetVotesWithParams(opts *bind.CallOpts, account common.Address, blockNumber *big.Int, params []byte) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getVotesWithParams", account, blockNumber, params)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetVotesWithParams is a free data retrieval call binding the contract method 0x9a802a6d.
//
// Solidity: function getVotesWithParams(address account, uint256 blockNumber, bytes params) view returns(uint256)
func (_Contract *ContractSession) GetVotesWithParams(account common.Address, blockNumber *big.Int, params []byte) (*big.Int, error) {
	return _Contract.Contract.GetVotesWithParams(&_Contract.CallOpts, account, blockNumber, params)
}

// GetVotesWithParams is a free data retrieval call binding the contract method 0x9a802a6d.
//
// Solidity: function getVotesWithParams(address account, uint256 blockNumber, bytes params) view returns(uint256)
func (_Contract *ContractCallerSession) GetVotesWithParams(account common.Address, blockNumber *big.Int, params []byte) (*big.Int, error) {
	return _Contract.Contract.GetVotesWithParams(&_Contract.CallOpts, account, blockNumber, params)
}

// HasVoted is a free data retrieval call binding the contract method 0x43859632.
//
// Solidity: function hasVoted(uint256 proposalId, address account) view returns(bool)
func (_Contract *ContractCaller) HasVoted(opts *bind.CallOpts, proposalId *big.Int, account common.Address) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "hasVoted", proposalId, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasVoted is a free data retrieval call binding the contract method 0x43859632.
//
// Solidity: function hasVoted(uint256 proposalId, address account) view returns(bool)
func (_Contract *ContractSession) HasVoted(proposalId *big.Int, account common.Address) (bool, error) {
	return _Contract.Contract.HasVoted(&_Contract.CallOpts, proposalId, account)
}

// HasVoted is a free data retrieval call binding the contract method 0x43859632.
//
// Solidity: function hasVoted(uint256 proposalId, address account) view returns(bool)
func (_Contract *ContractCallerSession) HasVoted(proposalId *big.Int, account common.Address) (bool, error) {
	return _Contract.Contract.HasVoted(&_Contract.CallOpts, proposalId, account)
}

// HashProposal is a free data retrieval call binding the contract method 0xc59057e4.
//
// Solidity: function hashProposal(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) pure returns(uint256)
func (_Contract *ContractCaller) HashProposal(opts *bind.CallOpts, targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "hashProposal", targets, values, calldatas, descriptionHash)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// HashProposal is a free data retrieval call binding the contract method 0xc59057e4.
//
// Solidity: function hashProposal(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) pure returns(uint256)
func (_Contract *ContractSession) HashProposal(targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*big.Int, error) {
	return _Contract.Contract.HashProposal(&_Contract.CallOpts, targets, values, calldatas, descriptionHash)
}

// HashProposal is a free data retrieval call binding the contract method 0xc59057e4.
//
// Solidity: function hashProposal(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) pure returns(uint256)
func (_Contract *ContractCallerSession) HashProposal(targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*big.Int, error) {
	return _Contract.Contract.HashProposal(&_Contract.CallOpts, targets, values, calldatas, descriptionHash)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Contract *ContractCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Contract *ContractSession) Name() (string, error) {
	return _Contract.Contract.Name(&_Contract.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Contract *ContractCallerSession) Name() (string, error) {
	return _Contract.Contract.Name(&_Contract.CallOpts)
}

// ProposalDeadline is a free data retrieval call binding the contract method 0xc01f9e37.
//
// Solidity: function proposalDeadline(uint256 proposalId) view returns(uint256)
func (_Contract *ContractCaller) ProposalDeadline(opts *bind.CallOpts, proposalId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "proposalDeadline", proposalId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalDeadline is a free data retrieval call binding the contract method 0xc01f9e37.
//
// Solidity: function proposalDeadline(uint256 proposalId) view returns(uint256)
func (_Contract *ContractSession) ProposalDeadline(proposalId *big.Int) (*big.Int, error) {
	return _Contract.Contract.ProposalDeadline(&_Contract.CallOpts, proposalId)
}

// ProposalDeadline is a free data retrieval call binding the contract method 0xc01f9e37.
//
// Solidity: function proposalDeadline(uint256 proposalId) view returns(uint256)
func (_Contract *ContractCallerSession) ProposalDeadline(proposalId *big.Int) (*big.Int, error) {
	return _Contract.Contract.ProposalDeadline(&_Contract.CallOpts, proposalId)
}

// ProposalEta is a free data retrieval call binding the contract method 0xab58fb8e.
//
// Solidity: function proposalEta(uint256 proposalId) view returns(uint256)
func (_Contract *ContractCaller) ProposalEta(opts *bind.CallOpts, proposalId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "proposalEta", proposalId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalEta is a free data retrieval call binding the contract method 0xab58fb8e.
//
// Solidity: function proposalEta(uint256 proposalId) view returns(uint256)
func (_Contract *ContractSession) ProposalEta(proposalId *big.Int) (*big.Int, error) {
	return _Contract.Contract.ProposalEta(&_Contract.CallOpts, proposalId)
}

// ProposalEta is a free data retrieval call binding the contract method 0xab58fb8e.
//
// Solidity: function proposalEta(uint256 proposalId) view returns(uint256)
func (_Contract *ContractCallerSession) ProposalEta(proposalId *big.Int) (*big.Int, error) {
	return _Contract.Contract.ProposalEta(&_Contract.CallOpts, proposalId)
}

// ProposalSnapshot is a free data retrieval call binding the contract method 0x2d63f693.
//
// Solidity: function proposalSnapshot(uint256 proposalId) view returns(uint256)
func (_Contract *ContractCaller) ProposalSnapshot(opts *bind.CallOpts, proposalId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "proposalSnapshot", proposalId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalSnapshot is a free data retrieval call binding the contract method 0x2d63f693.
//
// Solidity: function proposalSnapshot(uint256 proposalId) view returns(uint256)
func (_Contract *ContractSession) ProposalSnapshot(proposalId *big.Int) (*big.Int, error) {
	return _Contract.Contract.ProposalSnapshot(&_Contract.CallOpts, proposalId)
}

// ProposalSnapshot is a free data retrieval call binding the contract method 0x2d63f693.
//
// Solidity: function proposalSnapshot(uint256 proposalId) view returns(uint256)
func (_Contract *ContractCallerSession) ProposalSnapshot(proposalId *big.Int) (*big.Int, error) {
	return _Contract.Contract.ProposalSnapshot(&_Contract.CallOpts, proposalId)
}

// ProposalThreshold is a free data retrieval call binding the contract method 0xb58131b0.
//
// Solidity: function proposalThreshold() pure returns(uint256)
func (_Contract *ContractCaller) ProposalThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "proposalThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalThreshold is a free data retrieval call binding the contract method 0xb58131b0.
//
// Solidity: function proposalThreshold() pure returns(uint256)
func (_Contract *ContractSession) ProposalThreshold() (*big.Int, error) {
	return _Contract.Contract.ProposalThreshold(&_Contract.CallOpts)
}

// ProposalThreshold is a free data retrieval call binding the contract method 0xb58131b0.
//
// Solidity: function proposalThreshold() pure returns(uint256)
func (_Contract *ContractCallerSession) ProposalThreshold() (*big.Int, error) {
	return _Contract.Contract.ProposalThreshold(&_Contract.CallOpts)
}

// ProposalVotes is a free data retrieval call binding the contract method 0x544ffc9c.
//
// Solidity: function proposalVotes(uint256 proposalId) view returns(uint256 againstVotes, uint256 forVotes, uint256 abstainVotes)
func (_Contract *ContractCaller) ProposalVotes(opts *bind.CallOpts, proposalId *big.Int) (struct {
	AgainstVotes *big.Int
	ForVotes     *big.Int
	AbstainVotes *big.Int
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "proposalVotes", proposalId)

	outstruct := new(struct {
		AgainstVotes *big.Int
		ForVotes     *big.Int
		AbstainVotes *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.AgainstVotes = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ForVotes = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.AbstainVotes = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// ProposalVotes is a free data retrieval call binding the contract method 0x544ffc9c.
//
// Solidity: function proposalVotes(uint256 proposalId) view returns(uint256 againstVotes, uint256 forVotes, uint256 abstainVotes)
func (_Contract *ContractSession) ProposalVotes(proposalId *big.Int) (struct {
	AgainstVotes *big.Int
	ForVotes     *big.Int
	AbstainVotes *big.Int
}, error) {
	return _Contract.Contract.ProposalVotes(&_Contract.CallOpts, proposalId)
}

// ProposalVotes is a free data retrieval call binding the contract method 0x544ffc9c.
//
// Solidity: function proposalVotes(uint256 proposalId) view returns(uint256 againstVotes, uint256 forVotes, uint256 abstainVotes)
func (_Contract *ContractCallerSession) ProposalVotes(proposalId *big.Int) (struct {
	AgainstVotes *big.Int
	ForVotes     *big.Int
	AbstainVotes *big.Int
}, error) {
	return _Contract.Contract.ProposalVotes(&_Contract.CallOpts, proposalId)
}

// Quorum is a free data retrieval call binding the contract method 0xf8ce560a.
//
// Solidity: function quorum(uint256 blockNumber) view returns(uint256)
func (_Contract *ContractCaller) Quorum(opts *bind.CallOpts, blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "quorum", blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Quorum is a free data retrieval call binding the contract method 0xf8ce560a.
//
// Solidity: function quorum(uint256 blockNumber) view returns(uint256)
func (_Contract *ContractSession) Quorum(blockNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.Quorum(&_Contract.CallOpts, blockNumber)
}

// Quorum is a free data retrieval call binding the contract method 0xf8ce560a.
//
// Solidity: function quorum(uint256 blockNumber) view returns(uint256)
func (_Contract *ContractCallerSession) Quorum(blockNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.Quorum(&_Contract.CallOpts, blockNumber)
}

// QuorumDenominator is a free data retrieval call binding the contract method 0x97c3d334.
//
// Solidity: function quorumDenominator() view returns(uint256)
func (_Contract *ContractCaller) QuorumDenominator(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "quorumDenominator")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// QuorumDenominator is a free data retrieval call binding the contract method 0x97c3d334.
//
// Solidity: function quorumDenominator() view returns(uint256)
func (_Contract *ContractSession) QuorumDenominator() (*big.Int, error) {
	return _Contract.Contract.QuorumDenominator(&_Contract.CallOpts)
}

// QuorumDenominator is a free data retrieval call binding the contract method 0x97c3d334.
//
// Solidity: function quorumDenominator() view returns(uint256)
func (_Contract *ContractCallerSession) QuorumDenominator() (*big.Int, error) {
	return _Contract.Contract.QuorumDenominator(&_Contract.CallOpts)
}

// QuorumNumerator is a free data retrieval call binding the contract method 0x60c4247f.
//
// Solidity: function quorumNumerator(uint256 blockNumber) view returns(uint256)
func (_Contract *ContractCaller) QuorumNumerator(opts *bind.CallOpts, blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "quorumNumerator", blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// QuorumNumerator is a free data retrieval call binding the contract method 0x60c4247f.
//
// Solidity: function quorumNumerator(uint256 blockNumber) view returns(uint256)
func (_Contract *ContractSession) QuorumNumerator(blockNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.QuorumNumerator(&_Contract.CallOpts, blockNumber)
}

// QuorumNumerator is a free data retrieval call binding the contract method 0x60c4247f.
//
// Solidity: function quorumNumerator(uint256 blockNumber) view returns(uint256)
func (_Contract *ContractCallerSession) QuorumNumerator(blockNumber *big.Int) (*big.Int, error) {
	return _Contract.Contract.QuorumNumerator(&_Contract.CallOpts, blockNumber)
}

// QuorumNumerator0 is a free data retrieval call binding the contract method 0xa7713a70.
//
// Solidity: function quorumNumerator() view returns(uint256)
func (_Contract *ContractCaller) QuorumNumerator0(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "quorumNumerator0")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// QuorumNumerator0 is a free data retrieval call binding the contract method 0xa7713a70.
//
// Solidity: function quorumNumerator() view returns(uint256)
func (_Contract *ContractSession) QuorumNumerator0() (*big.Int, error) {
	return _Contract.Contract.QuorumNumerator0(&_Contract.CallOpts)
}

// QuorumNumerator0 is a free data retrieval call binding the contract method 0xa7713a70.
//
// Solidity: function quorumNumerator() view returns(uint256)
func (_Contract *ContractCallerSession) QuorumNumerator0() (*big.Int, error) {
	return _Contract.Contract.QuorumNumerator0(&_Contract.CallOpts)
}

// SlotCloseTime is a free data retrieval call binding the contract method 0x1b2d1728.
//
// Solidity: function slotCloseTime() view returns(uint256)
func (_Contract *ContractCaller) SlotCloseTime(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "slotCloseTime")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotCloseTime is a free data retrieval call binding the contract method 0x1b2d1728.
//
// Solidity: function slotCloseTime() view returns(uint256)
func (_Contract *ContractSession) SlotCloseTime() (*big.Int, error) {
	return _Contract.Contract.SlotCloseTime(&_Contract.CallOpts)
}

// SlotCloseTime is a free data retrieval call binding the contract method 0x1b2d1728.
//
// Solidity: function slotCloseTime() view returns(uint256)
func (_Contract *ContractCallerSession) SlotCloseTime() (*big.Int, error) {
	return _Contract.Contract.SlotCloseTime(&_Contract.CallOpts)
}

// Slots is a free data retrieval call binding the contract method 0x387dd9e9.
//
// Solidity: function slots(uint256 ) view returns(uint256)
func (_Contract *ContractCaller) Slots(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "slots", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Slots is a free data retrieval call binding the contract method 0x387dd9e9.
//
// Solidity: function slots(uint256 ) view returns(uint256)
func (_Contract *ContractSession) Slots(arg0 *big.Int) (*big.Int, error) {
	return _Contract.Contract.Slots(&_Contract.CallOpts, arg0)
}

// Slots is a free data retrieval call binding the contract method 0x387dd9e9.
//
// Solidity: function slots(uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) Slots(arg0 *big.Int) (*big.Int, error) {
	return _Contract.Contract.Slots(&_Contract.CallOpts, arg0)
}

// State is a free data retrieval call binding the contract method 0x3e4f49e6.
//
// Solidity: function state(uint256 proposalId) view returns(uint8)
func (_Contract *ContractCaller) State(opts *bind.CallOpts, proposalId *big.Int) (uint8, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "state", proposalId)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// State is a free data retrieval call binding the contract method 0x3e4f49e6.
//
// Solidity: function state(uint256 proposalId) view returns(uint8)
func (_Contract *ContractSession) State(proposalId *big.Int) (uint8, error) {
	return _Contract.Contract.State(&_Contract.CallOpts, proposalId)
}

// State is a free data retrieval call binding the contract method 0x3e4f49e6.
//
// Solidity: function state(uint256 proposalId) view returns(uint8)
func (_Contract *ContractCallerSession) State(proposalId *big.Int) (uint8, error) {
	return _Contract.Contract.State(&_Contract.CallOpts, proposalId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Contract *ContractCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Contract *ContractSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Contract.Contract.SupportsInterface(&_Contract.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Contract *ContractCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Contract.Contract.SupportsInterface(&_Contract.CallOpts, interfaceId)
}

// Timelock is a free data retrieval call binding the contract method 0xd33219b4.
//
// Solidity: function timelock() view returns(address)
func (_Contract *ContractCaller) Timelock(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "timelock")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Timelock is a free data retrieval call binding the contract method 0xd33219b4.
//
// Solidity: function timelock() view returns(address)
func (_Contract *ContractSession) Timelock() (common.Address, error) {
	return _Contract.Contract.Timelock(&_Contract.CallOpts)
}

// Timelock is a free data retrieval call binding the contract method 0xd33219b4.
//
// Solidity: function timelock() view returns(address)
func (_Contract *ContractCallerSession) Timelock() (common.Address, error) {
	return _Contract.Contract.Timelock(&_Contract.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Contract *ContractCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Contract *ContractSession) Token() (common.Address, error) {
	return _Contract.Contract.Token(&_Contract.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Contract *ContractCallerSession) Token() (common.Address, error) {
	return _Contract.Contract.Token(&_Contract.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_Contract *ContractCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_Contract *ContractSession) Version() (string, error) {
	return _Contract.Contract.Version(&_Contract.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_Contract *ContractCallerSession) Version() (string, error) {
	return _Contract.Contract.Version(&_Contract.CallOpts)
}

// VotingDelay is a free data retrieval call binding the contract method 0x3932abb1.
//
// Solidity: function votingDelay() pure returns(uint256)
func (_Contract *ContractCaller) VotingDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "votingDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VotingDelay is a free data retrieval call binding the contract method 0x3932abb1.
//
// Solidity: function votingDelay() pure returns(uint256)
func (_Contract *ContractSession) VotingDelay() (*big.Int, error) {
	return _Contract.Contract.VotingDelay(&_Contract.CallOpts)
}

// VotingDelay is a free data retrieval call binding the contract method 0x3932abb1.
//
// Solidity: function votingDelay() pure returns(uint256)
func (_Contract *ContractCallerSession) VotingDelay() (*big.Int, error) {
	return _Contract.Contract.VotingDelay(&_Contract.CallOpts)
}

// VotingPeriod is a free data retrieval call binding the contract method 0x02a251a3.
//
// Solidity: function votingPeriod() pure returns(uint256)
func (_Contract *ContractCaller) VotingPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "votingPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VotingPeriod is a free data retrieval call binding the contract method 0x02a251a3.
//
// Solidity: function votingPeriod() pure returns(uint256)
func (_Contract *ContractSession) VotingPeriod() (*big.Int, error) {
	return _Contract.Contract.VotingPeriod(&_Contract.CallOpts)
}

// VotingPeriod is a free data retrieval call binding the contract method 0x02a251a3.
//
// Solidity: function votingPeriod() pure returns(uint256)
func (_Contract *ContractCallerSession) VotingPeriod() (*big.Int, error) {
	return _Contract.Contract.VotingPeriod(&_Contract.CallOpts)
}

// VotingSlotAnnouncementPeriod is a free data retrieval call binding the contract method 0xe041d6dc.
//
// Solidity: function votingSlotAnnouncementPeriod() view returns(uint256)
func (_Contract *ContractCaller) VotingSlotAnnouncementPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "votingSlotAnnouncementPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VotingSlotAnnouncementPeriod is a free data retrieval call binding the contract method 0xe041d6dc.
//
// Solidity: function votingSlotAnnouncementPeriod() view returns(uint256)
func (_Contract *ContractSession) VotingSlotAnnouncementPeriod() (*big.Int, error) {
	return _Contract.Contract.VotingSlotAnnouncementPeriod(&_Contract.CallOpts)
}

// VotingSlotAnnouncementPeriod is a free data retrieval call binding the contract method 0xe041d6dc.
//
// Solidity: function votingSlotAnnouncementPeriod() view returns(uint256)
func (_Contract *ContractCallerSession) VotingSlotAnnouncementPeriod() (*big.Int, error) {
	return _Contract.Contract.VotingSlotAnnouncementPeriod(&_Contract.CallOpts)
}

// VotingSlots is a free data retrieval call binding the contract method 0x05106e53.
//
// Solidity: function votingSlots(uint256 , uint256 ) view returns(uint256)
func (_Contract *ContractCaller) VotingSlots(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "votingSlots", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VotingSlots is a free data retrieval call binding the contract method 0x05106e53.
//
// Solidity: function votingSlots(uint256 , uint256 ) view returns(uint256)
func (_Contract *ContractSession) VotingSlots(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.VotingSlots(&_Contract.CallOpts, arg0, arg1)
}

// VotingSlots is a free data retrieval call binding the contract method 0x05106e53.
//
// Solidity: function votingSlots(uint256 , uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) VotingSlots(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.VotingSlots(&_Contract.CallOpts, arg0, arg1)
}

// CancelVotingSlot is a paid mutator transaction binding the contract method 0x496911bb.
//
// Solidity: function cancelVotingSlot(uint256 blockNumber, string reason) returns()
func (_Contract *ContractTransactor) CancelVotingSlot(opts *bind.TransactOpts, blockNumber *big.Int, reason string) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "cancelVotingSlot", blockNumber, reason)
}

// CancelVotingSlot is a paid mutator transaction binding the contract method 0x496911bb.
//
// Solidity: function cancelVotingSlot(uint256 blockNumber, string reason) returns()
func (_Contract *ContractSession) CancelVotingSlot(blockNumber *big.Int, reason string) (*types.Transaction, error) {
	return _Contract.Contract.CancelVotingSlot(&_Contract.TransactOpts, blockNumber, reason)
}

// CancelVotingSlot is a paid mutator transaction binding the contract method 0x496911bb.
//
// Solidity: function cancelVotingSlot(uint256 blockNumber, string reason) returns()
func (_Contract *ContractTransactorSession) CancelVotingSlot(blockNumber *big.Int, reason string) (*types.Transaction, error) {
	return _Contract.Contract.CancelVotingSlot(&_Contract.TransactOpts, blockNumber, reason)
}

// CastVote is a paid mutator transaction binding the contract method 0x56781388.
//
// Solidity: function castVote(uint256 proposalId, uint8 support) returns(uint256)
func (_Contract *ContractTransactor) CastVote(opts *bind.TransactOpts, proposalId *big.Int, support uint8) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "castVote", proposalId, support)
}

// CastVote is a paid mutator transaction binding the contract method 0x56781388.
//
// Solidity: function castVote(uint256 proposalId, uint8 support) returns(uint256)
func (_Contract *ContractSession) CastVote(proposalId *big.Int, support uint8) (*types.Transaction, error) {
	return _Contract.Contract.CastVote(&_Contract.TransactOpts, proposalId, support)
}

// CastVote is a paid mutator transaction binding the contract method 0x56781388.
//
// Solidity: function castVote(uint256 proposalId, uint8 support) returns(uint256)
func (_Contract *ContractTransactorSession) CastVote(proposalId *big.Int, support uint8) (*types.Transaction, error) {
	return _Contract.Contract.CastVote(&_Contract.TransactOpts, proposalId, support)
}

// CastVoteBySig is a paid mutator transaction binding the contract method 0x3bccf4fd.
//
// Solidity: function castVoteBySig(uint256 proposalId, uint8 support, uint8 v, bytes32 r, bytes32 s) returns(uint256)
func (_Contract *ContractTransactor) CastVoteBySig(opts *bind.TransactOpts, proposalId *big.Int, support uint8, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "castVoteBySig", proposalId, support, v, r, s)
}

// CastVoteBySig is a paid mutator transaction binding the contract method 0x3bccf4fd.
//
// Solidity: function castVoteBySig(uint256 proposalId, uint8 support, uint8 v, bytes32 r, bytes32 s) returns(uint256)
func (_Contract *ContractSession) CastVoteBySig(proposalId *big.Int, support uint8, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteBySig(&_Contract.TransactOpts, proposalId, support, v, r, s)
}

// CastVoteBySig is a paid mutator transaction binding the contract method 0x3bccf4fd.
//
// Solidity: function castVoteBySig(uint256 proposalId, uint8 support, uint8 v, bytes32 r, bytes32 s) returns(uint256)
func (_Contract *ContractTransactorSession) CastVoteBySig(proposalId *big.Int, support uint8, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteBySig(&_Contract.TransactOpts, proposalId, support, v, r, s)
}

// CastVoteWithReason is a paid mutator transaction binding the contract method 0x7b3c71d3.
//
// Solidity: function castVoteWithReason(uint256 proposalId, uint8 support, string reason) returns(uint256)
func (_Contract *ContractTransactor) CastVoteWithReason(opts *bind.TransactOpts, proposalId *big.Int, support uint8, reason string) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "castVoteWithReason", proposalId, support, reason)
}

// CastVoteWithReason is a paid mutator transaction binding the contract method 0x7b3c71d3.
//
// Solidity: function castVoteWithReason(uint256 proposalId, uint8 support, string reason) returns(uint256)
func (_Contract *ContractSession) CastVoteWithReason(proposalId *big.Int, support uint8, reason string) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteWithReason(&_Contract.TransactOpts, proposalId, support, reason)
}

// CastVoteWithReason is a paid mutator transaction binding the contract method 0x7b3c71d3.
//
// Solidity: function castVoteWithReason(uint256 proposalId, uint8 support, string reason) returns(uint256)
func (_Contract *ContractTransactorSession) CastVoteWithReason(proposalId *big.Int, support uint8, reason string) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteWithReason(&_Contract.TransactOpts, proposalId, support, reason)
}

// CastVoteWithReasonAndParams is a paid mutator transaction binding the contract method 0x5f398a14.
//
// Solidity: function castVoteWithReasonAndParams(uint256 proposalId, uint8 support, string reason, bytes params) returns(uint256)
func (_Contract *ContractTransactor) CastVoteWithReasonAndParams(opts *bind.TransactOpts, proposalId *big.Int, support uint8, reason string, params []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "castVoteWithReasonAndParams", proposalId, support, reason, params)
}

// CastVoteWithReasonAndParams is a paid mutator transaction binding the contract method 0x5f398a14.
//
// Solidity: function castVoteWithReasonAndParams(uint256 proposalId, uint8 support, string reason, bytes params) returns(uint256)
func (_Contract *ContractSession) CastVoteWithReasonAndParams(proposalId *big.Int, support uint8, reason string, params []byte) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteWithReasonAndParams(&_Contract.TransactOpts, proposalId, support, reason, params)
}

// CastVoteWithReasonAndParams is a paid mutator transaction binding the contract method 0x5f398a14.
//
// Solidity: function castVoteWithReasonAndParams(uint256 proposalId, uint8 support, string reason, bytes params) returns(uint256)
func (_Contract *ContractTransactorSession) CastVoteWithReasonAndParams(proposalId *big.Int, support uint8, reason string, params []byte) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteWithReasonAndParams(&_Contract.TransactOpts, proposalId, support, reason, params)
}

// CastVoteWithReasonAndParamsBySig is a paid mutator transaction binding the contract method 0x03420181.
//
// Solidity: function castVoteWithReasonAndParamsBySig(uint256 proposalId, uint8 support, string reason, bytes params, uint8 v, bytes32 r, bytes32 s) returns(uint256)
func (_Contract *ContractTransactor) CastVoteWithReasonAndParamsBySig(opts *bind.TransactOpts, proposalId *big.Int, support uint8, reason string, params []byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "castVoteWithReasonAndParamsBySig", proposalId, support, reason, params, v, r, s)
}

// CastVoteWithReasonAndParamsBySig is a paid mutator transaction binding the contract method 0x03420181.
//
// Solidity: function castVoteWithReasonAndParamsBySig(uint256 proposalId, uint8 support, string reason, bytes params, uint8 v, bytes32 r, bytes32 s) returns(uint256)
func (_Contract *ContractSession) CastVoteWithReasonAndParamsBySig(proposalId *big.Int, support uint8, reason string, params []byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteWithReasonAndParamsBySig(&_Contract.TransactOpts, proposalId, support, reason, params, v, r, s)
}

// CastVoteWithReasonAndParamsBySig is a paid mutator transaction binding the contract method 0x03420181.
//
// Solidity: function castVoteWithReasonAndParamsBySig(uint256 proposalId, uint8 support, string reason, bytes params, uint8 v, bytes32 r, bytes32 s) returns(uint256)
func (_Contract *ContractTransactorSession) CastVoteWithReasonAndParamsBySig(proposalId *big.Int, support uint8, reason string, params []byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.CastVoteWithReasonAndParamsBySig(&_Contract.TransactOpts, proposalId, support, reason, params, v, r, s)
}

// DissolveDAO is a paid mutator transaction binding the contract method 0x81894d34.
//
// Solidity: function dissolveDAO() returns()
func (_Contract *ContractTransactor) DissolveDAO(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "dissolveDAO")
}

// DissolveDAO is a paid mutator transaction binding the contract method 0x81894d34.
//
// Solidity: function dissolveDAO() returns()
func (_Contract *ContractSession) DissolveDAO() (*types.Transaction, error) {
	return _Contract.Contract.DissolveDAO(&_Contract.TransactOpts)
}

// DissolveDAO is a paid mutator transaction binding the contract method 0x81894d34.
//
// Solidity: function dissolveDAO() returns()
func (_Contract *ContractTransactorSession) DissolveDAO() (*types.Transaction, error) {
	return _Contract.Contract.DissolveDAO(&_Contract.TransactOpts)
}

// Execute is a paid mutator transaction binding the contract method 0x2656227d.
//
// Solidity: function execute(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) payable returns(uint256)
func (_Contract *ContractTransactor) Execute(opts *bind.TransactOpts, targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "execute", targets, values, calldatas, descriptionHash)
}

// Execute is a paid mutator transaction binding the contract method 0x2656227d.
//
// Solidity: function execute(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) payable returns(uint256)
func (_Contract *ContractSession) Execute(targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Execute(&_Contract.TransactOpts, targets, values, calldatas, descriptionHash)
}

// Execute is a paid mutator transaction binding the contract method 0x2656227d.
//
// Solidity: function execute(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) payable returns(uint256)
func (_Contract *ContractTransactorSession) Execute(targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Execute(&_Contract.TransactOpts, targets, values, calldatas, descriptionHash)
}

// Initialize is a paid mutator transaction binding the contract method 0x2016a0d2.
//
// Solidity: function initialize(address _membership, address _timelock, string bylawsHash, string bylawsUrl) returns()
func (_Contract *ContractTransactor) Initialize(opts *bind.TransactOpts, _membership common.Address, _timelock common.Address, bylawsHash string, bylawsUrl string) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initialize", _membership, _timelock, bylawsHash, bylawsUrl)
}

// Initialize is a paid mutator transaction binding the contract method 0x2016a0d2.
//
// Solidity: function initialize(address _membership, address _timelock, string bylawsHash, string bylawsUrl) returns()
func (_Contract *ContractSession) Initialize(_membership common.Address, _timelock common.Address, bylawsHash string, bylawsUrl string) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _membership, _timelock, bylawsHash, bylawsUrl)
}

// Initialize is a paid mutator transaction binding the contract method 0x2016a0d2.
//
// Solidity: function initialize(address _membership, address _timelock, string bylawsHash, string bylawsUrl) returns()
func (_Contract *ContractTransactorSession) Initialize(_membership common.Address, _timelock common.Address, bylawsHash string, bylawsUrl string) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _membership, _timelock, bylawsHash, bylawsUrl)
}

// Propose is a paid mutator transaction binding the contract method 0x7d5e81e2.
//
// Solidity: function propose(address[] targets, uint256[] values, bytes[] calldatas, string description) returns(uint256)
func (_Contract *ContractTransactor) Propose(opts *bind.TransactOpts, targets []common.Address, values []*big.Int, calldatas [][]byte, description string) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "propose", targets, values, calldatas, description)
}

// Propose is a paid mutator transaction binding the contract method 0x7d5e81e2.
//
// Solidity: function propose(address[] targets, uint256[] values, bytes[] calldatas, string description) returns(uint256)
func (_Contract *ContractSession) Propose(targets []common.Address, values []*big.Int, calldatas [][]byte, description string) (*types.Transaction, error) {
	return _Contract.Contract.Propose(&_Contract.TransactOpts, targets, values, calldatas, description)
}

// Propose is a paid mutator transaction binding the contract method 0x7d5e81e2.
//
// Solidity: function propose(address[] targets, uint256[] values, bytes[] calldatas, string description) returns(uint256)
func (_Contract *ContractTransactorSession) Propose(targets []common.Address, values []*big.Int, calldatas [][]byte, description string) (*types.Transaction, error) {
	return _Contract.Contract.Propose(&_Contract.TransactOpts, targets, values, calldatas, description)
}

// Queue is a paid mutator transaction binding the contract method 0x160cbed7.
//
// Solidity: function queue(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) returns(uint256)
func (_Contract *ContractTransactor) Queue(opts *bind.TransactOpts, targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "queue", targets, values, calldatas, descriptionHash)
}

// Queue is a paid mutator transaction binding the contract method 0x160cbed7.
//
// Solidity: function queue(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) returns(uint256)
func (_Contract *ContractSession) Queue(targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Queue(&_Contract.TransactOpts, targets, values, calldatas, descriptionHash)
}

// Queue is a paid mutator transaction binding the contract method 0x160cbed7.
//
// Solidity: function queue(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) returns(uint256)
func (_Contract *ContractTransactorSession) Queue(targets []common.Address, values []*big.Int, calldatas [][]byte, descriptionHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Queue(&_Contract.TransactOpts, targets, values, calldatas, descriptionHash)
}

// Relay is a paid mutator transaction binding the contract method 0xc28bc2fa.
//
// Solidity: function relay(address target, uint256 value, bytes data) payable returns()
func (_Contract *ContractTransactor) Relay(opts *bind.TransactOpts, target common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "relay", target, value, data)
}

// Relay is a paid mutator transaction binding the contract method 0xc28bc2fa.
//
// Solidity: function relay(address target, uint256 value, bytes data) payable returns()
func (_Contract *ContractSession) Relay(target common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _Contract.Contract.Relay(&_Contract.TransactOpts, target, value, data)
}

// Relay is a paid mutator transaction binding the contract method 0xc28bc2fa.
//
// Solidity: function relay(address target, uint256 value, bytes data) payable returns()
func (_Contract *ContractTransactorSession) Relay(target common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _Contract.Contract.Relay(&_Contract.TransactOpts, target, value, data)
}

// SetAssociationDissolutionQuorumNominator is a paid mutator transaction binding the contract method 0x5eacd3c6.
//
// Solidity: function setAssociationDissolutionQuorumNominator(uint256 newValue) returns()
func (_Contract *ContractTransactor) SetAssociationDissolutionQuorumNominator(opts *bind.TransactOpts, newValue *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setAssociationDissolutionQuorumNominator", newValue)
}

// SetAssociationDissolutionQuorumNominator is a paid mutator transaction binding the contract method 0x5eacd3c6.
//
// Solidity: function setAssociationDissolutionQuorumNominator(uint256 newValue) returns()
func (_Contract *ContractSession) SetAssociationDissolutionQuorumNominator(newValue *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetAssociationDissolutionQuorumNominator(&_Contract.TransactOpts, newValue)
}

// SetAssociationDissolutionQuorumNominator is a paid mutator transaction binding the contract method 0x5eacd3c6.
//
// Solidity: function setAssociationDissolutionQuorumNominator(uint256 newValue) returns()
func (_Contract *ContractTransactorSession) SetAssociationDissolutionQuorumNominator(newValue *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetAssociationDissolutionQuorumNominator(&_Contract.TransactOpts, newValue)
}

// SetExtraOrdinaryAssemblyVotingPeriod is a paid mutator transaction binding the contract method 0x12a26e5d.
//
// Solidity: function setExtraOrdinaryAssemblyVotingPeriod(uint64 newExtraOrdinaryAssemblyVotingPeriod) returns()
func (_Contract *ContractTransactor) SetExtraOrdinaryAssemblyVotingPeriod(opts *bind.TransactOpts, newExtraOrdinaryAssemblyVotingPeriod uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setExtraOrdinaryAssemblyVotingPeriod", newExtraOrdinaryAssemblyVotingPeriod)
}

// SetExtraOrdinaryAssemblyVotingPeriod is a paid mutator transaction binding the contract method 0x12a26e5d.
//
// Solidity: function setExtraOrdinaryAssemblyVotingPeriod(uint64 newExtraOrdinaryAssemblyVotingPeriod) returns()
func (_Contract *ContractSession) SetExtraOrdinaryAssemblyVotingPeriod(newExtraOrdinaryAssemblyVotingPeriod uint64) (*types.Transaction, error) {
	return _Contract.Contract.SetExtraOrdinaryAssemblyVotingPeriod(&_Contract.TransactOpts, newExtraOrdinaryAssemblyVotingPeriod)
}

// SetExtraOrdinaryAssemblyVotingPeriod is a paid mutator transaction binding the contract method 0x12a26e5d.
//
// Solidity: function setExtraOrdinaryAssemblyVotingPeriod(uint64 newExtraOrdinaryAssemblyVotingPeriod) returns()
func (_Contract *ContractTransactorSession) SetExtraOrdinaryAssemblyVotingPeriod(newExtraOrdinaryAssemblyVotingPeriod uint64) (*types.Transaction, error) {
	return _Contract.Contract.SetExtraOrdinaryAssemblyVotingPeriod(&_Contract.TransactOpts, newExtraOrdinaryAssemblyVotingPeriod)
}

// SetExtraordinaryVoteQuorumNominator is a paid mutator transaction binding the contract method 0x990581d4.
//
// Solidity: function setExtraordinaryVoteQuorumNominator(uint256 newValue) returns()
func (_Contract *ContractTransactor) SetExtraordinaryVoteQuorumNominator(opts *bind.TransactOpts, newValue *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setExtraordinaryVoteQuorumNominator", newValue)
}

// SetExtraordinaryVoteQuorumNominator is a paid mutator transaction binding the contract method 0x990581d4.
//
// Solidity: function setExtraordinaryVoteQuorumNominator(uint256 newValue) returns()
func (_Contract *ContractSession) SetExtraordinaryVoteQuorumNominator(newValue *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetExtraordinaryVoteQuorumNominator(&_Contract.TransactOpts, newValue)
}

// SetExtraordinaryVoteQuorumNominator is a paid mutator transaction binding the contract method 0x990581d4.
//
// Solidity: function setExtraordinaryVoteQuorumNominator(uint256 newValue) returns()
func (_Contract *ContractTransactorSession) SetExtraordinaryVoteQuorumNominator(newValue *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetExtraordinaryVoteQuorumNominator(&_Contract.TransactOpts, newValue)
}

// SetNewBylaws is a paid mutator transaction binding the contract method 0x527e098f.
//
// Solidity: function setNewBylaws(string newBylawsHash, string newBylawsUrl) returns()
func (_Contract *ContractTransactor) SetNewBylaws(opts *bind.TransactOpts, newBylawsHash string, newBylawsUrl string) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setNewBylaws", newBylawsHash, newBylawsUrl)
}

// SetNewBylaws is a paid mutator transaction binding the contract method 0x527e098f.
//
// Solidity: function setNewBylaws(string newBylawsHash, string newBylawsUrl) returns()
func (_Contract *ContractSession) SetNewBylaws(newBylawsHash string, newBylawsUrl string) (*types.Transaction, error) {
	return _Contract.Contract.SetNewBylaws(&_Contract.TransactOpts, newBylawsHash, newBylawsUrl)
}

// SetNewBylaws is a paid mutator transaction binding the contract method 0x527e098f.
//
// Solidity: function setNewBylaws(string newBylawsHash, string newBylawsUrl) returns()
func (_Contract *ContractTransactorSession) SetNewBylaws(newBylawsHash string, newBylawsUrl string) (*types.Transaction, error) {
	return _Contract.Contract.SetNewBylaws(&_Contract.TransactOpts, newBylawsHash, newBylawsUrl)
}

// SetSlotCloseTime is a paid mutator transaction binding the contract method 0x96537303.
//
// Solidity: function setSlotCloseTime(uint256 newSlotCloseTime) returns()
func (_Contract *ContractTransactor) SetSlotCloseTime(opts *bind.TransactOpts, newSlotCloseTime *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setSlotCloseTime", newSlotCloseTime)
}

// SetSlotCloseTime is a paid mutator transaction binding the contract method 0x96537303.
//
// Solidity: function setSlotCloseTime(uint256 newSlotCloseTime) returns()
func (_Contract *ContractSession) SetSlotCloseTime(newSlotCloseTime *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetSlotCloseTime(&_Contract.TransactOpts, newSlotCloseTime)
}

// SetSlotCloseTime is a paid mutator transaction binding the contract method 0x96537303.
//
// Solidity: function setSlotCloseTime(uint256 newSlotCloseTime) returns()
func (_Contract *ContractTransactorSession) SetSlotCloseTime(newSlotCloseTime *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetSlotCloseTime(&_Contract.TransactOpts, newSlotCloseTime)
}

// SetVotingSlot is a paid mutator transaction binding the contract method 0xe12a6b77.
//
// Solidity: function setVotingSlot(uint256 blockNumber) returns(uint256)
func (_Contract *ContractTransactor) SetVotingSlot(opts *bind.TransactOpts, blockNumber *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setVotingSlot", blockNumber)
}

// SetVotingSlot is a paid mutator transaction binding the contract method 0xe12a6b77.
//
// Solidity: function setVotingSlot(uint256 blockNumber) returns(uint256)
func (_Contract *ContractSession) SetVotingSlot(blockNumber *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetVotingSlot(&_Contract.TransactOpts, blockNumber)
}

// SetVotingSlot is a paid mutator transaction binding the contract method 0xe12a6b77.
//
// Solidity: function setVotingSlot(uint256 blockNumber) returns(uint256)
func (_Contract *ContractTransactorSession) SetVotingSlot(blockNumber *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetVotingSlot(&_Contract.TransactOpts, blockNumber)
}

// SetVotingSlotAnnouncementPeriod is a paid mutator transaction binding the contract method 0x04e108dc.
//
// Solidity: function setVotingSlotAnnouncementPeriod(uint64 newVotingSlotAnnouncementPeriod) returns()
func (_Contract *ContractTransactor) SetVotingSlotAnnouncementPeriod(opts *bind.TransactOpts, newVotingSlotAnnouncementPeriod uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setVotingSlotAnnouncementPeriod", newVotingSlotAnnouncementPeriod)
}

// SetVotingSlotAnnouncementPeriod is a paid mutator transaction binding the contract method 0x04e108dc.
//
// Solidity: function setVotingSlotAnnouncementPeriod(uint64 newVotingSlotAnnouncementPeriod) returns()
func (_Contract *ContractSession) SetVotingSlotAnnouncementPeriod(newVotingSlotAnnouncementPeriod uint64) (*types.Transaction, error) {
	return _Contract.Contract.SetVotingSlotAnnouncementPeriod(&_Contract.TransactOpts, newVotingSlotAnnouncementPeriod)
}

// SetVotingSlotAnnouncementPeriod is a paid mutator transaction binding the contract method 0x04e108dc.
//
// Solidity: function setVotingSlotAnnouncementPeriod(uint64 newVotingSlotAnnouncementPeriod) returns()
func (_Contract *ContractTransactorSession) SetVotingSlotAnnouncementPeriod(newVotingSlotAnnouncementPeriod uint64) (*types.Transaction, error) {
	return _Contract.Contract.SetVotingSlotAnnouncementPeriod(&_Contract.TransactOpts, newVotingSlotAnnouncementPeriod)
}

// UpdateQuorumNumerator is a paid mutator transaction binding the contract method 0x06f3f9e6.
//
// Solidity: function updateQuorumNumerator(uint256 newQuorumNumerator) returns()
func (_Contract *ContractTransactor) UpdateQuorumNumerator(opts *bind.TransactOpts, newQuorumNumerator *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateQuorumNumerator", newQuorumNumerator)
}

// UpdateQuorumNumerator is a paid mutator transaction binding the contract method 0x06f3f9e6.
//
// Solidity: function updateQuorumNumerator(uint256 newQuorumNumerator) returns()
func (_Contract *ContractSession) UpdateQuorumNumerator(newQuorumNumerator *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateQuorumNumerator(&_Contract.TransactOpts, newQuorumNumerator)
}

// UpdateQuorumNumerator is a paid mutator transaction binding the contract method 0x06f3f9e6.
//
// Solidity: function updateQuorumNumerator(uint256 newQuorumNumerator) returns()
func (_Contract *ContractTransactorSession) UpdateQuorumNumerator(newQuorumNumerator *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateQuorumNumerator(&_Contract.TransactOpts, newQuorumNumerator)
}

// UpdateTimelock is a paid mutator transaction binding the contract method 0xa890c910.
//
// Solidity: function updateTimelock(address newTimelock) returns()
func (_Contract *ContractTransactor) UpdateTimelock(opts *bind.TransactOpts, newTimelock common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateTimelock", newTimelock)
}

// UpdateTimelock is a paid mutator transaction binding the contract method 0xa890c910.
//
// Solidity: function updateTimelock(address newTimelock) returns()
func (_Contract *ContractSession) UpdateTimelock(newTimelock common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpdateTimelock(&_Contract.TransactOpts, newTimelock)
}

// UpdateTimelock is a paid mutator transaction binding the contract method 0xa890c910.
//
// Solidity: function updateTimelock(address newTimelock) returns()
func (_Contract *ContractTransactorSession) UpdateTimelock(newTimelock common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpdateTimelock(&_Contract.TransactOpts, newTimelock)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Contract *ContractTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Contract *ContractSession) Receive() (*types.Transaction, error) {
	return _Contract.Contract.Receive(&_Contract.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Contract *ContractTransactorSession) Receive() (*types.Transaction, error) {
	return _Contract.Contract.Receive(&_Contract.TransactOpts)
}

// ContractBylawsChangedIterator is returned from FilterBylawsChanged and is used to iterate over the raw logs and unpacked data for BylawsChanged events raised by the Contract contract.
type ContractBylawsChangedIterator struct {
	Event *ContractBylawsChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractBylawsChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractBylawsChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractBylawsChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractBylawsChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractBylawsChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractBylawsChanged represents a BylawsChanged event raised by the Contract contract.
type ContractBylawsChanged struct {
	NewBylawsUrl  common.Hash
	NewBylawsHash common.Hash
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterBylawsChanged is a free log retrieval operation binding the contract event 0x20ef4846bccbf147482caf1cceca1dbc8b3c065fc5b6915a07485464545ebaa0.
//
// Solidity: event BylawsChanged(string indexed newBylawsUrl, string indexed newBylawsHash)
func (_Contract *ContractFilterer) FilterBylawsChanged(opts *bind.FilterOpts, newBylawsUrl []string, newBylawsHash []string) (*ContractBylawsChangedIterator, error) {

	var newBylawsUrlRule []interface{}
	for _, newBylawsUrlItem := range newBylawsUrl {
		newBylawsUrlRule = append(newBylawsUrlRule, newBylawsUrlItem)
	}
	var newBylawsHashRule []interface{}
	for _, newBylawsHashItem := range newBylawsHash {
		newBylawsHashRule = append(newBylawsHashRule, newBylawsHashItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "BylawsChanged", newBylawsUrlRule, newBylawsHashRule)
	if err != nil {
		return nil, err
	}
	return &ContractBylawsChangedIterator{contract: _Contract.contract, event: "BylawsChanged", logs: logs, sub: sub}, nil
}

// WatchBylawsChanged is a free log subscription operation binding the contract event 0x20ef4846bccbf147482caf1cceca1dbc8b3c065fc5b6915a07485464545ebaa0.
//
// Solidity: event BylawsChanged(string indexed newBylawsUrl, string indexed newBylawsHash)
func (_Contract *ContractFilterer) WatchBylawsChanged(opts *bind.WatchOpts, sink chan<- *ContractBylawsChanged, newBylawsUrl []string, newBylawsHash []string) (event.Subscription, error) {

	var newBylawsUrlRule []interface{}
	for _, newBylawsUrlItem := range newBylawsUrl {
		newBylawsUrlRule = append(newBylawsUrlRule, newBylawsUrlItem)
	}
	var newBylawsHashRule []interface{}
	for _, newBylawsHashItem := range newBylawsHash {
		newBylawsHashRule = append(newBylawsHashRule, newBylawsHashItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "BylawsChanged", newBylawsUrlRule, newBylawsHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractBylawsChanged)
				if err := _Contract.contract.UnpackLog(event, "BylawsChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBylawsChanged is a log parse operation binding the contract event 0x20ef4846bccbf147482caf1cceca1dbc8b3c065fc5b6915a07485464545ebaa0.
//
// Solidity: event BylawsChanged(string indexed newBylawsUrl, string indexed newBylawsHash)
func (_Contract *ContractFilterer) ParseBylawsChanged(log types.Log) (*ContractBylawsChanged, error) {
	event := new(ContractBylawsChanged)
	if err := _Contract.contract.UnpackLog(event, "BylawsChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDAOProposalCreatedIterator is returned from FilterDAOProposalCreated and is used to iterate over the raw logs and unpacked data for DAOProposalCreated events raised by the Contract contract.
type ContractDAOProposalCreatedIterator struct {
	Event *ContractDAOProposalCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDAOProposalCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDAOProposalCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDAOProposalCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDAOProposalCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDAOProposalCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDAOProposalCreated represents a DAOProposalCreated event raised by the Contract contract.
type ContractDAOProposalCreated struct {
	ProposalId  *big.Int
	Proposer    common.Address
	Targets     []common.Address
	Values      []*big.Int
	Signatures  []string
	Calldatas   [][]byte
	StartBlock  *big.Int
	EndBlock    *big.Int
	Description string
	Category    uint8
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterDAOProposalCreated is a free log retrieval operation binding the contract event 0x2546e02ec051925f1c459bd5da30927611040a32607c07d26172816e97d07bd1.
//
// Solidity: event DAOProposalCreated(uint256 indexed proposalId, address indexed proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description, uint8 indexed category)
func (_Contract *ContractFilterer) FilterDAOProposalCreated(opts *bind.FilterOpts, proposalId []*big.Int, proposer []common.Address, category []uint8) (*ContractDAOProposalCreatedIterator, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	var categoryRule []interface{}
	for _, categoryItem := range category {
		categoryRule = append(categoryRule, categoryItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "DAOProposalCreated", proposalIdRule, proposerRule, categoryRule)
	if err != nil {
		return nil, err
	}
	return &ContractDAOProposalCreatedIterator{contract: _Contract.contract, event: "DAOProposalCreated", logs: logs, sub: sub}, nil
}

// WatchDAOProposalCreated is a free log subscription operation binding the contract event 0x2546e02ec051925f1c459bd5da30927611040a32607c07d26172816e97d07bd1.
//
// Solidity: event DAOProposalCreated(uint256 indexed proposalId, address indexed proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description, uint8 indexed category)
func (_Contract *ContractFilterer) WatchDAOProposalCreated(opts *bind.WatchOpts, sink chan<- *ContractDAOProposalCreated, proposalId []*big.Int, proposer []common.Address, category []uint8) (event.Subscription, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	var categoryRule []interface{}
	for _, categoryItem := range category {
		categoryRule = append(categoryRule, categoryItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "DAOProposalCreated", proposalIdRule, proposerRule, categoryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDAOProposalCreated)
				if err := _Contract.contract.UnpackLog(event, "DAOProposalCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDAOProposalCreated is a log parse operation binding the contract event 0x2546e02ec051925f1c459bd5da30927611040a32607c07d26172816e97d07bd1.
//
// Solidity: event DAOProposalCreated(uint256 indexed proposalId, address indexed proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description, uint8 indexed category)
func (_Contract *ContractFilterer) ParseDAOProposalCreated(log types.Log) (*ContractDAOProposalCreated, error) {
	event := new(ContractDAOProposalCreated)
	if err := _Contract.contract.UnpackLog(event, "DAOProposalCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractExtraOrdinaryAssemblyRequestedIterator is returned from FilterExtraOrdinaryAssemblyRequested and is used to iterate over the raw logs and unpacked data for ExtraOrdinaryAssemblyRequested events raised by the Contract contract.
type ContractExtraOrdinaryAssemblyRequestedIterator struct {
	Event *ContractExtraOrdinaryAssemblyRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractExtraOrdinaryAssemblyRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractExtraOrdinaryAssemblyRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractExtraOrdinaryAssemblyRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractExtraOrdinaryAssemblyRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractExtraOrdinaryAssemblyRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractExtraOrdinaryAssemblyRequested represents a ExtraOrdinaryAssemblyRequested event raised by the Contract contract.
type ContractExtraOrdinaryAssemblyRequested struct {
	ProposalId  *big.Int
	Proposer    common.Address
	Targets     []common.Address
	Values      []*big.Int
	Signatures  []string
	Calldatas   [][]byte
	StartBlock  *big.Int
	EndBlock    *big.Int
	Description string
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterExtraOrdinaryAssemblyRequested is a free log retrieval operation binding the contract event 0xaf70b28462b17da839f4faad55a92330c675fb47d34733424983f6a314a1bd43.
//
// Solidity: event ExtraOrdinaryAssemblyRequested(uint256 indexed proposalId, address indexed proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description)
func (_Contract *ContractFilterer) FilterExtraOrdinaryAssemblyRequested(opts *bind.FilterOpts, proposalId []*big.Int, proposer []common.Address) (*ContractExtraOrdinaryAssemblyRequestedIterator, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ExtraOrdinaryAssemblyRequested", proposalIdRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return &ContractExtraOrdinaryAssemblyRequestedIterator{contract: _Contract.contract, event: "ExtraOrdinaryAssemblyRequested", logs: logs, sub: sub}, nil
}

// WatchExtraOrdinaryAssemblyRequested is a free log subscription operation binding the contract event 0xaf70b28462b17da839f4faad55a92330c675fb47d34733424983f6a314a1bd43.
//
// Solidity: event ExtraOrdinaryAssemblyRequested(uint256 indexed proposalId, address indexed proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description)
func (_Contract *ContractFilterer) WatchExtraOrdinaryAssemblyRequested(opts *bind.WatchOpts, sink chan<- *ContractExtraOrdinaryAssemblyRequested, proposalId []*big.Int, proposer []common.Address) (event.Subscription, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ExtraOrdinaryAssemblyRequested", proposalIdRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractExtraOrdinaryAssemblyRequested)
				if err := _Contract.contract.UnpackLog(event, "ExtraOrdinaryAssemblyRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseExtraOrdinaryAssemblyRequested is a log parse operation binding the contract event 0xaf70b28462b17da839f4faad55a92330c675fb47d34733424983f6a314a1bd43.
//
// Solidity: event ExtraOrdinaryAssemblyRequested(uint256 indexed proposalId, address indexed proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description)
func (_Contract *ContractFilterer) ParseExtraOrdinaryAssemblyRequested(log types.Log) (*ContractExtraOrdinaryAssemblyRequested, error) {
	event := new(ContractExtraOrdinaryAssemblyRequested)
	if err := _Contract.contract.UnpackLog(event, "ExtraOrdinaryAssemblyRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Contract contract.
type ContractInitializedIterator struct {
	Event *ContractInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractInitialized represents a Initialized event raised by the Contract contract.
type ContractInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Contract *ContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*ContractInitializedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ContractInitializedIterator{contract: _Contract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Contract *ContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractInitialized) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractInitialized)
				if err := _Contract.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Contract *ContractFilterer) ParseInitialized(log types.Log) (*ContractInitialized, error) {
	event := new(ContractInitialized)
	if err := _Contract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractNewTimeslotSetIterator is returned from FilterNewTimeslotSet and is used to iterate over the raw logs and unpacked data for NewTimeslotSet events raised by the Contract contract.
type ContractNewTimeslotSetIterator struct {
	Event *ContractNewTimeslotSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractNewTimeslotSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractNewTimeslotSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractNewTimeslotSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractNewTimeslotSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractNewTimeslotSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractNewTimeslotSet represents a NewTimeslotSet event raised by the Contract contract.
type ContractNewTimeslotSet struct {
	Timeslot *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNewTimeslotSet is a free log retrieval operation binding the contract event 0xf92ed14e81753ef9799e397ab10278c4107b6807c59683338ced6b3598297c6b.
//
// Solidity: event NewTimeslotSet(uint256 timeslot)
func (_Contract *ContractFilterer) FilterNewTimeslotSet(opts *bind.FilterOpts) (*ContractNewTimeslotSetIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "NewTimeslotSet")
	if err != nil {
		return nil, err
	}
	return &ContractNewTimeslotSetIterator{contract: _Contract.contract, event: "NewTimeslotSet", logs: logs, sub: sub}, nil
}

// WatchNewTimeslotSet is a free log subscription operation binding the contract event 0xf92ed14e81753ef9799e397ab10278c4107b6807c59683338ced6b3598297c6b.
//
// Solidity: event NewTimeslotSet(uint256 timeslot)
func (_Contract *ContractFilterer) WatchNewTimeslotSet(opts *bind.WatchOpts, sink chan<- *ContractNewTimeslotSet) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "NewTimeslotSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractNewTimeslotSet)
				if err := _Contract.contract.UnpackLog(event, "NewTimeslotSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewTimeslotSet is a log parse operation binding the contract event 0xf92ed14e81753ef9799e397ab10278c4107b6807c59683338ced6b3598297c6b.
//
// Solidity: event NewTimeslotSet(uint256 timeslot)
func (_Contract *ContractFilterer) ParseNewTimeslotSet(log types.Log) (*ContractNewTimeslotSet, error) {
	event := new(ContractNewTimeslotSet)
	if err := _Contract.contract.UnpackLog(event, "NewTimeslotSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractProposalCanceledIterator is returned from FilterProposalCanceled and is used to iterate over the raw logs and unpacked data for ProposalCanceled events raised by the Contract contract.
type ContractProposalCanceledIterator struct {
	Event *ContractProposalCanceled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractProposalCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractProposalCanceled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractProposalCanceled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractProposalCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractProposalCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractProposalCanceled represents a ProposalCanceled event raised by the Contract contract.
type ContractProposalCanceled struct {
	ProposalId *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalCanceled is a free log retrieval operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 proposalId)
func (_Contract *ContractFilterer) FilterProposalCanceled(opts *bind.FilterOpts) (*ContractProposalCanceledIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ProposalCanceled")
	if err != nil {
		return nil, err
	}
	return &ContractProposalCanceledIterator{contract: _Contract.contract, event: "ProposalCanceled", logs: logs, sub: sub}, nil
}

// WatchProposalCanceled is a free log subscription operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 proposalId)
func (_Contract *ContractFilterer) WatchProposalCanceled(opts *bind.WatchOpts, sink chan<- *ContractProposalCanceled) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ProposalCanceled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractProposalCanceled)
				if err := _Contract.contract.UnpackLog(event, "ProposalCanceled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProposalCanceled is a log parse operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 proposalId)
func (_Contract *ContractFilterer) ParseProposalCanceled(log types.Log) (*ContractProposalCanceled, error) {
	event := new(ContractProposalCanceled)
	if err := _Contract.contract.UnpackLog(event, "ProposalCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractProposalCreatedIterator is returned from FilterProposalCreated and is used to iterate over the raw logs and unpacked data for ProposalCreated events raised by the Contract contract.
type ContractProposalCreatedIterator struct {
	Event *ContractProposalCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractProposalCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractProposalCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractProposalCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractProposalCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractProposalCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractProposalCreated represents a ProposalCreated event raised by the Contract contract.
type ContractProposalCreated struct {
	ProposalId  *big.Int
	Proposer    common.Address
	Targets     []common.Address
	Values      []*big.Int
	Signatures  []string
	Calldatas   [][]byte
	StartBlock  *big.Int
	EndBlock    *big.Int
	Description string
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterProposalCreated is a free log retrieval operation binding the contract event 0x7d84a6263ae0d98d3329bd7b46bb4e8d6f98cd35a7adb45c274c8b7fd5ebd5e0.
//
// Solidity: event ProposalCreated(uint256 proposalId, address proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description)
func (_Contract *ContractFilterer) FilterProposalCreated(opts *bind.FilterOpts) (*ContractProposalCreatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ProposalCreated")
	if err != nil {
		return nil, err
	}
	return &ContractProposalCreatedIterator{contract: _Contract.contract, event: "ProposalCreated", logs: logs, sub: sub}, nil
}

// WatchProposalCreated is a free log subscription operation binding the contract event 0x7d84a6263ae0d98d3329bd7b46bb4e8d6f98cd35a7adb45c274c8b7fd5ebd5e0.
//
// Solidity: event ProposalCreated(uint256 proposalId, address proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description)
func (_Contract *ContractFilterer) WatchProposalCreated(opts *bind.WatchOpts, sink chan<- *ContractProposalCreated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ProposalCreated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractProposalCreated)
				if err := _Contract.contract.UnpackLog(event, "ProposalCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProposalCreated is a log parse operation binding the contract event 0x7d84a6263ae0d98d3329bd7b46bb4e8d6f98cd35a7adb45c274c8b7fd5ebd5e0.
//
// Solidity: event ProposalCreated(uint256 proposalId, address proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description)
func (_Contract *ContractFilterer) ParseProposalCreated(log types.Log) (*ContractProposalCreated, error) {
	event := new(ContractProposalCreated)
	if err := _Contract.contract.UnpackLog(event, "ProposalCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractProposalExecutedIterator is returned from FilterProposalExecuted and is used to iterate over the raw logs and unpacked data for ProposalExecuted events raised by the Contract contract.
type ContractProposalExecutedIterator struct {
	Event *ContractProposalExecuted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractProposalExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractProposalExecuted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractProposalExecuted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractProposalExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractProposalExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractProposalExecuted represents a ProposalExecuted event raised by the Contract contract.
type ContractProposalExecuted struct {
	ProposalId *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalExecuted is a free log retrieval operation binding the contract event 0x712ae1383f79ac853f8d882153778e0260ef8f03b504e2866e0593e04d2b291f.
//
// Solidity: event ProposalExecuted(uint256 proposalId)
func (_Contract *ContractFilterer) FilterProposalExecuted(opts *bind.FilterOpts) (*ContractProposalExecutedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ProposalExecuted")
	if err != nil {
		return nil, err
	}
	return &ContractProposalExecutedIterator{contract: _Contract.contract, event: "ProposalExecuted", logs: logs, sub: sub}, nil
}

// WatchProposalExecuted is a free log subscription operation binding the contract event 0x712ae1383f79ac853f8d882153778e0260ef8f03b504e2866e0593e04d2b291f.
//
// Solidity: event ProposalExecuted(uint256 proposalId)
func (_Contract *ContractFilterer) WatchProposalExecuted(opts *bind.WatchOpts, sink chan<- *ContractProposalExecuted) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ProposalExecuted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractProposalExecuted)
				if err := _Contract.contract.UnpackLog(event, "ProposalExecuted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProposalExecuted is a log parse operation binding the contract event 0x712ae1383f79ac853f8d882153778e0260ef8f03b504e2866e0593e04d2b291f.
//
// Solidity: event ProposalExecuted(uint256 proposalId)
func (_Contract *ContractFilterer) ParseProposalExecuted(log types.Log) (*ContractProposalExecuted, error) {
	event := new(ContractProposalExecuted)
	if err := _Contract.contract.UnpackLog(event, "ProposalExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractProposalQueuedIterator is returned from FilterProposalQueued and is used to iterate over the raw logs and unpacked data for ProposalQueued events raised by the Contract contract.
type ContractProposalQueuedIterator struct {
	Event *ContractProposalQueued // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractProposalQueuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractProposalQueued)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractProposalQueued)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractProposalQueuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractProposalQueuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractProposalQueued represents a ProposalQueued event raised by the Contract contract.
type ContractProposalQueued struct {
	ProposalId *big.Int
	Eta        *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalQueued is a free log retrieval operation binding the contract event 0x9a2e42fd6722813d69113e7d0079d3d940171428df7373df9c7f7617cfda2892.
//
// Solidity: event ProposalQueued(uint256 proposalId, uint256 eta)
func (_Contract *ContractFilterer) FilterProposalQueued(opts *bind.FilterOpts) (*ContractProposalQueuedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ProposalQueued")
	if err != nil {
		return nil, err
	}
	return &ContractProposalQueuedIterator{contract: _Contract.contract, event: "ProposalQueued", logs: logs, sub: sub}, nil
}

// WatchProposalQueued is a free log subscription operation binding the contract event 0x9a2e42fd6722813d69113e7d0079d3d940171428df7373df9c7f7617cfda2892.
//
// Solidity: event ProposalQueued(uint256 proposalId, uint256 eta)
func (_Contract *ContractFilterer) WatchProposalQueued(opts *bind.WatchOpts, sink chan<- *ContractProposalQueued) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ProposalQueued")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractProposalQueued)
				if err := _Contract.contract.UnpackLog(event, "ProposalQueued", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProposalQueued is a log parse operation binding the contract event 0x9a2e42fd6722813d69113e7d0079d3d940171428df7373df9c7f7617cfda2892.
//
// Solidity: event ProposalQueued(uint256 proposalId, uint256 eta)
func (_Contract *ContractFilterer) ParseProposalQueued(log types.Log) (*ContractProposalQueued, error) {
	event := new(ContractProposalQueued)
	if err := _Contract.contract.UnpackLog(event, "ProposalQueued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractProposalVotingTimeChangedIterator is returned from FilterProposalVotingTimeChanged and is used to iterate over the raw logs and unpacked data for ProposalVotingTimeChanged events raised by the Contract contract.
type ContractProposalVotingTimeChangedIterator struct {
	Event *ContractProposalVotingTimeChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractProposalVotingTimeChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractProposalVotingTimeChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractProposalVotingTimeChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractProposalVotingTimeChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractProposalVotingTimeChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractProposalVotingTimeChanged represents a ProposalVotingTimeChanged event raised by the Contract contract.
type ContractProposalVotingTimeChanged struct {
	ProposalId *big.Int
	OldTime    uint64
	NewTime    uint64
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalVotingTimeChanged is a free log retrieval operation binding the contract event 0x53982bcc38da32ef033ffe2a31abe0b4704a5da42de9e54f28f69c1253307b37.
//
// Solidity: event ProposalVotingTimeChanged(uint256 proposalId, uint64 oldTime, uint64 newTime)
func (_Contract *ContractFilterer) FilterProposalVotingTimeChanged(opts *bind.FilterOpts) (*ContractProposalVotingTimeChangedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ProposalVotingTimeChanged")
	if err != nil {
		return nil, err
	}
	return &ContractProposalVotingTimeChangedIterator{contract: _Contract.contract, event: "ProposalVotingTimeChanged", logs: logs, sub: sub}, nil
}

// WatchProposalVotingTimeChanged is a free log subscription operation binding the contract event 0x53982bcc38da32ef033ffe2a31abe0b4704a5da42de9e54f28f69c1253307b37.
//
// Solidity: event ProposalVotingTimeChanged(uint256 proposalId, uint64 oldTime, uint64 newTime)
func (_Contract *ContractFilterer) WatchProposalVotingTimeChanged(opts *bind.WatchOpts, sink chan<- *ContractProposalVotingTimeChanged) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ProposalVotingTimeChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractProposalVotingTimeChanged)
				if err := _Contract.contract.UnpackLog(event, "ProposalVotingTimeChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProposalVotingTimeChanged is a log parse operation binding the contract event 0x53982bcc38da32ef033ffe2a31abe0b4704a5da42de9e54f28f69c1253307b37.
//
// Solidity: event ProposalVotingTimeChanged(uint256 proposalId, uint64 oldTime, uint64 newTime)
func (_Contract *ContractFilterer) ParseProposalVotingTimeChanged(log types.Log) (*ContractProposalVotingTimeChanged, error) {
	event := new(ContractProposalVotingTimeChanged)
	if err := _Contract.contract.UnpackLog(event, "ProposalVotingTimeChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractQuorumNumeratorUpdatedIterator is returned from FilterQuorumNumeratorUpdated and is used to iterate over the raw logs and unpacked data for QuorumNumeratorUpdated events raised by the Contract contract.
type ContractQuorumNumeratorUpdatedIterator struct {
	Event *ContractQuorumNumeratorUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractQuorumNumeratorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractQuorumNumeratorUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractQuorumNumeratorUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractQuorumNumeratorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractQuorumNumeratorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractQuorumNumeratorUpdated represents a QuorumNumeratorUpdated event raised by the Contract contract.
type ContractQuorumNumeratorUpdated struct {
	OldQuorumNumerator *big.Int
	NewQuorumNumerator *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterQuorumNumeratorUpdated is a free log retrieval operation binding the contract event 0x0553476bf02ef2726e8ce5ced78d63e26e602e4a2257b1f559418e24b4633997.
//
// Solidity: event QuorumNumeratorUpdated(uint256 oldQuorumNumerator, uint256 newQuorumNumerator)
func (_Contract *ContractFilterer) FilterQuorumNumeratorUpdated(opts *bind.FilterOpts) (*ContractQuorumNumeratorUpdatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "QuorumNumeratorUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractQuorumNumeratorUpdatedIterator{contract: _Contract.contract, event: "QuorumNumeratorUpdated", logs: logs, sub: sub}, nil
}

// WatchQuorumNumeratorUpdated is a free log subscription operation binding the contract event 0x0553476bf02ef2726e8ce5ced78d63e26e602e4a2257b1f559418e24b4633997.
//
// Solidity: event QuorumNumeratorUpdated(uint256 oldQuorumNumerator, uint256 newQuorumNumerator)
func (_Contract *ContractFilterer) WatchQuorumNumeratorUpdated(opts *bind.WatchOpts, sink chan<- *ContractQuorumNumeratorUpdated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "QuorumNumeratorUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractQuorumNumeratorUpdated)
				if err := _Contract.contract.UnpackLog(event, "QuorumNumeratorUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseQuorumNumeratorUpdated is a log parse operation binding the contract event 0x0553476bf02ef2726e8ce5ced78d63e26e602e4a2257b1f559418e24b4633997.
//
// Solidity: event QuorumNumeratorUpdated(uint256 oldQuorumNumerator, uint256 newQuorumNumerator)
func (_Contract *ContractFilterer) ParseQuorumNumeratorUpdated(log types.Log) (*ContractQuorumNumeratorUpdated, error) {
	event := new(ContractQuorumNumeratorUpdated)
	if err := _Contract.contract.UnpackLog(event, "QuorumNumeratorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTimelockChangeIterator is returned from FilterTimelockChange and is used to iterate over the raw logs and unpacked data for TimelockChange events raised by the Contract contract.
type ContractTimelockChangeIterator struct {
	Event *ContractTimelockChange // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractTimelockChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTimelockChange)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractTimelockChange)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractTimelockChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTimelockChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTimelockChange represents a TimelockChange event raised by the Contract contract.
type ContractTimelockChange struct {
	OldTimelock common.Address
	NewTimelock common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterTimelockChange is a free log retrieval operation binding the contract event 0x08f74ea46ef7894f65eabfb5e6e695de773a000b47c529ab559178069b226401.
//
// Solidity: event TimelockChange(address oldTimelock, address newTimelock)
func (_Contract *ContractFilterer) FilterTimelockChange(opts *bind.FilterOpts) (*ContractTimelockChangeIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "TimelockChange")
	if err != nil {
		return nil, err
	}
	return &ContractTimelockChangeIterator{contract: _Contract.contract, event: "TimelockChange", logs: logs, sub: sub}, nil
}

// WatchTimelockChange is a free log subscription operation binding the contract event 0x08f74ea46ef7894f65eabfb5e6e695de773a000b47c529ab559178069b226401.
//
// Solidity: event TimelockChange(address oldTimelock, address newTimelock)
func (_Contract *ContractFilterer) WatchTimelockChange(opts *bind.WatchOpts, sink chan<- *ContractTimelockChange) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "TimelockChange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTimelockChange)
				if err := _Contract.contract.UnpackLog(event, "TimelockChange", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTimelockChange is a log parse operation binding the contract event 0x08f74ea46ef7894f65eabfb5e6e695de773a000b47c529ab559178069b226401.
//
// Solidity: event TimelockChange(address oldTimelock, address newTimelock)
func (_Contract *ContractFilterer) ParseTimelockChange(log types.Log) (*ContractTimelockChange, error) {
	event := new(ContractTimelockChange)
	if err := _Contract.contract.UnpackLog(event, "TimelockChange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractVoteCastIterator is returned from FilterVoteCast and is used to iterate over the raw logs and unpacked data for VoteCast events raised by the Contract contract.
type ContractVoteCastIterator struct {
	Event *ContractVoteCast // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractVoteCastIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractVoteCast)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractVoteCast)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractVoteCastIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractVoteCastIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractVoteCast represents a VoteCast event raised by the Contract contract.
type ContractVoteCast struct {
	Voter      common.Address
	ProposalId *big.Int
	Support    uint8
	Weight     *big.Int
	Reason     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterVoteCast is a free log retrieval operation binding the contract event 0xb8e138887d0aa13bab447e82de9d5c1777041ecd21ca36ba824ff1e6c07ddda4.
//
// Solidity: event VoteCast(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason)
func (_Contract *ContractFilterer) FilterVoteCast(opts *bind.FilterOpts, voter []common.Address) (*ContractVoteCastIterator, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "VoteCast", voterRule)
	if err != nil {
		return nil, err
	}
	return &ContractVoteCastIterator{contract: _Contract.contract, event: "VoteCast", logs: logs, sub: sub}, nil
}

// WatchVoteCast is a free log subscription operation binding the contract event 0xb8e138887d0aa13bab447e82de9d5c1777041ecd21ca36ba824ff1e6c07ddda4.
//
// Solidity: event VoteCast(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason)
func (_Contract *ContractFilterer) WatchVoteCast(opts *bind.WatchOpts, sink chan<- *ContractVoteCast, voter []common.Address) (event.Subscription, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "VoteCast", voterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractVoteCast)
				if err := _Contract.contract.UnpackLog(event, "VoteCast", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVoteCast is a log parse operation binding the contract event 0xb8e138887d0aa13bab447e82de9d5c1777041ecd21ca36ba824ff1e6c07ddda4.
//
// Solidity: event VoteCast(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason)
func (_Contract *ContractFilterer) ParseVoteCast(log types.Log) (*ContractVoteCast, error) {
	event := new(ContractVoteCast)
	if err := _Contract.contract.UnpackLog(event, "VoteCast", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractVoteCastWithParamsIterator is returned from FilterVoteCastWithParams and is used to iterate over the raw logs and unpacked data for VoteCastWithParams events raised by the Contract contract.
type ContractVoteCastWithParamsIterator struct {
	Event *ContractVoteCastWithParams // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractVoteCastWithParamsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractVoteCastWithParams)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractVoteCastWithParams)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractVoteCastWithParamsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractVoteCastWithParamsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractVoteCastWithParams represents a VoteCastWithParams event raised by the Contract contract.
type ContractVoteCastWithParams struct {
	Voter      common.Address
	ProposalId *big.Int
	Support    uint8
	Weight     *big.Int
	Reason     string
	Params     []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterVoteCastWithParams is a free log retrieval operation binding the contract event 0xe2babfbac5889a709b63bb7f598b324e08bc5a4fb9ec647fb3cbc9ec07eb8712.
//
// Solidity: event VoteCastWithParams(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason, bytes params)
func (_Contract *ContractFilterer) FilterVoteCastWithParams(opts *bind.FilterOpts, voter []common.Address) (*ContractVoteCastWithParamsIterator, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "VoteCastWithParams", voterRule)
	if err != nil {
		return nil, err
	}
	return &ContractVoteCastWithParamsIterator{contract: _Contract.contract, event: "VoteCastWithParams", logs: logs, sub: sub}, nil
}

// WatchVoteCastWithParams is a free log subscription operation binding the contract event 0xe2babfbac5889a709b63bb7f598b324e08bc5a4fb9ec647fb3cbc9ec07eb8712.
//
// Solidity: event VoteCastWithParams(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason, bytes params)
func (_Contract *ContractFilterer) WatchVoteCastWithParams(opts *bind.WatchOpts, sink chan<- *ContractVoteCastWithParams, voter []common.Address) (event.Subscription, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "VoteCastWithParams", voterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractVoteCastWithParams)
				if err := _Contract.contract.UnpackLog(event, "VoteCastWithParams", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVoteCastWithParams is a log parse operation binding the contract event 0xe2babfbac5889a709b63bb7f598b324e08bc5a4fb9ec647fb3cbc9ec07eb8712.
//
// Solidity: event VoteCastWithParams(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason, bytes params)
func (_Contract *ContractFilterer) ParseVoteCastWithParams(log types.Log) (*ContractVoteCastWithParams, error) {
	event := new(ContractVoteCastWithParams)
	if err := _Contract.contract.UnpackLog(event, "VoteCastWithParams", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractVotingSlotCancelledIterator is returned from FilterVotingSlotCancelled and is used to iterate over the raw logs and unpacked data for VotingSlotCancelled events raised by the Contract contract.
type ContractVotingSlotCancelledIterator struct {
	Event *ContractVotingSlotCancelled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractVotingSlotCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractVotingSlotCancelled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractVotingSlotCancelled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractVotingSlotCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractVotingSlotCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractVotingSlotCancelled represents a VotingSlotCancelled event raised by the Contract contract.
type ContractVotingSlotCancelled struct {
	BlockNumber *big.Int
	Reason      string
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterVotingSlotCancelled is a free log retrieval operation binding the contract event 0xee810d41c3fb919057fecdc242e74ea856ae93f10c8be7192ffe22244bd93c2b.
//
// Solidity: event VotingSlotCancelled(uint256 blockNumber, string reason)
func (_Contract *ContractFilterer) FilterVotingSlotCancelled(opts *bind.FilterOpts) (*ContractVotingSlotCancelledIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "VotingSlotCancelled")
	if err != nil {
		return nil, err
	}
	return &ContractVotingSlotCancelledIterator{contract: _Contract.contract, event: "VotingSlotCancelled", logs: logs, sub: sub}, nil
}

// WatchVotingSlotCancelled is a free log subscription operation binding the contract event 0xee810d41c3fb919057fecdc242e74ea856ae93f10c8be7192ffe22244bd93c2b.
//
// Solidity: event VotingSlotCancelled(uint256 blockNumber, string reason)
func (_Contract *ContractFilterer) WatchVotingSlotCancelled(opts *bind.WatchOpts, sink chan<- *ContractVotingSlotCancelled) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "VotingSlotCancelled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractVotingSlotCancelled)
				if err := _Contract.contract.UnpackLog(event, "VotingSlotCancelled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVotingSlotCancelled is a log parse operation binding the contract event 0xee810d41c3fb919057fecdc242e74ea856ae93f10c8be7192ffe22244bd93c2b.
//
// Solidity: event VotingSlotCancelled(uint256 blockNumber, string reason)
func (_Contract *ContractFilterer) ParseVotingSlotCancelled(log types.Log) (*ContractVotingSlotCancelled, error) {
	event := new(ContractVotingSlotCancelled)
	if err := _Contract.contract.UnpackLog(event, "VotingSlotCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
