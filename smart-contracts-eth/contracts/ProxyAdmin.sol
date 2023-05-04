// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";

contract MyProxyAdmin is ProxyAdmin {
    constructor(address owner) ProxyAdmin() {
        // We just need this for our hardhat tooling right now
    }
}
