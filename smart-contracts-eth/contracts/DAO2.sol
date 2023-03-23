// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/governance/Governor.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorCountingSimple.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotes.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotesQuorumFraction.sol";
import "./SBT.sol";

contract FlatFeeStackDAO is Governor, GovernorSettings, GovernorCountingSimple, GovernorVotes, GovernorVotesQuorumFraction {

    string public bylawsHash;

    constructor(IVotes _token)
        Governor("FFSDAO")
        GovernorSettings(2 days, 1 days, 1)
        GovernorVotes(_token)
        GovernorVotesQuorumFraction(25) {}

    // The following functions are overrides required by Solidity.

    function votingDelay()
        public
        view
        override(IGovernor, GovernorSettings)
        returns (uint256) {
        //slot is each 14 days, and you need to submit votingDelay() in advance.
        uint256 nextSlot = ((block.timestamp + super.votingDelay()) / 60 / 60 / 24 / 14) + 1;
        return nextSlot * 60 * 60 * 24 * 14;
    }

    function votingPeriod()
        public
        view
        override(IGovernor, GovernorSettings)
        returns (uint256) {
        return super.votingPeriod();
    }

    function quorum(uint256 blockNumber)
        public
        view
        override(IGovernor, GovernorVotesQuorumFraction)
        returns (uint256) {
        return super.quorum(blockNumber);
    }

    function proposalThreshold()
        public
        view
        override(Governor, GovernorSettings)
        returns (uint256) {
        return super.proposalThreshold();
    }

    function execute(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public payable override returns (uint256 proposalId) {
        uint256 proposalId0 = hashProposal(targets, values, calldatas, descriptionHash);
        //timelock is votingDelay, so we have before voting the same delay as the timelock
        require(proposalDeadline(proposalId0) + super.votingDelay() < block.timestamp, "Governor: timelock not expired yet");
        return super.execute(targets, values, calldatas, descriptionHash);
    }

    function cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash, uint8 v1, bytes32 r1, bytes32 s1, uint8 v2, bytes32 r2, bytes32 s2
    ) public virtual override returns (uint256) {
        uint256 proposalId = hashProposal(targets, values, calldatas, descriptionHash);
        require(state(proposalId) == ProposalState.Pending, "Governor: too late to cancel");

        boolean isCouncil = SBT(token).isCouncil(ecrecover(keccak256(abi.encodePacked(to, "#", timestamp)), v1, r1, s1))
            && SBT(token).isCouncil(ecrecover(keccak256(abi.encodePacked(to, "#", timestamp)), v2, r2, s2));

        require(_msgSender() == _proposals[proposalId].proposer || isCouncil, "Governor: only proposer can cancel");
        return _cancel(targets, values, calldatas, descriptionHash);
    }

    function clock() public view virtual returns (uint48) {
        return SafeCast.toUint48(block.timestamp);
    }

    /**
     * @dev Machine-readable description of the clock as specified in EIP-6372.
     */
    // solhint-disable-next-line func-name-mixedcase
    function CLOCK_MODE() public view virtual returns (string memory) {
        // Check that the clock was not modified
        // https://eips.ethereum.org/EIPS/eip-6372
        require(clock() == block.timestamp);
        return "mode=timestamp";
    }

    function setNewBylawsHash(string memory newHash) external onlyGovernance {
        string memory oldHash = bylawsHash;
        bylawsHash = newHash;
        emit BylawsChanged(oldHash, bylawsHash);
    }
}
