// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

abstract contract Base is Ownable {
    /**
     * @dev Maps each userId to its current already payed out amount. The userId never changes. 
     */
    mapping(uint256 => uint256) public payedOut;
    string public symbol;

    constructor(string memory _symbol) Ownable(msg.sender) {
        symbol = _symbol;
    }

    /**
     * @dev Gets the tea for the provided address.
     */
    function getPayedAmount(uint256 userId) external view returns (uint256 amount) {
        return payedOut[userId];
    }

    /**
     * @dev Gets the tea for the provided address.
     */
    function getClaimableAmount(uint256 userId, uint256 totalPayOut) external view returns (uint256 amount) {
        return totalPayOut - payedOut[userId];
    }

    /**
     * @dev Prepares everything to withdraw the earned amount. The signature has to be created by the contract owner and the signed message
     * is the hash of the concatenation of the account and tea.
     *
     * @param userId The user id that never changes
     * @param totalPayOut The total amount that the user earned.
     * @param signature The signature of the server.
     */
    function calculateWithdraw(uint256 userId, uint256 totalPayOut, bytes calldata signature) internal returns (uint256 amount) {
        require(totalPayOut > payedOut[userId], "Nothing to withdraw");

        bytes32 payloadHash = keccak256(
            abi.encodePacked(address(this), "calculateWithdraw", userId, "#", totalPayOut));

        bytes32 messageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", payloadHash));

        require(ECDSA.recover(messageHash, signature) == owner(), "Signature mismatch");

        uint256 old = payedOut[userId];
        payedOut[userId] = totalPayOut;

        return totalPayOut - old;
    }

    function sendRecover(address receiver, uint256 amount) external virtual;

    function getBalance(address dev) external view virtual returns (uint256);

    function withdraw(address dev, uint256 userId, uint256 totalPayOut, bytes calldata signature) external virtual;
}

contract PayoutEth is Base {

    constructor() Base("ETH") {}

    receive() external payable {}

    function getBalance(address addr) external view override returns (uint256 amount) {
        return addr.balance;
    }

    /**
     * @dev Send back from contract in case something is wrong. This should rarely happen
     */
    function sendRecover(address receiver, uint256 amount) external override onlyOwner {
        payable(receiver).transfer(amount);
    }

    /**
     * @dev Withdraw the earned amount. The signature has to be created by the contract owner and the signed message
     * is the hash of the concatenation of the account and tea.
     *
     * @param dev The address to withdraw to.
     * @param userId The user id that never changes
     * @param totalPayOut The total amount that the user earned.
     * @param signature The signature of the server.
     */
    function withdraw(address dev, uint256 userId, uint256 totalPayOut, bytes calldata signature
    ) external override {
        uint256 toBePaid = calculateWithdraw(userId, totalPayOut, signature);
        payable(dev).transfer(toBePaid);
    }
}

contract PayoutERC20 is Base {
    ERC20 public token;

    constructor(address _token) 
        Base(ERC20(_token).symbol()) {
        token = ERC20(_token);
    }

    function getBalance(address addr) external view override returns (uint256 amount) {
        return token.balanceOf(addr);
    }

    /**
     * @dev Send back from contract in case something is wrong. This should rarely happen
     */
    function sendRecover(address receiver, uint256 amount) external override onlyOwner {
        SafeERC20.safeTransfer(token, receiver, amount);
    }

    /**
     * @dev Withdraw the earned amount. The signature has to be created by the contract owner and the signed message
     * is the hash of the concatenation of the account and tea.
     *
     * @param dev The address to withdraw to.
     * @param userId The user id that never changes
     * @param totalPayOut The total amount that the user earned.
     * @param signature The signature of the server.
     */
    function withdraw(address dev, uint256 userId, uint256 totalPayOut, bytes calldata signature) external override {
        uint256 toBePaid = calculateWithdraw(userId, totalPayOut, signature);
        SafeERC20.safeTransfer(token, dev, toBePaid);
    }
}


