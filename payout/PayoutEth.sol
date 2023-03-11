// SPDX-License-Identifier: MIT
pragma solidity ^0.8.7;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract PayoutEth {

    /**
    * @dev Maps each userId to its current already payed out amount. The userId never changes
    */
    mapping(bytes32 => uint256) public payedOut;

    /**
    * @dev The contract owner
    */
    address public owner;

    modifier onlyOwner() {
        require(msg.sender == owner, "No authorization");
        _;
    }

    constructor () {
        owner = payable(msg.sender);
    }

    receive() external payable {
    }

    /**
    * @dev Send back from contract in case something is wrong. This should rarely happen
    */
    function sndRecoverEth(address payable receiver, uint256 amount) external onlyOwner() {
        receiver.transfer(amount);
    }

    /**
    * @dev Send back from contract in case something is wrong. This should never happen
    */
    function sndRecoverToken(address receiver, address contractAddress, uint256 amount) external onlyOwner() {
        IERC20(contractAddress).transfer(receiver, amount);
    }

    /**
    * @dev Changes the owner of this contract.
    */
    function changeOwner(address payable newOwner) external onlyOwner() {
        owner = newOwner;
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
    function getClaimableAmount(bytes32 userId, uint256 totalPayOut) external view returns (uint256) {
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
    function withdraw(address payable dev, bytes32 userId, uint256 totalPayOut, uint8 v, bytes32 r, bytes32 s) external {
        require(totalPayOut > payedOut[userId], "No new funds to be withdrawn");
        require(ecrecover(keccak256(abi.encodePacked(userId, "#", totalPayOut)), v, r, s) == owner, "Signature no match");
        uint256 old = payedOut[userId];
        payedOut[userId] = totalPayOut;
        // transfer reverts transaction if not successful.
        dev.transfer(totalPayOut - old);
    }

}
