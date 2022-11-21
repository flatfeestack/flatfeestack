// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts (last updated v4.6.0) (governance/extensions/GovernorTimelockControl.sol)

pragma solidity ^0.8.17;

import "./GovernorUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/TimelockControllerUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/extensions/IGovernorTimelockUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

abstract contract GovernorTimelockControlUpgradeable is
    Initializable,
    IGovernorTimelockUpgradeable,
    GovernorUpgradeable
{
    TimelockControllerUpgradeable private _timelock;
    mapping(uint256 => bytes32) private _timelockIds;

    event TimelockChange(address oldTimelock, address newTimelock);

    function governorTimelockControlInit(
        TimelockControllerUpgradeable timelockAddress
    ) internal onlyInitializing {
        governorTimelockControlInitUnchained(timelockAddress);
    }

    function governorTimelockControlInitUnchained(
        TimelockControllerUpgradeable timelockAddress
    ) internal onlyInitializing {
        _updateTimelock(timelockAddress);
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        virtual
        override(IERC165Upgradeable, GovernorUpgradeable)
        returns (bool)
    {
        return
            interfaceId == type(IGovernorTimelockUpgradeable).interfaceId ||
            super.supportsInterface(interfaceId);
    }

    function state(uint256 proposalId)
        public
        view
        virtual
        override(IGovernorUpgradeable, GovernorUpgradeable)
        returns (ProposalState)
    {
        ProposalState status = super.state(proposalId);

        if (status != ProposalState.Succeeded) {
            return status;
        }

        // core tracks execution, so we just have to check if successful proposal have been queued.
        bytes32 queueid = _timelockIds[proposalId];
        if (queueid == bytes32(0)) {
            return status;
        } else if (_timelock.isOperationDone(queueid)) {
            return ProposalState.Executed;
        } else if (_timelock.isOperationPending(queueid)) {
            return ProposalState.Queued;
        } else {
            return ProposalState.Canceled;
        }
    }

    function timelock() public view virtual override returns (address) {
        return address(_timelock);
    }

    function proposalEta(uint256 proposalId)
        public
        view
        virtual
        override
        returns (uint256)
    {
        uint256 eta = _timelock.getTimestamp(_timelockIds[proposalId]);
        return eta == 1 ? 0 : eta; // _DONE_TIMESTAMP (1) should be replaced with a 0 value
    }

    function queue(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public virtual override returns (uint256) {
        uint256 proposalId = hashProposal(
            targets,
            values,
            calldatas,
            descriptionHash
        );

        require(
            state(proposalId) == ProposalState.Succeeded,
            "Proposal not successful"
        );

        uint256 delay = _timelock.getMinDelay();
        _timelockIds[proposalId] = _timelock.hashOperationBatch(
            targets,
            values,
            calldatas,
            0,
            descriptionHash
        );
        _timelock.scheduleBatch(
            targets,
            values,
            calldatas,
            0,
            descriptionHash,
            delay
        );

        // solhint-disable-next-line not-rely-on-time
        emit ProposalQueued(proposalId, block.timestamp + delay);

        return proposalId;
    }

    function _execute(
        uint256, /* proposalId */
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) internal virtual override {
        _timelock.executeBatch{value: msg.value}(
            targets,
            values,
            calldatas,
            0,
            descriptionHash
        );
    }

    function _cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) internal virtual override returns (uint256) {
        uint256 proposalId = super._cancel(
            targets,
            values,
            calldatas,
            descriptionHash
        );

        if (_timelockIds[proposalId] != 0) {
            _timelock.cancel(_timelockIds[proposalId]);
            delete _timelockIds[proposalId];
        }

        return proposalId;
    }

    function _executor() internal view virtual override returns (address) {
        return address(_timelock);
    }

    function updateTimelock(TimelockControllerUpgradeable newTimelock)
        external
        virtual
        onlyGovernance
    {
        _updateTimelock(newTimelock);
    }

    function _updateTimelock(TimelockControllerUpgradeable newTimelock)
        private
    {
        emit TimelockChange(address(_timelock), address(newTimelock));
        _timelock = newTimelock;
    }

    uint256[48] private __gap;
}
