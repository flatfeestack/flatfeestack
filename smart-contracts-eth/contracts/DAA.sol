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

    string public bylawsHash;
    bool private _foundingSetupDone;

    event BylawsChanged(string indexed oldHash, string indexed newHash);

    function initialize(
        Membership _membership,
        TimelockControllerUpgradeable _timelock,
        string memory bylaws
    ) public initializer {
        membershipContract = Membership(_membership);
        _foundingSetupDone = false;
        extraOrdinaryAssemblyVotingPeriod = 50400;

        governorInit("FlatFeeStack");
        governorVotesInit(_membership);
        governorCountingSimpleInit();
        governorVotesQuorumFractionInit(5);
        governorTimelockControlInit(_timelock);
        setupDAAFoundingSlotAndProposal(bylaws);
    }

    function votingDelay() public pure override returns (uint256) {
        return 0;
        // Votes get assigned to slots, so delay is differs every time
    }

    function votingPeriod() public pure override returns (uint256) {
        return 7200;
        // 1 day in blocks
    }

    function proposalThreshold() public pure override returns (uint256) {
        return 1;
    }

    function quorum(
        uint256 blockNumber
    )
        public
        view
        override(IGovernorUpgradeable, GovernorVotesQuorumFractionUpgradeable)
        returns (uint256)
    {
        return super.quorum(blockNumber);
    }

    function getVotes(
        address account,
        uint256 blockNumber
    )
        public
        view
        override(GovernorUpgradeable, IGovernorUpgradeable)
        returns (uint256)
    {
        return super.getVotes(account, blockNumber);
    }

    function state(
        uint256 proposalId
    )
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
        require(
            membershipContract.isCouncilMember(msg.sender) ||
                _msgSender() == _executor(),
            "only council member or governor"
        );

        require(
            blockNumber >= block.number + 201600,
            "Must be a least a month from now"
        );

        uint256 previousMaxIndex = slots.length - 1;

        for (uint256 i = previousMaxIndex; i >= 0; i--) {
            if (slots[i] == blockNumber) {
                revert("Vote slot already exists");
            }

            if (i == 0) {
                // prevent underflow
                break;
            }
        }

        uint256 targetIndex = 0;
        for (uint256 i = previousMaxIndex; i >= 0; i--) {
            if (slots[i] < blockNumber) {
                targetIndex = i + 1;
                break;
            }

            if (i == 0) {
                // prevent underflow
                break;
            }
        }

        slots.push(blockNumber);
        if (targetIndex < slots.length - 1) {
            for (uint256 i = previousMaxIndex; i >= targetIndex; i--) {
                slots[i + 1] = slots[i];

                if (i == 0) {
                    // prevent underflow
                    break;
                }
            }
        }

        slots[targetIndex] = blockNumber;

        emit NewTimeslotSet(blockNumber);
        return blockNumber;
    }

    function cancelVotingSlot(
        uint256 blockNumber,
        string calldata reason
    ) public {
        require(
            membershipContract.isCouncilMember(msg.sender),
            "only council member"
        );
        require(
            blockNumber >= block.number + 7200,
            "Must be a day before slot!"
        );

        uint256 index;
        bool slotExists = false;

        for (index = 0; index < slots.length; index++) {
            if (slots[index] == blockNumber) {
                slotExists = true;
                break;
            }
        }

        if (!slotExists) {
            revert("Voting slot does not exist!");
        }

        for (uint256 i = index; i < slots.length - 1; i++) {
            slots[i] = slots[i + 1];
        }
        slots.pop();

        uint256[] memory proposalIds = votingSlots[blockNumber];

        delete votingSlots[blockNumber];
        uint256 nextSlot = _getNextPossibleVotingSlot();

        for (uint256 j = 0; j < proposalIds.length; j++) {
            ProposalCore storage proposal = _proposals[proposalIds[j]];

            uint64 oldStart = proposal.voteStart.getDeadline();
            uint64 start = nextSlot.toUint64();
            uint64 end = start + votingPeriod().toUint64();

            proposal.voteStart.setDeadline(start);
            proposal.voteEnd.setDeadline(end);

            votingSlots[nextSlot].push(proposalIds[j]);

            emit ProposalVotingTimeChanged(
                proposalIds[j],
                oldStart,
                proposal.voteStart.getDeadline()
            );
        }

        emit VotingSlotCancelled(blockNumber, reason);
    }

    function supportsInterface(
        bytes4 interfaceId
    )
        public
        view
        override(GovernorUpgradeable, GovernorTimelockControlUpgradeable)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }

    function setNewBylawsHash(string memory newHash) external onlyGovernance {
        string memory oldHash = bylawsHash;
        bylawsHash = newHash;
        emit BylawsChanged(oldHash, bylawsHash);
    }

    function setupDAAFoundingSlotAndProposal(string memory bylaws) internal {
        require(_foundingSetupDone == false, "already done");

        // Create slot
        uint256 slotBlockNumber = block.number + slotCloseTime + 1;
        // First slot is in a week
        slots.push(slotBlockNumber);
        emit NewTimeslotSet(slotBlockNumber);

        // CreateProposal
        bytes memory calldatas = abi.encodeCall(DAA.setNewBylawsHash, bylaws);
        string memory description = "Founding Proposal. Set initial bylaws.";
        address[] memory targets = new address[](1);
        targets[0] = address(this);

        uint256[] memory values = new uint256[](1);
        values[0] = 0;

        bytes[] memory calldatasArray = new bytes[](1);
        calldatasArray[0] = calldatas;

        uint256 proposalId = hashProposal(
            targets,
            values,
            calldatasArray,
            keccak256(bytes(description))
        );

        ProposalCore storage proposal = _buildProposal(
            proposalId,
            calldatasArray
        );

        emit ProposalCreated(
            proposalId,
            _msgSender(),
            targets,
            values,
            new string[](targets.length),
            calldatasArray,
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
            calldatasArray,
            proposal.voteStart._deadline,
            proposal.voteEnd._deadline,
            description,
            proposal.category
        );

        _foundingSetupDone = true;
    }

    function setSlotCloseTime(
        uint256 newSlotCloseTime
    ) external onlyGovernance {
        slotCloseTime = newSlotCloseTime;
    }

    function setExtraOrdinaryAssemblyVotingPeriod(
        uint64 newExtraOrdinaryAssemblyVotingPeriod
    ) external onlyGovernance {
        extraOrdinaryAssemblyVotingPeriod = newExtraOrdinaryAssemblyVotingPeriod;
    }

    // this overs the case that an extraordinary vote needs 20% of all members to participate
    function _quorumReached(
        uint256 proposalId
    )
        internal
        view
        virtual
        override(GovernorCountingSimpleUpgradeable, GovernorUpgradeable)
        returns (bool)
    {
        ProposalCategory proposalCategory = _proposals[proposalId].category;
        if (proposalCategory != ProposalCategory.ExtraordinaryVote) {
            return super._quorumReached(proposalId);
        }

        ProposalVote storage proposalVote = _proposalVotes[proposalId];

        uint256 voteStart = proposalSnapshot(proposalId);
        uint256 neededQuorum = (token.getPastTotalSupply(voteStart) * 20) /
            quorumDenominator();
        return
            neededQuorum <= proposalVote.forVotes + proposalVote.abstainVotes;
    }
}
