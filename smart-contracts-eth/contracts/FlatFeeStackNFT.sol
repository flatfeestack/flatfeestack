// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

import "@openzeppelin/contracts/access/Ownable.sol";

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Enumerable.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Pausable.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Burnable.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Votes.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

contract FlatFeeStackNFT is ERC721, ERC721Enumerable, ERC721URIStorage, ERC721Pausable, ERC721Burnable, EIP712, ERC721Votes, Ownable {

    uint48 constant public MAX_UINT48 = 281474976710655;
    uint256 public membershipFee = 1 ether;
    uint48 public membershipPeriod = 1 * 365 * 24 * 60 * 60; // 1 year
    mapping(uint256 => uint48) public membershipPayed;
    uint256 public currentTokenId;
    uint256 public councilCount;

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
        require(!isCouncil(tokenId), "Is council");
        require(
            ownerOf(tokenId) == msg.sender ||
            membershipPayed[tokenId] < block.timestamp ||
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
        if(status) {
            councilCount++;
        } else {
            require(councilCount > 3, "Two councils req");
            councilCount--;
        }
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

interface IFlatFeeStackNFT {
    function membershipPayed(uint256 tokenId) external view returns (uint48);
    function balanceOf(address owner) external view returns (uint256);
    function tokenOfOwnerByIndex(address owner, uint256 index) external view returns (uint256);
}
