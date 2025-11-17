// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@account-abstraction/contracts/interfaces/IAccount.sol";
import "@account-abstraction/contracts/interfaces/IEntryPoint.sol";
import "@account-abstraction/contracts/interfaces/PackedUserOperation.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "hardhat/console.sol";

contract SimpleAccount is IAccount, Ownable {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    IEntryPoint public immutable entryPoint = IEntryPoint(0x0000000071727De22E5E9d8BAf0edAc6f37da032);
    uint256 public nonce;

    constructor() Ownable(msg.sender) {}

    function validateUserOp(
        PackedUserOperation calldata userOp,
        bytes32 userOpHash,
        uint256 missingAccountFunds
    ) external override returns (uint256 validationData) {
        require(msg.sender == address(entryPoint), "Only EntryPoint");
        // Verify signature
        bytes32 hash = userOpHash.toEthSignedMessageHash();
        address signer = hash.recover(userOp.signature);

        console.log("userOpHash:", uint256(userOpHash));
        console.log("signer:", signer);
        console.log("owner:", owner());

        if (signer != owner()) {
            return 1; // Invalid signature
        }

        // Validate and increment nonce
        require(nonce++ == userOp.nonce, "Invalid nonce");

        // Pay EntryPoint for gas if needed
        if (missingAccountFunds > 0) {
            (bool success,) = payable(msg.sender).call{value: missingAccountFunds}("");
            require(success, "Failed to pay EntryPoint");
        }

        return 0; // Signature valid
    }

    function execute(address dest, uint256 value, bytes calldata func) external {
        require(msg.sender == address(entryPoint), "Only EntryPoint");
        (bool success,) = dest.call{value: value}(func);
        require(success, "Execution failed");
    }

    function executeBatch(address[] calldata dest, bytes[] calldata func) external {
        require(msg.sender == address(entryPoint), "Only EntryPoint");
        require(dest.length == func.length, "Length mismatch");
        for (uint256 i = 0; i < dest.length; i++) {
            (bool success,) = dest[i].call(func[i]);
            require(success, "Batch execution failed");
        }
    }

    // Receive ETH
    receive() external payable {}

    // Deposit to EntryPoint for gas
    function deposit() public payable {
        entryPoint.depositTo{value: msg.value}(address(this));
    }

    // Get deposit balance at EntryPoint
    function getDeposit() public view returns (uint256) {
        return entryPoint.balanceOf(address(this));
    }

    function withdrawTo(address payable withdrawAddress, uint256 withdrawAmount) external onlyOwner {
        // Get current deposit balance at EntryPoint
        uint256 currentDeposit = entryPoint.balanceOf(address(this));
        require(withdrawAmount <= currentDeposit, "Insufficient deposit");

        // Withdraw from EntryPoint to this account
        entryPoint.withdrawTo(withdrawAddress, withdrawAmount);
    }

    // Withdraw ETH - only owner
    function withdraw(address payable to, uint256 amount) external onlyOwner {
        (bool success,) = to.call{value: amount}("");
        require(success, "Withdraw failed");
    }
}
