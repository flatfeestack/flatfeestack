// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts (last updated v4.8.0-rc.1) (governance/extensions/GovernorVotesQuorumFraction.sol)

pragma solidity ^0.8.17;

import "./GovernorVotesUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/CheckpointsUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/math/SafeCastUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

abstract contract GovernorVotesQuorumFractionUpgradeable is
    Initializable,
    GovernorVotesUpgradeable
{
    using CheckpointsUpgradeable for CheckpointsUpgradeable.History;

    uint256 private _quorumNumerator; // DEPRECATED
    CheckpointsUpgradeable.History private _quorumNumeratorHistory;

    event QuorumNumeratorUpdated(
        uint256 oldQuorumNumerator,
        uint256 newQuorumNumerator
    );

    function governorVotesQuorumFractionInit(
        uint256 quorumNumeratorValue
    ) internal onlyInitializing {
        governorVotesQuorumFractionInitUnchained(quorumNumeratorValue);
    }

    function governorVotesQuorumFractionInitUnchained(
        uint256 quorumNumeratorValue
    ) internal onlyInitializing {
        _updateQuorumNumerator(quorumNumeratorValue);
    }

    function quorumNumerator() public view virtual returns (uint256) {
        return
            _quorumNumeratorHistory._checkpoints.length == 0
                ? _quorumNumerator
                : _quorumNumeratorHistory.latest();
    }

    function quorumNumerator(
        uint256 blockNumber
    ) public view virtual returns (uint256) {
        // If history is empty, fallback to old storage
        uint256 length = _quorumNumeratorHistory._checkpoints.length;
        if (length == 0) {
            return _quorumNumerator;
        }

        // Optimistic search, check the latest checkpoint
        CheckpointsUpgradeable.Checkpoint
            memory latest = _quorumNumeratorHistory._checkpoints[length - 1];
        if (latest._blockNumber <= blockNumber) {
            return latest._value;
        }

        // Otherwise, do the binary search
        return _quorumNumeratorHistory.getAtBlock(blockNumber);
    }

    function quorumDenominator() public view virtual returns (uint256) {
        return 100;
    }

    function quorum(
        uint256 blockNumber
    ) public view virtual override returns (uint256) {
        return
            (token.getPastTotalSupply(blockNumber) *
                quorumNumerator(blockNumber)) / quorumDenominator();
    }

    function updateQuorumNumerator(
        uint256 newQuorumNumerator
    ) external virtual onlyGovernance {
        _updateQuorumNumerator(newQuorumNumerator);
    }

    function _updateQuorumNumerator(
        uint256 newQuorumNumerator
    ) internal virtual {
        require(
            newQuorumNumerator <= quorumDenominator(),
            "Numerator over denominator"
        );

        uint256 oldQuorumNumerator = quorumNumerator();

        // Make sure we keep track of the original numerator in contracts upgraded from a version without checkpoints.
        if (
            oldQuorumNumerator != 0 &&
            _quorumNumeratorHistory._checkpoints.length == 0
        ) {
            _quorumNumeratorHistory._checkpoints.push(
                CheckpointsUpgradeable.Checkpoint({
                    _blockNumber: 0,
                    _value: SafeCastUpgradeable.toUint224(oldQuorumNumerator)
                })
            );
        }

        // Set new quorum for future proposals
        _quorumNumeratorHistory.push(newQuorumNumerator);

        emit QuorumNumeratorUpdated(oldQuorumNumerator, newQuorumNumerator);
    }

    uint256[50] private __gap;
}
