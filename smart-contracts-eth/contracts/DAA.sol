// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/utils/math/SafeCastUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/TimersUpgradeable.sol";

import "./governor/GovernorUpgradeable.sol";
import "./governor/GovernorVotesUpgradeable.sol";
import "./governor/GovernorCountingSimpleUpgradeable.sol";
import "./governor/GovernorVotesQuorumFractionUpgradeable.sol";
import "./governor/GovernorTimelockControlUpgradeable.sol";
import "./Membership.sol";

contract DAA is
    Initializable,
    GovernorUpgradeable,
    GovernorVotesUpgradeable,
    GovernorCountingSimpleUpgradeable,
    GovernorVotesQuorumFractionUpgradeable,
    GovernorTimelockControlUpgradeable
{
    Membership public membershipContract;

    using SafeCastUpgradeable for uint256;
    using TimersUpgradeable for TimersUpgradeable.BlockNumber;

    event ProposalVotingTimeChanged(
        uint256 proposalId,
        uint64 oldTime,
        uint64 newTime
    );

    event VotingSlotCancelled(uint256 blockNumber, string reason);

    function initialize(
        Membership _membership,
        TimelockControllerUpgradeable _timelock
    ) public initializer {
        membershipContract = Membership(_membership);

        governorInit("FlatFeeStack");
        governorVotesInit(_membership);
        governorCountingSimpleInit();
        governorVotesQuorumFractionInit(0);
        governorTimelockControlInit(_timelock);
    }

    function votingDelay() public pure override returns (uint256) {
        return 0; // Votes get assigned to slots, so delay is differt ervery time
    }

    function votingPeriod() public pure override returns (uint256) {
        return 7200; // 1 day in blocks
    }

    function proposalThreshold() public pure override returns (uint256) {
        return 1;
    }

    function quorum(uint256 blockNumber)
        public
        view
        override(IGovernorUpgradeable, GovernorVotesQuorumFractionUpgradeable)
        returns (uint256)
    {
        return super.quorum(blockNumber);
    }

    function getVotes(address account, uint256 blockNumber)
        public
        view
        override(GovernorUpgradeable, IGovernorUpgradeable)
        returns (uint256)
    {
        return super.getVotes(account, blockNumber);
    }

    function state(uint256 proposalId)
        public
        view
        override(GovernorUpgradeable, GovernorTimelockControlUpgradeable)
        returns (ProposalState)
    {
        return super.state(proposalId);
    }

    function propose(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description
    )
        public
        override(GovernorUpgradeable, IGovernorUpgradeable)
        returns (uint256)
    {
        return super.propose(targets, values, calldatas, description);
    }

    function _execute(
        uint256 proposalId,
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    )
        internal
        override(GovernorUpgradeable, GovernorTimelockControlUpgradeable)
    {
        super._execute(proposalId, targets, values, calldatas, descriptionHash);
    }

    function _cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    )
        internal
        override(GovernorUpgradeable, GovernorTimelockControlUpgradeable)
        returns (uint256)
    {
        return super._cancel(targets, values, calldatas, descriptionHash);
    }

    function _executor()
        internal
        view
        override(GovernorUpgradeable, GovernorTimelockControlUpgradeable)
        returns (address)
    {
        return super._executor();
    }

    // Sets a new voting slot
    // the voting slot has to be four weeks from now
    // it is calculated in blocks and we assume that 7200 blocks will be mined in a day
    function setVotingSlot(uint256 blockNumber) public returns (uint256) {
        require(msg.sender == membershipContract.chairman(), "only chairman");
        require(
            blockNumber >= block.number + 201600,
            "Must be a least a month from now"
        );
        for (uint256 i = 0; i < slots.length; i++) {
            if (slots[i] == blockNumber) {
                revert("Vote slot already exists");
            }
        }
        slots.push(blockNumber);
        emit NewTimeslotSet(blockNumber);
        return blockNumber;
    }

    function cancelVotingSlot(uint256 blockNumber, string calldata reason)
        public
    {
        require(msg.sender == membershipContract.chairman(), "only chairman");
        require(
            blockNumber >= block.number + 7200,
            "Must be a day before slot!"
        );

        uint256 i;
        bool slotExists = false;

        for (i = 0; i < slots.length; i++) {
            if (slots[i] == blockNumber) {
                slotExists = true;
                break;
            }
        }

        if (!slotExists) {
            revert("Voting slot does not exist!");
        }

        if (i != slots.length - 1) {
            slots[i] = slots[slots.length - 1];
        }
        slots.pop();

        uint256[] memory proposalIds = votingSlots[blockNumber];

        delete votingSlots[blockNumber];
        uint256 nextSlot = _getNextPossibleVotingSlot();

        for (i = 0; i < proposalIds.length; i++) {
            ProposalCore storage proposal = _proposals[proposalIds[i]];

            uint64 oldStart = proposal.voteStart.getDeadline();
            uint64 start = nextSlot.toUint64();
            uint64 end = start + votingPeriod().toUint64();

            proposal.voteStart.setDeadline(start);
            proposal.voteEnd.setDeadline(end);

            votingSlots[nextSlot].push(proposalIds[i]);

            emit ProposalVotingTimeChanged(
                proposalIds[i],
                oldStart,
                proposal.voteStart.getDeadline()
            );
        }

        emit VotingSlotCancelled(blockNumber, reason);
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(GovernorUpgradeable, GovernorTimelockControlUpgradeable)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }
}
