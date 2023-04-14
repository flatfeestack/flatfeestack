// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC20/IERC20Upgradeable.sol";

import "hardhat/console.sol";

contract Payout is OwnableUpgradeable {
    /**
     * @dev Maps each userId to its current already payed out amount. The userId never changes
     */
    mapping(bytes32 => uint256) public payedOut;

    function initialize() public initializer {
        __Ownable_init();
    }

    receive() external payable {}

    /**
     * @dev Send back from contract in case something is wrong. This should rarely happen
     */
    function sndRecoverEth(
        address payable receiver,
        uint256 amount
    ) external onlyOwner {
        receiver.transfer(amount);
    }

    /**
     * @dev Send back from contract in case something is wrong. This should never happen
     */
    function sndRecoverToken(
        address receiver,
        address contractAddress,
        uint256 amount
    ) external onlyOwner {
        IERC20Upgradeable(contractAddress).transfer(receiver, amount);
    }

    /**
     * @dev Gets the tea for the provided address.
     */
    function getPayedOut(bytes32 userId) external view returns (uint256) {
        return payedOut[userId];
    }

    /**
     * @dev Gets the tea for the provided address.
     */
    function getClaimableAmount(
        bytes32 userId,
        uint256 totalPayOut
    ) external view returns (uint256) {
        return totalPayOut - payedOut[userId];
    }

    /**
     * @dev Withdraws the earned amount. The signature has to be created by the contract owner and the signed message
     * is the hash of the concatenation of the account and tea.
     *
     * @param dev The address to withdraw to.
     * @param userId The user id that never changes
     * @param totalPayOut The total amount that the user earned.
     * @param v The recovery byte of the signature.
     * @param r The r value of the signature.
     * @param s The s value of the signature.
     */
    function withdraw(
        address payable dev,
        bytes32 userId,
        uint256 totalPayOut,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external {
        bytes32 payloadHash = keccak256(abi.encode(userId, "#", totalPayOut));
        bytes32 messageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", payloadHash)
        );

        require(totalPayOut > payedOut[userId], "No new funds to be withdrawn");
        require(
            ecrecover(messageHash, v, r, s) == owner(),
            "Signature no match"
        );
        uint256 old = payedOut[userId];
        payedOut[userId] = totalPayOut;
        // transfer reverts transaction if not successful.
        dev.transfer(totalPayOut - old);
    }
}
