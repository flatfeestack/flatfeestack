// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/*
 * Used for testing only
 */
contract USDC is ERC20 {
    constructor() ERC20("USDC", "USDC") {
        _mint(msg.sender, 100000 * (10 ** decimals()));
    }
    function decimals() public pure override returns (uint8) {
        return 6;
    }
}
