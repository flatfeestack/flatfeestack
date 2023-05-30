// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/token/ERC721/ERC721Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC721/extensions/ERC721EnumerableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/security/PausableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/cryptography/draft-EIP712Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC721/extensions/draft-ERC721VotesUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/utils/CountersUpgradeable.sol";


contract FlatFeeStackDAOSBT is Initializable, ERC721Upgradeable, ERC721EnumerableUpgradeable, PausableUpgradeable, AccessControlUpgradeable, EIP712Upgradeable, ERC721VotesUpgradeable {
    using CountersUpgradeable for CountersUpgradeable.Counter;

    CountersUpgradeable.Counter private _tokenIdCounter;

    uint256 public membershipFee = 1 ether;
    uint48 public membershipPeriod = 10 * 365 * 24 * 60 * 60; // 10 year
    mapping(address => uint48) public membershipPayed;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize() initializer public {
        __ERC721_init("FlatFeeStack DAO SBT", "FFSDS");
        __ERC721Enumerable_init();
        __Pausable_init();
        __AccessControl_init();
        __EIP712_init("FlatFeeStack DAO SBT", "1");
        __ERC721Votes_init();
        //the DAO contract need to become the default admin, for start its the contract creator
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
    }

    function _baseURI() internal pure override returns (string memory) {
        return "https://flatfeestack.io/sbt/";
    }

    function safeMint(
        address to,
        uint256 tokenId,
        uint8 v1,
        bytes32 r1,
        bytes32 s1,
        uint8 v2,
        bytes32 r2,
        bytes32 s2
    ) public payable {
        require(balanceOf(to) == 0, "1 address cannot have 2 NFTs");
        address council1 = ecrecover(
            keccak256(abi.encodePacked("safeMint", to, "#", tokenId)),
            v1,
            r1,
            s1
        );
        address council2 = ecrecover(
            keccak256(abi.encodePacked("safeMint", to, "#", tokenId)),
            v2,
            r2,
            s2
        );
        require(
            isCouncil(council1) && isCouncil(council2) && council1 != council2,
            "Signature not from council member"
        );

        require(msg.value >= membershipFee);
        membershipPayed[msg.sender] = SafeCastUpgradeable.toUint48(block.timestamp) + membershipPeriod;

        //member will have an id of 100 and more, the council will have id 1-99
        uint256 nextTokenId = _tokenIdCounter.current() + 100;
        require(tokenId == nextTokenId, "wrong tokenId");
        _tokenIdCounter.increment();
        _safeMint(to, tokenId);
    }

    function safeMintCouncil(
        address to,
        uint256 tokenId
    ) public onlyRole(DEFAULT_ADMIN_ROLE) {
        require(
            tokenId > 0 && tokenId < 100,
            "Cannot have more than 99 council members"
        );
        require(balanceOf(to) == 0, "1 address cannot have 2 NFTs");
        address owner = _ownerOf(tokenId);
        if (owner == address(0)) {
            _safeMint(to, tokenId); //we create a new NFT
        } else {
            _burn(tokenId);
            _safeMint(to, tokenId); //we create a new NFT
        }
    }

    function burn(uint256 tokenId) external {
        require(
            ownerOf(tokenId) == msg.sender ||
            (isMember(ownerOf(tokenId)) &&
                membershipPayed[ownerOf(tokenId)] < block.timestamp) ||
            hasRole(DEFAULT_ADMIN_ROLE, msg.sender),
            "Only tokenowner or unpayed membership"
        );
        _burn(tokenId);
    }

    function setMembershipSettings(
        uint256 _membershipFee,
        uint48 _membershipPeriod
    ) public onlyRole(DEFAULT_ADMIN_ROLE) {
        if (_membershipFee > 0) {
            membershipFee = _membershipFee;
        }
        if (_membershipPeriod > 0) {
            membershipPeriod = _membershipPeriod;
        }
    }

    function isCouncil(address owner) public view returns (bool) {
        return tokenOfOwnerByIndex(owner, 0) > 0 && tokenOfOwnerByIndex(owner, 0) < 100;
    }

    function isMember(address owner) public view returns (bool) {
        return tokenOfOwnerByIndex(owner, 0) >= 100;
    }


    function _beforeTokenTransfer(address from, address to, uint256 tokenId, uint256 batchSize)
    internal
    whenNotPaused
    override(ERC721Upgradeable, ERC721EnumerableUpgradeable)
    {
        require(
            from == address(0) || to == address(0),
            "This a Soulbound token."
        );
        super._beforeTokenTransfer(from, to, tokenId, batchSize);
    }

    // The following functions are overrides required by Solidity.

    function _afterTokenTransfer(address from, address to, uint256 tokenId, uint256 batchSize) internal
    override(ERC721Upgradeable, ERC721VotesUpgradeable)
    {
        super._afterTokenTransfer(from, to, tokenId, batchSize);
    }

    function supportsInterface(bytes4 interfaceId) public view
    override(ERC721Upgradeable, ERC721EnumerableUpgradeable, AccessControlUpgradeable) returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }

    function clock() public view virtual override returns (uint48)  {
        return SafeCastUpgradeable.toUint48(block.timestamp);
    }

    /**
     * @dev Machine-readable description of the clock as specified in EIP-6372.
     */
    // solhint-disable-next-line func-name-mixedcase
    function CLOCK_MODE() public view virtual override returns (string memory) {
        // Check that the clock was not modified
        // https://eips.ethereum.org/EIPS/eip-6372
        require(clock() == block.timestamp);
        return "mode=timestamp";
    }

    function pause() public onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }

    function unpause() public onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }
}
