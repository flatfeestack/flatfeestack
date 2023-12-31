// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import "@openzeppelin/contracts/access/Ownable.sol";

import "@openzeppelin/contracts/governance/Governor.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorCountingSimple.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotes.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotesQuorumFraction.sol";

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Enumerable.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Pausable.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Burnable.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Votes.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

contract FlatFeeStackNFT is ERC721, ERC721Enumerable, ERC721URIStorage, ERC721Pausable, Ownable, ERC721Burnable, EIP712, ERC721Votes {

    uint48 constant public MAX_UINT48 = 281474976710655;
    uint256 public membershipFee = 1 ether;
    uint48 public membershipPeriod = 1 * 365 * 24 * 60 * 60; // 1 year
    mapping(uint256 => uint48) public membershipPayed;
    uint256 public currentTokenId;

    event FlatFeeStackNFTCreated(address indexed addr, address indexed creator);
    event MembershipPayed(address indexed addr, uint256 indexed tokenId, uint256 indexed val);
    event CouncilSet(uint256 indexed tokenId, bool indexed status);
    event MembershipSettingsSet(uint256 indexed membershipFee, uint48 indexed membershipPeriod);

    constructor(address initialOwner, address council1, address council2)
        ERC721("FlatFeeStackNFT", "FlatFeeStackNFT")
        Ownable(initialOwner)
        EIP712("FlatFeeStackNFT", "1") {

        setCouncil(1, true);
        _safeMint(council1, 1);
        _delegate(council1, council1);

        setCouncil(2, true);
        _safeMint(council2, 2);
        _delegate(council2, council2);

        currentTokenId = 2;
        emit FlatFeeStackNFTCreated(address(this), msg.sender);
    }

    function _baseURI() internal pure override returns (string memory) {
        return "https://flatfeestack.io/nft/";
    }

    function pause() public onlyOwner {
        _pause();
    }

    function unpause() public onlyOwner {
        _unpause();
    }

    function safeMint(address addr, uint256 index1, bytes calldata signature1, uint256 index2, bytes calldata signature2) 
        external payable returns (uint256 tokenId) {

        tokenId = ++currentTokenId;
        bytes32 payloadHash = keccak256(
            abi.encodePacked(address(this), "safeMint", addr, "#", tokenId));

        bytes32 messageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", payloadHash));

        address council1 = ECDSA.recover(messageHash, signature1);
        address council2 = ECDSA.recover(messageHash, signature2);
        //we don't need to worry about signature used twice, as a tokenId must be unique. Thus,
        //calling a second time won't creat an NFT

        require(
            isCouncilIndex(council1, index1) && 
            isCouncilIndex(council2, index2) && 
            (council1 != council2) || (index1 != index2),
            "Signature err");
        
        _safeMint(addr, tokenId);
        _delegate(addr, addr);
        payMembership(tokenId);
    }

    function payMembership(uint256 tokenId) public payable {
        require(msg.value == membershipFee, "fee mismatch");
        require(!isCouncil(tokenId), "is council");
        
        uint48 old = membershipPayed[tokenId];
        if(old < block.timestamp) {
            old = SafeCast.toUint48(block.timestamp);
        }
        membershipPayed[tokenId] = old + membershipPeriod;
        
        //send to DAO
        payable(owner()).transfer(msg.value);
        emit MembershipPayed(msg.sender, tokenId, msg.value);
    }

    function burn(uint256 tokenId) public virtual override {
        require(
            ownerOf(tokenId) == msg.sender ||
            (membershipPayed[tokenId] < block.timestamp && !isCouncil(tokenId)) ||
            msg.sender == owner(),
            "Not tokenowner, payed membership, not contactowner");
            
        _burn(tokenId);
    }

    // The following functions are overrides required by Solidity.
    function _update(address to, uint256 tokenId, address auth) internal
        override(ERC721, ERC721Enumerable, ERC721Pausable, ERC721Votes)
        returns (address) {
        return super._update(to, tokenId, auth);
    }

    function _increaseBalance(address account, uint128 value) internal
        override(ERC721, ERC721Enumerable, ERC721Votes) {
        super._increaseBalance(account, value);
    }

    function isCouncilIndex(address council, uint256 index) public view returns (bool) {
        if(balanceOf(council) <= index) {
            return false;
        }
        uint256 tokenId = tokenOfOwnerByIndex(council, index);
        return isCouncil(tokenId);
    }

    function isCouncil(uint256 tokenId) public view returns (bool) {
        return membershipPayed[tokenId] == MAX_UINT48;
    }

    function setCouncil(uint256 tokenId, bool status) public onlyOwner {
        membershipPayed[tokenId] = status ? MAX_UINT48 : SafeCast.toUint48(block.timestamp) + membershipPeriod;
        emit CouncilSet(tokenId, status);
    }

    function setMembershipSettings(uint256 _membershipFee, uint48 _membershipPeriod) external onlyOwner {
        if (_membershipFee > 0) {
            membershipFee = _membershipFee;
        }
        if (_membershipPeriod > 0) {
            membershipPeriod = _membershipPeriod;
        }
        emit MembershipSettingsSet(_membershipFee, _membershipPeriod);
    }

    function tokenURI(uint256 tokenId) public view
        override(ERC721, ERC721URIStorage)
        returns (string memory) {

        string memory tokenBase;
        if(isCouncil(tokenId)) {
            tokenBase = string.concat(_baseURI(), "c/");
        } else {
            tokenBase = string.concat(_baseURI(), "m/");
        }
        return string.concat(tokenBase, Strings.toString(tokenId));
    }

    function supportsInterface(bytes4 interfaceId) public view 
        override(ERC721, ERC721Enumerable, ERC721URIStorage)
        returns (bool) {
        return super.supportsInterface(interfaceId);
    }

    function clock() public view virtual override(Votes)
        returns (uint48) {
        return SafeCast.toUint48(block.timestamp);
    }

    /**
     * @dev Machine-readable description of the clock as specified in EIP-6372.
     */
    // solhint-disable-next-line func-name-mixedcase
    function CLOCK_MODE() public view virtual override(Votes)
        returns (string memory)
    {
        // Check that the clock was not modified
        // https://eips.ethereum.org/EIPS/eip-6372
        require(clock() == block.timestamp);
        return "mode=timestamp";
    }

    function emergencyEth() external onlyOwner {
        payable(owner()).transfer(address(this).balance);
    }

    function emergencyERC20(address token) external onlyOwner {
        SafeERC20.safeTransfer(IERC20(token), owner(), IERC20(token).balanceOf(address(this)));
    }
}

contract FlatFeeStackDAO is Governor, GovernorSettings, GovernorCountingSimple, GovernorVotes, GovernorVotesQuorumFraction {

    uint256 public bylawsHash;
    mapping(uint256 => bool) public councilExecution;

    event BylawsChanged(uint256 indexed oldHash, uint256 indexed newHash);

    constructor(address council1, address council2)
        Governor("FlatFeeStackDAO")
        GovernorSettings(7 days, 1 days, 1)
        GovernorVotes(IVotes(new FlatFeeStackNFT(address(this), address(council1), address(council2))))
        GovernorVotesQuorumFraction(20) /* 20% */ {}

    function votingDelay() public view
        override(Governor, GovernorSettings) returns (uint256) {
        /* 
        The width of a slot is 7 days, so if a proposer proposes a vote in the middle of slot 1, 
        the delay will be set that this vote starts at end of slot 2 and beginning of slot 3. This
        gives a buffer of min 7 days, max ~14 days.

        | Slot 1 | Slot 2 | Slot 3 | Slot 4|

        Example : 1697068799 (Wed Oct 11 2023 23:59:59 GMT+0000), so the slot is: 2805 (2805.9999)
        Round up: (1697068799 + ((7 * 24 * 60 * 60) -1)) / (7 * 24 * 60 * 60) = 2806
        Round up: (1697068800 + ((7 * 24 * 60 * 60) -1)) / (7 * 24 * 60 * 60) = 2806
        Round up: (1697068801 + ((7 * 24 * 60 * 60) -1)) / (7 * 24 * 60 * 60) = 2807
        Dealy until next slot: (2807 * (7 * 24 * 60 * 60)) - 1697068799 = 604801 (7d, 1s)
        Dealy until next slot: (2807 * (7 * 24 * 60 * 60)) - 1697068800 = 604800 (7d)
        Dealy until next slot: (2808 * (7 * 24 * 60 * 60)) - 1697068801 = 604801 (13d, 23h, 59m, 59s)
        */

        uint256 nextSlot = ((block.timestamp + super.votingDelay() -1) / super.votingDelay()) + 1;
        return (nextSlot * super.votingDelay()) - block.timestamp;
    }

    function votingPeriod() public view 
        override(Governor, GovernorSettings) returns (uint256) {
        return super.votingPeriod();
    }

    function quorum(uint256 blockNumber) public view
        override(Governor, GovernorVotesQuorumFraction) returns (uint256) {
        return super.quorum(blockNumber);
    }

    function proposalThreshold() public view 
        override(Governor, GovernorSettings) returns (uint256) {
        return super.proposalThreshold();
    }

    function _queueOperations(uint256, address[] memory, uint256[] memory, bytes[] memory, bytes32) 
        internal view override returns (uint48) {
        return SafeCast.toUint48(super.votingDelay());
    }

    function requireTwoCouncil(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata calldatas,
        bytes32 descriptionHash,
        uint256 index1,
        uint256 index2,
        bytes calldata signature2
    ) internal returns (uint256 proposalId) {
        bytes32 proposalHash = keccak256(
            abi.encode(targets, values, calldatas, descriptionHash));

        bytes32 messageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", proposalHash));

        proposalId = uint256(proposalHash);
        
        require(councilExecution[proposalId] == false, "Cannot execute twice");
        councilExecution[proposalId] = true;

        address council2 = ECDSA.recover(messageHash, signature2);
        require(msg.sender != council2);

        require(
            FlatFeeStackNFT(address(token())).isCouncilIndex(msg.sender, index1) &&
            FlatFeeStackNFT(address(token())).isCouncilIndex(council2, index2),
            "No council sigs");

        return proposalId;
    }

    function councilExecute(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata calldatas,
        bytes32 descriptionHash,
        uint256 index1,
        uint256 index2,
        bytes calldata signature2
    ) external returns (uint256 proposalId) {
        proposalId = requireTwoCouncil(targets, values, calldatas, descriptionHash, index1, index2, signature2);
        _executeOperations(proposalId, targets, values, calldatas, descriptionHash);
        emit ProposalExecuted(proposalId);
        return proposalId;
    }

    function councilCancel(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata calldatas,
        bytes32 descriptionHash,
        uint256 index1,
        uint256 index2,
        bytes calldata signature2
    ) external returns (uint256 proposalId) {
        requireTwoCouncil(targets, values, calldatas, descriptionHash, index1, index2, signature2);
        proposalId = _cancel(targets, values, calldatas, descriptionHash);
        emit ProposalCanceled(proposalId);
        return proposalId;
    }


    /**
     * Sets a new hash value (newHash) of bylaws and emits an event indicating 
     * the change in bylaws hash from the old to the new value.
     */
    function setNewBylawsHash(uint256 newHash) external onlyGovernance {
        uint256 oldHash = bylawsHash;
        bylawsHash = newHash;
        emit BylawsChanged(oldHash, bylawsHash);
    }

    function clock() public view virtual override(Governor, GovernorVotes)
        returns (uint48) {
        return SafeCast.toUint48(block.timestamp);
    }

    /**
     * @dev Machine-readable description of the clock as specified in EIP-6372.
     */
    // solhint-disable-next-line func-name-mixedcase
    function CLOCK_MODE() public view virtual override(Governor, GovernorVotes)
        returns (string memory) {
        // Check that the clock was not modified
        // https://eips.ethereum.org/EIPS/eip-6372
        require(clock() == block.timestamp);
        return "mode=timestamp";
    }
}
