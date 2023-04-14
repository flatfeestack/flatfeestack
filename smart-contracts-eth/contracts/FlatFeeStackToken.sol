// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import {ERC20Upgradeable} from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";

contract FlatFeeStackToken is ERC20Upgradeable {
    function initialize() public initializer {
        __ERC20_init("FlatFeeStackToken", "FFST");
        _mint(msg.sender, 1000 * 10 ** decimals());
    }
}
