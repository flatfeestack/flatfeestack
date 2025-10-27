// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import "@openzeppelin/contracts/utils/cryptography/SignatureChecker.sol";
import "@openzeppelin/contracts/utils/math/SafeCast.sol";

interface IEntryPoint {
    function depositTo(address account) external payable;
    function getDeposit(address account) external view returns (uint256);
    function withdrawTo(address payable withdrawAddress, uint256 amount) external;
}

interface IPaymaster {
    enum PostOpMode { opSucceeded, opReverted, postOpReverted }
    function validatePaymasterUserOp(
        UserOperation calldata userOp,
        bytes32 userOpHash,
        uint256 maxCost
    ) external returns (bytes memory context, uint256 validationData);
    function postOp(PostOpMode mode, bytes calldata context, uint256 actualGasCost) external;
}

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

contract Paymaster is IPaymaster, EIP712, Ownable2Step {
    using SafeCast for uint256;

    IEntryPoint public immutable entryPoint;
    address public signer; // signer for tickets

    mapping(address => bool) public targetAllowed;

    struct Ticket {
        address sender;
        address target;    // optional: 0x0 -> any
        uint256 maxCost;
        uint48  deadline;  // seconds
        uint256 chainId;
    }

    bytes32 private constant TICKET_HASH =
        keccak256("Ticket(address sender,address target,uint256 maxCost,uint48 deadline,uint256 chainId)");

    error NotEntryPoint();
    error DeadlinePassed();
    error MaxCostExceeded();
    error TargetNotAllowed();
    error InvalidSignature();

    constructor(IEntryPoint _entryPoint, address _signer)
        EIP712("Paymaster", "1")
        Ownable(msg.sender)
    {
        entryPoint = _entryPoint;
        signer = _signer;
    }

    function setSigner(address _signer) external onlyOwner {
        signer = _signer;
    }

    function setTargetAllowed(address target, bool allowed) external onlyOwner {
        targetAllowed[target] = allowed;
    }

    function deposit() external payable onlyOwner {
        entryPoint.depositTo{value: msg.value}(address(this));
    }

    function withdraw(address payable to, uint256 amount) external onlyOwner {
        entryPoint.withdrawTo(to, amount);
    }

    function entryPointDeposit() external view returns (uint256) {
        return entryPoint.getDeposit(address(this));
    }

    function validatePaymasterUserOp(
        UserOperation calldata userOp,
        bytes32, // userOpHash,
        uint256 maxCost
    ) external view override returns (bytes memory context, uint256 validationData) {
        if (msg.sender != address(entryPoint)) revert NotEntryPoint();

        (Ticket memory t, bytes memory sig) = abi.decode(userOp.paymasterAndData[20:], (Ticket, bytes));

        if (t.deadline < block.timestamp.toUint48()) revert DeadlinePassed();
        if (maxCost > t.maxCost) revert MaxCostExceeded();

        if (t.target != address(0) && !targetAllowed[t.target]) revert TargetNotAllowed();

        bytes32 structHash = keccak256(abi.encode(
            TICKET_HASH,
            t.sender,
            t.target,
            t.maxCost,
            t.deadline,
            t.chainId
        ));
        bytes32 digest = _hashTypedDataV4(structHash);

        bool ok = SignatureChecker.isValidSignatureNow(signer, digest, sig);
        if (!ok) revert InvalidSignature();

        context = abi.encode(userOp.sender);
        uint256 validUntil = uint256(t.deadline);
        uint256 validAfter = 0;
        uint256 aggregator = 0;
        validationData = (validUntil << 160) | (validAfter << 208) | aggregator;
    }

    function postOp(
        PostOpMode, // mode,
        bytes calldata, //context,
        uint256 // gasCost
        ) external view override {
        if (msg.sender != address(entryPoint)) revert NotEntryPoint();
    }
}