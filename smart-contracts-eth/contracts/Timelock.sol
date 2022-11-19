// SPDX-License-Identifier: MIT

pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/governance/TimelockControllerUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract Timelock is Initializable, TimelockControllerUpgradeable {
    function initialize(address _admin) public initializer {
        address[] memory emptyArray;

        // allow everybody to execute proposals
        address[] memory nullAddress = new address[](1);
        nullAddress[0] = address(0);

        __TimelockController_init(86400, emptyArray, nullAddress, _admin);
    }
}
