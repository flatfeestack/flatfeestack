// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
// use for real tests
// import "@account-abstraction/contracts/interfaces/IAccount.sol";
// import "@account-abstraction/contracts/interfaces/IEntryPoint.sol";
// import "@account-abstraction/contracts/interfaces/PackedUserOperation.sol";

struct UserOperation {
    address sender;
    uint256 nonce;
    bytes initCode;
    bytes callData;
    uint256 callGasLimit;
    uint256 verificationGasLimit;
    uint256 preVerificationGas;
    uint256 maxFeePerGas;
    uint256 maxPriorityFeePerGas;
    bytes paymasterAndData;
    bytes signature;
}

interface IEntryPointLike {
    function depositTo(address account) external payable;
    function getDeposit(address account) external view returns (uint256);
    function withdrawTo(address payable withdrawAddress, uint256 amount) external;
}

contract FirstAccount is Ownable {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    address public immutable entryPoint;
    uint256 public nonce;

    constructor(address initialOwner, address entryPoint_)
        Ownable(initialOwner)
    {
        entryPoint = entryPoint_;
    }

    function execute(bytes calldata callData) external {
        require(msg.sender == entryPoint, "only EntryPoint");
        (address to, uint256 value, bytes memory data) = abi.decode(callData, (address, uint256, bytes));
        (bool ok, ) = to.call{value: value}(data);
        require(ok, "call failed");
    }

    function validateUserOp(
        UserOperation calldata userOp,
        bytes32 userOpHash,
        uint256 missingAccountFunds
    ) external returns (uint256 validationData) {
        require(msg.sender == entryPoint, "only EntryPoint");

        // Verify signature
        bytes32 hashVal = userOpHash.toEthSignedMessageHash();
        address signer = ECDSA.recover(hashVal, userOp.signature);
        
        if (signer != owner()) {
            return 1;
        }

        // Validate and increment nonce
        require(nonce++ == userOp.nonce, "Invalid nonce");

        if (missingAccountFunds > 0) {
            (bool ok, ) = payable(msg.sender).call{value: missingAccountFunds}("");
            require(ok, "failed to pay EntryPoint");
        }

        return 0;
    }

    // Receive ETH
    receive() external payable {}

    // Deposit to EntryPoint for gas
    function deposit() external payable {
        IEntryPointLike(entryPoint).depositTo{value: msg.value}(address(this));
    }

    // Get deposit balance at EntryPoint
    function getDeposit() external view returns (uint256) {
        return IEntryPointLike(entryPoint).getDeposit(address(this));
    }

    function withdrawTo(address payable withdrawAddress, uint256 amount) external onlyOwner {
        IEntryPointLike(entryPoint).withdrawTo(withdrawAddress, amount);
    }

    function withdraw(address payable to, uint256 amount) external onlyOwner {
        (bool ok, ) = to.call{value: amount}("");
        require(ok, "withdraw failed");
    }
}