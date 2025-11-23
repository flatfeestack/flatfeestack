// SPDX-License-Identifier: MIT
pragma solidity ^0.8.25;

import {IEntryPoint} from "@account-abstraction/contracts/interfaces/IEntryPoint.sol";
import {IPaymaster} from "@account-abstraction/contracts/interfaces/IPaymaster.sol";
import "@account-abstraction/contracts/interfaces/PackedUserOperation.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

interface IFlatFeeStackNFT {
    function balanceOf(address owner) external view returns (uint256);
    function tokenOfOwnerByIndex(address owner, uint256 index) external view returns (uint256);
    function membershipPayed(uint256 tokenId) external view returns (uint48);
    function isCouncil(uint256 tokenId) external view returns (bool);
}

contract FlatFeeStackDAOPaymaster is IPaymaster, Ownable {
    using MessageHashUtils for bytes32;
    IEntryPoint public immutable entryPoint;
    IFlatFeeStackNFT public immutable nft;

    event PaymasterWithdraw(address indexed to, uint256 amount);

    constructor(IEntryPoint _entryPoint, address _nft) Ownable(msg.sender) {
        entryPoint = _entryPoint;
        nft = IFlatFeeStackNFT(_nft);
    }

    modifier onlyEntryPoint() {
        require(msg.sender == address(entryPoint), "Paymaster: Sender not EntryPoint");
        _;
    }

    /**
     * Returns true if user owns:
     *  - A council token (always valid)
     *  - A member token with unexpired membershipPayed
     */
    function isAuthorizedMember(address user) public view returns (bool) {
        uint256 count = nft.balanceOf(user);
        if (count == 0) return false;

        for (uint256 i = 0; i < count; i++) {
            uint256 tokenId = nft.tokenOfOwnerByIndex(user, i);

            // Council tokens are always allowed
            if (nft.isCouncil(tokenId)) return true;

            // Normal membership
            if (nft.membershipPayed(tokenId) >= block.timestamp)
                return true;
        }

        return false;
    }

    function isAuthorizedUserOp(PackedUserOperation calldata userOp)
        public
        view
        returns (bool)
    {
        address smartAccount = userOp.sender;

        // Smart Account already deployed
        if (isAuthorizedMember(smartAccount)) {
            return true;
        }

        // else check for the EOA
        bytes32 userOpHash = entryPoint.getUserOpHash(userOp);
        bytes32 signedHash = userOpHash.toEthSignedMessageHash();

        address eoa = ECDSA.recover(
            signedHash,
            userOp.signature
        );

        return isAuthorizedMember(eoa);
    }

    /**
     * Accept UserOperations for members
     */
    function validatePaymasterUserOp(
        PackedUserOperation calldata userOp,
        bytes32,
        uint256 maxCost
    ) external view onlyEntryPoint returns (bytes memory context, uint256 validationData) {
        require(
            entryPoint.balanceOf(address(this)) >= maxCost,
            "Paymaster: insufficient deposit"
        );

        require(
            isAuthorizedUserOp(userOp),
            "Paymaster: not an active member or council"
        );

        return ("", 0);
    }

    /**
     * postOp runs after execution and does not track anything
     */
    function postOp(
        PostOpMode mode,
        bytes calldata context,
        uint256 actualGasCost,
        uint256 actualUserOpFeePerGas
    ) external onlyEntryPoint {}

    /**
     * Fund the Paymaster on the Entrypoint
     */
    function deposit() external payable {
        entryPoint.depositTo{value: msg.value}(address(this));
    }

    /**
     * Withdraw Entrypoint funds to the owner
     */
    function withdraw(uint256 amount) external onlyOwner {
        entryPoint.withdrawTo(payable(owner()), amount);
        emit PaymasterWithdraw(owner(), amount);
    }
}