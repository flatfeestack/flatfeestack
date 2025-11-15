// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import "@openzeppelin/contracts/governance/Governor.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorCountingSimple.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotes.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotesQuorumFraction.sol";

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";

import {FlatFeeStackNFT} from "./FlatFeeStackNFT.sol";

contract FlatFeeStackDAO is Governor, GovernorSettings, GovernorCountingSimple, GovernorVotes, GovernorVotesQuorumFraction {

    uint256 public bylawsHash;
    mapping(uint256 => bool) public councilExecution;

    event BylawsChanged(uint256 indexed oldHash, uint256 indexed newHash);

    constructor(address nftAddress)
        Governor("FlatFeeStackDAO")
        GovernorSettings(7 days, 1 days, 1)
        GovernorVotes(IVotes(nftAddress))
        GovernorVotesQuorumFraction(20) {}

    function votingDelay() public view
        override(Governor, GovernorSettings) returns (uint256) {
        /* 
        The width of a slot is 7 days, so if a proposer proposes a vote in the middle of slot 1, 
        the delay will be set that this vote starts at end of slot 2 and beginning of slot 3. This
        gives a buffer of min 7 days, max. 14 days - 1s.

        | Slot 1 | Slot 2 | Slot 3 | Slot 4|

        Example : 1697068799 (Wed Oct 11 2023 23:59:59 GMT+0000), so the slot is: 2805 (2805.9999)
        Round up: (1697068799 + ((7 * 24 * 60 * 60) -1)) / (7 * 24 * 60 * 60) = 2806
        Round up: (1697068800 + ((7 * 24 * 60 * 60) -1)) / (7 * 24 * 60 * 60) = 2806
        Round up: (1697068801 + ((7 * 24 * 60 * 60) -1)) / (7 * 24 * 60 * 60) = 2807
        Dealy until next slot: (2807 * (7 * 24 * 60 * 60)) - 1697068799 = 604801 (7d, 1s)
        Dealy until next slot: (2807 * (7 * 24 * 60 * 60)) - 1697068800 = 604800 (7d)
        Dealy until next slot: (2808 * (7 * 24 * 60 * 60)) - 1697068801 = 604801 (13d, 23h, 59m, 59s)
        */

        uint256 nextSlot = ((block.timestamp + super.votingDelay() -1) / super.votingDelay()) + 1;
        return (nextSlot * super.votingDelay()) - block.timestamp;
    }

    function votingPeriod() public view 
        override(Governor, GovernorSettings) returns (uint256) {
        return super.votingPeriod();
    }

    function quorum(uint256 timepoint) public view
        override(Governor, GovernorVotesQuorumFraction) returns (uint256) {

        // quorum with 20% for number of yes+abstain (ya) and total voters
        // total = 2 -> q:0/1 (50%) -> 0 is same as 1, as 0 votes does not get a proposal pass, 1 does.
        // total = 3 -> q:2 (67%)
        // total = 4 -> q:2 (50%)
        // total = 5 -> q:2 (40%)
        // total = 6 -> q:2 (33%)
        // total = 7 -> q:2 (28%)
        // total = 8 -> q:2 (25%)
        // total = 9 -> q:2 (22%)
        // total = 10-> q:2 (20%)
        // total = 11-> q:2 (18%)
        // total = 12-> q:2 (17%)
        // total = 13-> q:2 (15%)
        // total = 14-> q:2 (14%)
        // total = 15-> q:3 (20%)
        // total = 16-> q:3 (19%)
        uint256 q = super.quorum(timepoint);
        //corner case: if we have a quorum of 0 or 1, but we have 3+ voters
        //make quorum of 2 mandatory
        if(q < 2 && token().getPastTotalSupply(timepoint) >= 3) {
            return 2;
        }
        return q;
    }

    function proposalThreshold() public view 
        override(Governor, GovernorSettings) returns (uint256) {
        return super.proposalThreshold();
    }

    function _queueOperations(uint256, address[] memory, uint256[] memory, bytes[] memory, bytes32) 
        internal view override returns (uint48) {
        return SafeCast.toUint48(super.votingDelay());
    }

    function requireTwoCouncil(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata calldatas,
        bytes32 descriptionHash,
        uint256 index1,
        uint256 index2,
        bytes calldata signature2
    ) internal returns (uint256 proposalId) {
        bytes32 proposalHash = keccak256(
            abi.encode(targets, values, calldatas, descriptionHash));

        bytes32 messageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", proposalHash));

        proposalId = uint256(proposalHash);
        
        require(councilExecution[proposalId] == false, "Cannot execute twice");
        councilExecution[proposalId] = true;

        address council2 = ECDSA.recover(messageHash, signature2);
        require(msg.sender != council2);

        require(
            FlatFeeStackNFT(address(token())).isCouncilIndex(msg.sender, index1) &&
            FlatFeeStackNFT(address(token())).isCouncilIndex(council2, index2),
            "No council sigs");

        return proposalId;
    }

    function councilExecute(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata calldatas,
        bytes32 descriptionHash,
        uint256 index1,
        uint256 index2,
        bytes calldata signature2
    ) external returns (uint256 proposalId) {
        proposalId = requireTwoCouncil(targets, values, calldatas, descriptionHash, index1, index2, signature2);
        _executeOperations(proposalId, targets, values, calldatas, descriptionHash);
        emit ProposalExecuted(proposalId);
        return proposalId;
    }

    function councilCancel(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata calldatas,
        bytes32 descriptionHash,
        uint256 index1,
        uint256 index2,
        bytes calldata signature2
    ) external returns (uint256 proposalId) {
        requireTwoCouncil(targets, values, calldatas, descriptionHash, index1, index2, signature2);
        proposalId = _cancel(targets, values, calldatas, descriptionHash);
        emit ProposalCanceled(proposalId);
        return proposalId;
    }


    /**
     * Sets a new hash value (newHash) of bylaws and emits an event indicating 
     * the change in bylaws hash from the old to the new value.
     */
    function setNewBylawsHash(uint256 newHash) external onlyGovernance {
        uint256 oldHash = bylawsHash;
        bylawsHash = newHash;
        emit BylawsChanged(oldHash, bylawsHash);
    }

    function clock() public view virtual override(Governor, GovernorVotes)
        returns (uint48) {
        return SafeCast.toUint48(block.timestamp);
    }

    /**
     * @dev Machine-readable description of the clock as specified in EIP-6372.
     */
    // solhint-disable-next-line func-name-mixedcase
    function CLOCK_MODE() public view virtual override(Governor, GovernorVotes)
        returns (string memory) {
        // Check that the clock was not modified
        // https://eips.ethereum.org/EIPS/eip-6372
        require(clock() == block.timestamp);
        return "mode=timestamp";
    }
}