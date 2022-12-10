// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts (last updated v4.8.0-rc.1) (governance/Governor.sol)

pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/math/SafeCastUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/structs/DoubleEndedQueueUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/AddressUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ContextUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/TimersUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/IGovernorUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

abstract contract GovernorUpgradeable is
    Initializable,
    ContextUpgradeable,
    EIP712Upgradeable,
    IGovernorUpgradeable
{
    using DoubleEndedQueueUpgradeable for DoubleEndedQueueUpgradeable.Bytes32Deque;
    using SafeCastUpgradeable for uint256;
    using TimersUpgradeable for TimersUpgradeable.BlockNumber;

    bytes32 public constant BALLOT_TYPEHASH =
        keccak256("Ballot(uint256 proposalId,uint8 support)");
    bytes32 public constant EXTENDED_BALLOT_TYPEHASH =
        keccak256(
            "ExtendedBallot(uint256 proposalId,uint8 support,string reason,bytes params)"
        );
    bytes4 private constant setVotingSlotSignature =
        bytes4(keccak256("setVotingSlot(uint256)"));

    event DAAProposalCreated(
        uint256 indexed proposalId,
        address indexed proposer,
        address[] targets,
        uint256[] values,
        string[] signatures,
        bytes[] calldatas,
        uint256 startBlock,
        uint256 endBlock,
        string description,
        ProposalCategory indexed category
    );

    event ExtraOrdinaryAssemblyRequested(
        uint256 indexed proposalId,
        address indexed proposer,
        address[] targets,
        uint256[] values,
        string[] signatures,
        bytes[] calldatas,
        uint256 startBlock,
        uint256 endBlock,
        string description
    );

    enum ProposalCategory {
        Generic,
        ExtraordinaryVote,
        AssociationDissolution,
        ChangeBylaws
    }

    struct ProposalCore {
        TimersUpgradeable.BlockNumber voteStart;
        TimersUpgradeable.BlockNumber voteEnd;
        ProposalCategory category;
        bool executed;
        bool canceled;
    }

    string private _name;

    mapping(uint256 => ProposalCore) internal _proposals;

    uint256[] public slots;
    // BlockNumber => ProposalId[]
    mapping(uint256 => uint256[]) public votingSlots;

    // number of blocks before voting slot closes for submission
    uint256 public slotCloseTime;

    DoubleEndedQueueUpgradeable.Bytes32Deque private _governanceCall;

    event NewTimeslotSet(uint256 timeslot);

    modifier onlyGovernance() {
        require(_msgSender() == _executor(), "Governor: onlyGovernance");
        if (_executor() != address(this)) {
            bytes32 msgDataHash = keccak256(_msgData());
            // loop until popping the expected operation - throw if deque is empty (operation not authorized)
            // solhint-disable-next-line  no-empty-blocks
            while (_governanceCall.popFront() != msgDataHash) {}
        }
        _;
    }

    function governorInit(string memory name_) internal onlyInitializing {
        __EIP712_init_unchained(name_, version());
        governorInitUnchained(name_);
        slotCloseTime = 50400; // 1 week before
    }

    function governorInitUnchained(
        string memory name_
    ) internal onlyInitializing {
        _name = name_;
    }

    receive() external payable virtual {
        require(_executor() == address(this), "can only called by governor");
    }

    function supportsInterface(
        bytes4 interfaceId
    ) public view virtual override(IERC165Upgradeable) returns (bool) {
        return
            interfaceId ==
            (type(IGovernorUpgradeable).interfaceId ^
                this.castVoteWithReasonAndParams.selector ^
                this.castVoteWithReasonAndParamsBySig.selector ^
                this.getVotesWithParams.selector) ||
            interfaceId == type(IGovernorUpgradeable).interfaceId;
    }

    function name() public view virtual override returns (string memory) {
        return _name;
    }

    function version() public view virtual override returns (string memory) {
        return "1";
    }

    function hashProposal(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public pure virtual override returns (uint256) {
        return
            uint256(
                keccak256(
                    abi.encode(targets, values, calldatas, descriptionHash)
                )
            );
    }

    function state(
        uint256 proposalId
    ) public view virtual override returns (ProposalState) {
        ProposalCore storage proposal = _proposals[proposalId];

        if (proposal.executed) {
            return ProposalState.Executed;
        }

        if (proposal.canceled) {
            return ProposalState.Canceled;
        }

        uint256 proposalStart = proposalSnapshot(proposalId);

        if (proposalStart == 0) {
            revert("Governor: unknown proposal id");
        }

        if (proposalStart >= block.number) {
            return ProposalState.Pending;
        }

        uint256 deadline = proposalDeadline(proposalId);

        if (deadline >= block.number) {
            return ProposalState.Active;
        }

        if (_quorumReached(proposalId) && _voteSucceeded(proposalId)) {
            return ProposalState.Succeeded;
        } else {
            return ProposalState.Defeated;
        }
    }

    function proposalSnapshot(
        uint256 proposalId
    ) public view virtual override returns (uint256) {
        return _proposals[proposalId].voteStart.getDeadline();
    }

    function proposalDeadline(
        uint256 proposalId
    ) public view virtual override returns (uint256) {
        return _proposals[proposalId].voteEnd.getDeadline();
    }

    /**
     * The number of votes required in order for a voter to become a proposer
     * Must have voting power to create a proposal
     */
    function proposalThreshold() public view virtual returns (uint256) {
        return 1;
    }

    function _quorumReached(
        uint256 proposalId
    ) internal view virtual returns (bool);

    function _voteSucceeded(
        uint256 proposalId
    ) internal view virtual returns (bool);

    function _getVotes(
        address account,
        uint256 blockNumber,
        bytes memory params
    ) internal view virtual returns (uint256);

    function _countVote(
        uint256 proposalId,
        address account,
        uint8 support,
        uint256 weight,
        bytes memory params
    ) internal virtual;

    function propose(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description
    ) public virtual override(IGovernorUpgradeable) returns (uint256) {
        uint256 proposalId = _checkAndHashProposal(
            targets,
            values,
            calldatas,
            description
        );

        ProposalCore storage proposal = _buildProposal(proposalId, calldatas);

        emit ProposalCreated(
            proposalId,
            _msgSender(),
            targets,
            values,
            new string[](targets.length),
            calldatas,
            proposal.voteStart._deadline,
            proposal.voteEnd._deadline,
            description
        );

        emit DAAProposalCreated(
            proposalId,
            _msgSender(),
            targets,
            values,
            new string[](targets.length),
            calldatas,
            proposal.voteStart._deadline,
            proposal.voteEnd._deadline,
            description,
            proposal.category
        );

        if (proposal.category == ProposalCategory.ExtraordinaryVote) {
            emit ExtraOrdinaryAssemblyRequested(
                proposalId,
                _msgSender(),
                targets,
                values,
                new string[](targets.length),
                calldatas,
                proposal.voteStart._deadline,
                proposal.voteEnd._deadline,
                description
            );
        }

        return proposalId;
    }

    function _checkAndHashProposal(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description
    ) internal view returns (uint256) {
        require(
            getVotes(_msgSender(), block.number - 1) >= proposalThreshold(),
            "Proposer votes below threshold"
        );

        uint256 proposalId = hashProposal(
            targets,
            values,
            calldatas,
            keccak256(bytes(description))
        );

        require(targets.length == values.length, "Invalid proposal length");
        require(targets.length == calldatas.length, "Invalid proposal length");
        require(targets.length > 0, "Empty proposal");

        return proposalId;
    }

    function _buildProposal(
        uint256 proposalId,
        bytes[] memory calldatas
    ) internal returns (ProposalCore storage) {
        ProposalCore storage proposal = _proposals[proposalId];
        require(proposal.voteStart.isUnset(), "Proposal already exists");

        bool isRequestingExtraOrdinaryVotingSlot = false;
        for (uint256 i = 0; i < calldatas.length; i++) {
            bytes4 functionSignature = bytes4(calldatas[i]);
            if (functionSignature == setVotingSlotSignature) {
                isRequestingExtraOrdinaryVotingSlot = true;
                break;
            }
        }

        if (isRequestingExtraOrdinaryVotingSlot) {
            uint64 start = block.number.toUint64() + votingDelay().toUint64();
            uint64 end = start + votingPeriod().toUint64();

            proposal.voteStart.setDeadline(start);
            proposal.voteEnd.setDeadline(end);
            proposal.category = ProposalCategory.ExtraordinaryVote;
        } else {
            uint256 nextSlot = _getNextPossibleVotingSlot();

            uint64 start = nextSlot.toUint64();
            uint64 end = start + votingPeriod().toUint64();

            proposal.voteStart.setDeadline(start);
            proposal.voteEnd.setDeadline(end);
            proposal.category = ProposalCategory.Generic;

            votingSlots[nextSlot].push(proposalId);
        }

        return proposal;
    }

    function execute(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public payable virtual override returns (uint256) {
        uint256 proposalId = hashProposal(
            targets,
            values,
            calldatas,
            descriptionHash
        );

        ProposalState status = state(proposalId);
        require(
            status == ProposalState.Succeeded || status == ProposalState.Queued,
            "Proposal not successful"
        );
        _proposals[proposalId].executed = true;

        emit ProposalExecuted(proposalId);

        _beforeExecute(proposalId, targets, values, calldatas, descriptionHash);
        _execute(proposalId, targets, values, calldatas, descriptionHash);
        _afterExecute(proposalId, targets, values, calldatas, descriptionHash);

        return proposalId;
    }

    function _execute(
        uint256 /* proposalId */,
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 /*descriptionHash*/
    ) internal virtual {
        string memory errorMessage = "Call reverted without message";
        for (uint256 i = 0; i < targets.length; ++i) {
            // solhint-disable-next-line avoid-low-level-calls
            (bool success, bytes memory returndata) = targets[i].call{
                value: values[i]
            }(calldatas[i]);
            AddressUpgradeable.verifyCallResult(
                success,
                returndata,
                errorMessage
            );
        }
    }

    function _beforeExecute(
        uint256 /* proposalId */,
        address[] memory targets,
        uint256[] memory /* values */,
        bytes[] memory calldatas,
        bytes32 /*descriptionHash*/
    ) internal virtual {
        if (_executor() != address(this)) {
            for (uint256 i = 0; i < targets.length; ++i) {
                if (targets[i] == address(this)) {
                    _governanceCall.pushBack(keccak256(calldatas[i]));
                }
            }
        }
    }

    function _afterExecute(
        uint256 /* proposalId */,
        address[] memory /* targets */,
        uint256[] memory /* values */,
        bytes[] memory /* calldatas */,
        bytes32 /*descriptionHash*/
    ) internal virtual {
        if (_executor() != address(this)) {
            if (!_governanceCall.empty()) {
                _governanceCall.clear();
            }
        }
    }

    function _cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) internal virtual returns (uint256) {
        uint256 proposalId = hashProposal(
            targets,
            values,
            calldatas,
            descriptionHash
        );
        ProposalState status = state(proposalId);

        require(
            status != ProposalState.Canceled &&
                status != ProposalState.Expired &&
                status != ProposalState.Executed,
            "Governor: proposal not active"
        );
        _proposals[proposalId].canceled = true;

        emit ProposalCanceled(proposalId);

        return proposalId;
    }

    function getVotes(
        address account,
        uint256 blockNumber
    ) public view virtual override returns (uint256) {
        return _getVotes(account, blockNumber, "");
    }

    function getVotesWithParams(
        address account,
        uint256 blockNumber,
        bytes memory params
    ) public view virtual override returns (uint256) {
        return _getVotes(account, blockNumber, params);
    }

    function castVote(
        uint256 proposalId,
        uint8 support
    ) public virtual override returns (uint256) {
        address voter = _msgSender();
        return _castVote(proposalId, voter, support, "");
    }

    function castVoteWithReason(
        uint256 proposalId,
        uint8 support,
        string calldata reason
    ) public virtual override returns (uint256) {
        address voter = _msgSender();
        return _castVote(proposalId, voter, support, reason);
    }

    function castVoteWithReasonAndParams(
        uint256 proposalId,
        uint8 support,
        string calldata reason,
        bytes memory params
    ) public virtual override returns (uint256) {
        address voter = _msgSender();
        return _castVote(proposalId, voter, support, reason, params);
    }

    /* solhint-disable no-unused-vars */
    function castVoteBySig(
        uint256 proposalId,
        uint8 support,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public virtual override returns (uint256) {
        // The law says you always have to vote yourself
        require(1 == 0, "not possible");
        return 0;
    }

    function castVoteWithReasonAndParamsBySig(
        uint256 proposalId,
        uint8 support,
        string calldata reason,
        bytes memory params,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public virtual override returns (uint256) {
        // The law says you always have to vote yourself
        require(1 == 0, "not possible");
        return 0;
    }

    /* solhint-enable no-unused-vars */

    function _castVote(
        uint256 proposalId,
        address account,
        uint8 support,
        string memory reason
    ) internal virtual returns (uint256) {
        return _castVote(proposalId, account, support, reason, "");
    }

    function _castVote(
        uint256 proposalId,
        address account,
        uint8 support,
        string memory reason,
        bytes memory params
    ) internal virtual returns (uint256) {
        ProposalCore storage proposal = _proposals[proposalId];
        require(
            state(proposalId) == ProposalState.Active,
            "Vote not currently active"
        );

        uint256 weight = _getVotes(
            account,
            proposal.voteStart.getDeadline(),
            params
        );
        require(weight > 0, "no voting rights");
        _countVote(proposalId, account, support, weight, params);

        if (params.length == 0) {
            emit VoteCast(account, proposalId, support, weight, reason);
        } else {
            emit VoteCastWithParams(
                account,
                proposalId,
                support,
                weight,
                reason,
                params
            );
        }

        return weight;
    }

    function relay(
        address target,
        uint256 value,
        bytes calldata data
    ) external payable virtual onlyGovernance {
        // solhint-disable-next-line avoid-low-level-calls
        (bool success, bytes memory returndata) = target.call{value: value}(
            data
        );
        AddressUpgradeable.verifyCallResult(
            success,
            returndata,
            "Relay reverted without message"
        );
    }

    function _executor() internal view virtual returns (address) {
        return address(this);
    }

    function _getNextPossibleVotingSlot() internal view returns (uint256) {
        for (uint256 i = 0; i < slots.length; i++) {
            if (block.number < (slots[i] - slotCloseTime)) {
                return slots[i];
            }
        }
        revert("No voting slot found");
    }

    function getSlotsLength() external view returns (uint256) {
        return slots.length;
    }

    function getNumberOfProposalsInVotingSlot(
        uint256 slotNumber
    ) external view returns (uint256) {
        return votingSlots[slotNumber].length;
    }

    uint256[46] private __gap;
}
