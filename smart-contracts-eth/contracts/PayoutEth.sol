// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import {PayoutBase} from "./PayoutBase.sol";

contract PayoutEth is PayoutBase {
    function initialize() public initializer {
        payoutInit("ETH");
    }

    receive() external payable {}

    /**
     * @dev Send back from contract in case something is wrong. This should rarely happen
     */
    function sendRecover(
        address payable receiver,
        uint256 amount
    ) external override onlyOwner {
        receiver.transfer(amount);
    }

    /**
     * @dev Withdraw the earned amount. The signature has to be created by the contract owner and the signed message
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
    ) external override {
        uint256 toBePaid = calculateWithdraw(userId, totalPayOut, v, r, s);

        // transfer reverts transaction if not successful.
        dev.transfer(toBePaid);
    }
}
