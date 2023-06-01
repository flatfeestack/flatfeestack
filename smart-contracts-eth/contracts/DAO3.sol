// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/governance/GovernorUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/extensions/GovernorSettingsUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/extensions/GovernorCountingSimpleUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/extensions/GovernorVotesUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/extensions/GovernorVotesQuorumFractionUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/SBT2.sol";

contract FlatFeeStackDAO is
    Initializable,
    GovernorUpgradeable,
    GovernorSettingsUpgradeable,
    GovernorCountingSimpleUpgradeable,
    GovernorVotesUpgradeable,
    GovernorVotesQuorumFractionUpgradeable
{
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    string public bylawsHash;
    event BylawsChanged(string indexed oldHash, string indexed newHash);
    mapping(uint256 => bool) councilAction;

    function initialize(IVotesUpgradeable _token) public initializer {
        __Governor_init("FlatFeeStackDAO");
        __GovernorSettings_init(14 days, 1 days, 1);
        __GovernorCountingSimple_init();
        __GovernorVotes_init(_token);
        __GovernorVotesQuorumFraction_init(5);
    }

    function votingDelay()
        public
        view
        override(IGovernorUpgradeable, GovernorSettingsUpgradeable)
        returns (uint256)
    {
        //slot is each 14 days, and you need to submit votingDelay() up until (2 * votingDelay() - 1)  in advance.
        uint256 nextSlot = ((block.timestamp + super.votingDelay()) / super.votingDelay()) + 1;
        return nextSlot * super.votingDelay();
    }

    function votingPeriod()
        public
        view
        override(IGovernorUpgradeable, GovernorSettingsUpgradeable)
        returns (uint256)
    {
        return super.votingPeriod();
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

    function proposalThreshold()
        public
        view
        override(GovernorUpgradeable, GovernorSettingsUpgradeable)
        returns (uint256)
    {
        return super.proposalThreshold();
    }

    function execute(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public payable override returns (uint256 proposalId) {
        uint256 proposalId0 = hashProposal(
            targets,
            values,
            calldatas,
            descriptionHash
        );
        //timelock is votingDelay, so we have before voting the same delay as the timelock
        require(
            proposalDeadline(proposalId0) + super.votingDelay() <
                block.timestamp,
            "Governor: timelock not expired yet"
        );
        return super.execute(targets, values, calldatas, descriptionHash);
    }

    function councilExecute(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash,
        uint8 v2,
        bytes32 r2,
        bytes32 s2
    ) public {
        uint256 proposalId = hashProposal(
            targets,
            values,
            calldatas,
            descriptionHash
        );
        require(councilAction[proposalId] == false, "Cannot execute twice");
        councilAction[proposalId] = true;
        address council2 = ecrecover(
            keccak256(abi.encode(targets, values, calldatas, descriptionHash)),
            v2,
            r2,
            s2
        );

        require(
            FlatFeeStackDAOSBT(address(token)).isCouncil(msg.sender) &&
                FlatFeeStackDAOSBT(address(token)).isCouncil(council2) &&
                msg.sender != council2,
            "Signature not from council member"
        );

        emit ProposalExecuted(proposalId);
        _beforeExecute(proposalId, targets, values, calldatas, descriptionHash);
        _execute(proposalId, targets, values, calldatas, descriptionHash);
        _afterExecute(proposalId, targets, values, calldatas, descriptionHash);
    }

    function cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash,
        uint8 v2,
        bytes32 r2,
        bytes32 s2
    ) public virtual returns (uint256) {
        uint256 proposalId = hashProposal(
            targets,
            values,
            calldatas,
            descriptionHash
        );
        require(councilAction[proposalId] == false, "Cannot execute twice");
        councilAction[proposalId] = true;
        address council2 = ecrecover(
            keccak256(abi.encode(targets, values, calldatas, descriptionHash)),
            v2,
            r2,
            s2
        );

        require(
            FlatFeeStackDAOSBT(address(token)).isCouncil(msg.sender) &&
                FlatFeeStackDAOSBT(address(token)).isCouncil(council2) &&
                msg.sender != council2,
            "Signature not from council member"
        );

        return _cancel(targets, values, calldatas, descriptionHash);
    }

    function clock()
        public
        view
        virtual
        override(IGovernorUpgradeable, GovernorVotesUpgradeable)
        returns (uint48)
    {
        return SafeCastUpgradeable.toUint48(block.timestamp);
    }

    /**
     * @dev Machine-readable description of the clock as specified in EIP-6372.
     */
    // solhint-disable-next-line func-name-mixedcase
    function CLOCK_MODE()
        public
        view
        virtual
        override(IGovernorUpgradeable, GovernorVotesUpgradeable)
        returns (string memory)
    {
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
