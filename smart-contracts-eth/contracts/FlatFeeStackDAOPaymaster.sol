// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

//TODO clean

interface IEntryPoint {
    function depositTo(address) external payable;
    function balanceOf(address) external view returns (uint256);
}

interface IPaymaster {
    enum PostOpMode { opSucceeded, opReverted }
}

interface IFlatFeeStackNFT {
    function membershipPayed(uint256 tokenId) external view returns (uint48);
    function balanceOf(address owner) external view returns (uint256);
    function tokenOfOwnerByIndex(address owner, uint256 index) external view returns (uint256);
}

interface IFlatFeeStackNFT_Extended is IFlatFeeStackNFT {
    function isCouncil(uint256 tokenId) external view returns (bool);
}

contract FlatFeeStackDAOPaymaster {

    // -----------------------------------------------------------
    // STORAGE
    // -----------------------------------------------------------
    IEntryPoint public immutable entryPoint;
    IFlatFeeStackNFT_Extended public immutable nft;
    address public immutable dao;

    // -----------------------------------------------------------
    // EVENTS
    // -----------------------------------------------------------
    event UserOpSponsored(address indexed user, uint256 maxCost);
    event Deposit(address indexed from, uint256 amount);
    event EmergencyWithdraw(address indexed to, uint256 amount);

    // -----------------------------------------------------------
    // CONSTRUCTOR
    // -----------------------------------------------------------
    constructor(address entryPoint_, address nftAddress_, address daoAddress_) {
        require(entryPoint_ != address(0), "entryPoint zero");
        require(nftAddress_ != address(0), "NFT zero");
        require(daoAddress_ != address(0), "DAO zero");

        entryPoint = IEntryPoint(entryPoint_);
        nft = IFlatFeeStackNFT_Extended(nftAddress_);
        dao = daoAddress_;
    }

    modifier onlyEntryPoint() {
        require(msg.sender == address(entryPoint), "Not EntryPoint");
        _;
    }

    modifier onlyDAO() {
        require(msg.sender == dao, "Not DAO");
        _;
    }

    // -----------------------------------------------------------
    // MEMBERSHIP CHECK LOGIC
    // -----------------------------------------------------------

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

    // -----------------------------------------------------------
    // PAYMASTER LOGIC
    // -----------------------------------------------------------

    /**
     * validatePaymasterUserOp (before execution)
     * - Verifies deposit
     * - Verifies user membership
     * - Returns context for postOp
     */
    function validatePaymasterUserOp(
        bytes calldata userOp,
        bytes32,      // userOpHash
        uint256 maxCost
    )
        external
        view
        onlyEntryPoint
        returns (bytes memory context, uint256 validationData)
    {
        // Extract sender from PackedUserOperation
        address sender;
        assembly {
            sender := calldataload(userOp.offset)
        }

        require(isAuthorizedMember(sender), "Not a valid member");

        require(
            entryPoint.balanceOf(address(this)) >= maxCost,
            "Insufficient PM deposit"
        );

        context = abi.encode(sender, maxCost);
        validationData = 0; // valid
    }

    /**
     * postOp: runs after UserOperation finishes.
     * - You currently use flat-fee model, cost is deducted from deposit.
     */
    function postOp(
        IPaymaster.PostOpMode,
        bytes calldata context,
        uint256 actualGasCost
    )
        external
        onlyEntryPoint
    {
        (address sender, uint256 maxCost) = abi.decode(context, (address, uint256));

        // For now: no extra charging. Cost already deducted from deposit.
        emit UserOpSponsored(sender, maxCost);
    }

    // -----------------------------------------------------------
    // PAYMASTER DEPOSIT CONTROLS
    // -----------------------------------------------------------

    /// Adds ETH deposit to EntryPoint for this Paymaster
    function deposit() external payable {
        entryPoint.depositTo{value: msg.value}(address(this));
        emit Deposit(msg.sender, msg.value);
    }

    /// DAO can rescue ETH sent directly to contract (not deposit)
    function emergencyWithdraw(address to, uint256 amount) external onlyDAO {
        payable(to).transfer(amount);
        emit EmergencyWithdraw(to, amount);
    }

    // Debug helper
    function ping() external pure returns (string memory) {
        return "Full Functional Paymaster OK";
    }
}
