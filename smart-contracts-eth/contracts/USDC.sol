// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import {ERC20Upgradeable} from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";

contract USDC is ERC20Upgradeable {
    function initialize() public initializer {
        __ERC20_init("USDC", "USDC");
        _mint(msg.sender, 1000 * 10 ** decimals());
    }

    function decimals() public pure override returns (uint8) {
        return 6;
    }
}
