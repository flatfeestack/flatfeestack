// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

abstract contract PayoutBase is Initializable, OwnableUpgradeable {
    /**
     * @dev Maps each userId to its current already payed out amount. The userId never changes
     */
    mapping(string => uint256) public payedOut;
    string private currencyCode;

    function payoutInit(string memory _currencyCode) internal onlyInitializing {
        __Ownable_init();
        payoutInitUnchained(_currencyCode);
    }

    function payoutInitUnchained(
        string memory _currencyCode
    ) internal onlyInitializing {
        currencyCode = _currencyCode;
    }

    function sendRecover(
        address payable receiver,
        uint256 amount
    ) external virtual;

    /**
     * @dev Gets the tea for the provided address.
     */
    function getPayedOut(
        string calldata userId
    ) external view returns (uint256) {
        return payedOut[userId];
    }

    /**
     * @dev Gets the tea for the provided address.
     */
    function getClaimableAmount(
        string calldata userId,
        uint256 totalPayOut
    ) external view returns (uint256) {
        return totalPayOut - payedOut[userId];
    }

    /**
     * @dev Prepares everything to withdraw the earned amount. The signature has to be created by the contract owner and the signed message
     * is the hash of the concatenation of the account and tea.
     *
     * @param userId The user id that never changes
     * @param totalPayOut The total amount that the user earned.
     * @param v The recovery byte of the signature.
     * @param r The r value of the signature.
     * @param s The s value of the signature.
     */
    function calculateWithdraw(
        string calldata userId,
        uint256 totalPayOut,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) internal returns (uint256) {
        require(totalPayOut > payedOut[userId], "No new funds to be withdrawn");

        bytes32 payloadHash = keccak256(
            abi.encode(userId, "#", totalPayOut, currencyCode)
        );
        bytes32 messageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", payloadHash)
        );

        require(
            ecrecover(messageHash, v, r, s) == owner(),
            "Signature no match"
        );
        uint256 old = payedOut[userId];
        payedOut[userId] = totalPayOut;

        return totalPayOut - old;
    }

    function withdraw(
        address payable dev,
        string calldata userId,
        uint256 totalPayOut,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external virtual;
}
