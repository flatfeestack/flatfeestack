// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

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

interface IAccount {
    function validateUserOp(UserOperation calldata userOp, bytes32 userOpHash, uint256 missingAccountFunds)
        external returns (uint256 validationData);
}

interface IPaymaster {
    enum PostOpMode { opSucceeded, opReverted, postOpReverted }
    function validatePaymasterUserOp(UserOperation calldata, bytes32, uint256)
        external returns (bytes memory, uint256);
    function postOp(PostOpMode mode, bytes calldata context, uint256 actualGasCost) external;
}

contract TestEntryPoint {
    mapping(address => uint256) public deposits;

    function depositTo(address account) external payable {
        deposits[account] += msg.value;
    }
    function getDeposit(address account) external view returns (uint256) {
        return deposits[account];
    }
    function withdrawTo(address payable withdrawAddress, uint256 amount) external {
        require(deposits[msg.sender] >= amount, "not enough");
        deposits[msg.sender] -= amount;
        (bool ok,) = withdrawAddress.call{value: amount}("");
        require(ok, "withdraw failed");
    }

    function getUserOpHash(UserOperation calldata userOp) public pure returns (bytes32) {
        return keccak256(abi.encode(
            userOp.sender,
            userOp.nonce,
            keccak256(userOp.initCode),
            keccak256(userOp.callData),
            userOp.callGasLimit,
            userOp.verificationGasLimit,
            userOp.preVerificationGas,
            userOp.maxFeePerGas,
            userOp.maxPriorityFeePerGas,
            keccak256(userOp.paymasterAndData)
        ));
    }

    function handleOps(UserOperation[] calldata ops, address payable beneficiary) external {
        for (uint i=0; i<ops.length; i++) {
            UserOperation calldata op = ops[i];

            IAccount acc = IAccount(op.sender);
            acc.validateUserOp(op, getUserOpHash(op), 0);

            bytes memory context;
            address paymaster;
            if (op.paymasterAndData.length >= 20) {
                paymaster = address(bytes20(op.paymasterAndData[:20]));
                (context,) = IPaymaster(paymaster).validatePaymasterUserOp(
                    op,
                    getUserOpHash(op),
                    op.maxFeePerGas * (op.callGasLimit + op.verificationGasLimit + op.preVerificationGas)
                );
            }

            (bool ok,) = op.sender.call(abi.encodeWithSignature("execute(bytes)", op.callData));
            require(ok, "exec failed");

            if (paymaster != address(0)) {
                _charge(paymaster, 1 ether / 1000);
                IPaymaster(paymaster).postOp(IPaymaster.PostOpMode.opSucceeded, context, 0);
            }
        }

        beneficiary;
    }

    function _charge(address paymaster, uint256 amount) internal {
        require(deposits[paymaster] >= amount, "pm deposit low");
        deposits[paymaster] -= amount;
    }
}
