// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts (last updated v4.6.0) (governance/extensions/GovernorVotes.sol)

pragma solidity ^0.8.17;

import "./GovernorUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/utils/IVotesUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

abstract contract GovernorVotesUpgradeable is
    Initializable,
    GovernorUpgradeable
{
    IVotesUpgradeable public token;

    function governorVotesInit(IVotesUpgradeable tokenAddress)
        internal
        onlyInitializing
    {
        governorVotesInitUnchained(tokenAddress);
    }

    function governorVotesInitUnchained(IVotesUpgradeable tokenAddress)
        internal
        onlyInitializing
    {
        token = tokenAddress;
    }

    function _getVotes(
        address account,
        uint256 blockNumber,
        bytes memory /*params*/
    ) internal view virtual override returns (uint256) {
        return token.getPastVotes(account, blockNumber);
    }

    uint256[50] private __gap;
}
