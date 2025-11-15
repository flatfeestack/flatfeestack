// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

// This forces Hardhat to compile all ERC-4337 contracts
import "@account-abstraction/contracts/core/EntryPoint.sol";
import "@account-abstraction/contracts/core/EntryPointSimulations.sol";
import "@account-abstraction/contracts/core/BasePaymaster.sol";
import "@account-abstraction/contracts/interfaces/IEntryPoint.sol";
import "@account-abstraction/contracts/samples/SimpleAccount.sol";
import "@account-abstraction/contracts/samples/SimpleAccountFactory.sol";
