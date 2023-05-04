// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Enumerable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/cryptography/draft-EIP712.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/draft-ERC721Votes.sol";
import "@openzeppelin/contracts/utils/Counters.sol";
import "@openzeppelin/contracts/utils/Address.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

contract FlatFeeStackDAOVote is ERC721, Ownable, EIP712, ERC721Votes {
    using Counters for Counters.Counter;

    Counters.Counter private _tokenIdCounter;

    uint256 public membershipFee = 1 ether;
    uint256 public membershipPeriod = 365 * 24 * 60 * 60; // 1 year
    uint256 public membershipPayment;

    constructor()
        ERC721("FlatFeeStack DAO NFT", "FFSDAO NFT")
        EIP712("FlatFeeStack DAO NFT", "1")
    {
        //this needs to change to the address of the DAO contract
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _tokenIdCounter.current(100);
    }

    function safeMint(
        address to,
        uint256 timestamp,
        uint8 v1,
        bytes32 r1,
        bytes32 s1,
        uint8 v2,
        bytes32 r2,
        bytes32 s2
    ) public {
        require(_ownedTokens[to][0] == 0, "1 address cannot have 2 NFTs");
        address council1 = ecrecover(
            keccak256(abi.encodePacked(to, "#", timestamp)),
            v1,
            r1,
            s1
        );
        address council2 = ecrecover(
            keccak256(abi.encodePacked(to, "#", timestamp)),
            v2,
            r2,
            s2
        );
        require(
            isCouncil(council1) && isCouncil(council2) && council1 != council2,
            "Signature not from council member"
        );

        require(msg.value >= membershipFee);

        uint256 tokenId = _tokenIdCounter.current();
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
        require(_ownedTokens[to][0] == 0, "1 address cannot have 2 NFTs");
        address owner = _ownerOf(tokenId);
        if (owner == address(0)) {
            _tokenIdCounter.increment();
            _safeMint(to, tokenId); //we create a new NFT
        } else {
            _transfer(owner, to, tokenId); //we take it away, as this was the outcome of the vote
        }
    }

    function setMembershipSettings(
        uint256 _membershipFee,
        uint256 _membershipPeriod
    ) public onlyRole(DEFAULT_ADMIN_ROLE) {
        if (_membershipFee > 0) {
            membershipFee = _membershipFee;
        }
        if (_membershipPeriod > 0) {
            membershipPeriod = _membershipPeriod;
        }
    }

    function execute(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas
    ) public payable override onlyRole(DEFAULT_ADMIN_ROLE) {
        string memory errorMessage = "SBT: call reverted without message";
        for (uint256 i = 0; i < targets.length; ++i) {
            (bool success, bytes memory returndata) = targets[i].call{
                value: values[i]
            }(calldatas[i]);
            Address.verifyCallResult(success, returndata, errorMessage);
        }
    }

    function isCouncil(address owner) public view returns (boolean) {
        return _ownedTokens[owner][0] > 0 && _ownedTokens[owner][0] < 100;
    }

    function isMember(address owner) public view returns (boolean) {
        return _ownedTokens[owner][0] >= 100;
    }

    // The following functions are overrides required by Solidity.

    function _afterTokenTransfer(
        address from,
        address to,
        uint256 tokenId,
        uint256 batchSize
    ) internal override(ERC721, ERC721Votes) {
        super._afterTokenTransfer(from, to, tokenId, batchSize);
    }

    //https://docs.chainstack.com/tutorials/gnosis/simple-soulbound-token-with-remix-and-openzeppelin#interact-with-the-contract
    function _beforeTokenTransfer(
        address from,
        address to,
        uint256
    ) internal pure override {
        require(
            from == address(0) || to == address(0),
            "This a Soulbound token."
        );
    }

    function burn(uint256 tokenId) external {
        require(
            ownerOf(tokenId) == msg.sender ||
                (isMember(ownerOf(tokenId)) &&
                    membershipPayed[ownerOf(tokenId)] + membershipPeriod <
                    block.timestamp) ||
                hasRole(DAO_DECISION, msg.sender),
            "Only tokenowner or unpayed membership"
        );
        _burn(tokenId);
    }

    function clock() public view virtual returns (uint48) {
        return SafeCast.toUint48(block.timestamp);
    }

    /**
     * @dev Machine-readable description of the clock as specified in EIP-6372.
     */
    // solhint-disable-next-line func-name-mixedcase
    function CLOCK_MODE() public view virtual returns (string memory) {
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
