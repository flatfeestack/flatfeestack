// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

contract Accessible {
    enum MembershipStatus {
        nonMember,
        requesting,
        whitelistedByOne,
        isMember
    }

    address public representative;
    uint256 public minimumWhitelister;
    uint256 public whitelisterListLength;
    uint256 public membershipFee;

    mapping(uint256 => address) public whitelisterList;
    address[] public members;
    mapping(address => MembershipStatus) internal membershipList;
    mapping(address => address) internal firstWhiteLister;

    modifier nonMemberOnly() {
        require(
            membershipList[msg.sender] == MembershipStatus.nonMember,
            "only non-members"
        );
        _;
    }

    modifier memberOnly() {
        require(
            membershipList[msg.sender] == MembershipStatus.isMember,
            "only members"
        );
        _;
    }

    modifier representativeOnly() {
        require(msg.sender == representative, "only representative");
        _;
    }

    modifier whitelisterOnly() {
        require(isWhitelister(msg.sender) == true, "only whitelisters");
        _;
    }

    modifier representativeOrWhitelisterOnly() {
        require(
            msg.sender == representative || isWhitelister(msg.sender),
            "whitelister / represetative only"
        );
        _;
    }

    function isWhitelister(address _adr) public view returns (bool) {
        bool check = false;
        for (uint256 i = 0; i < whitelisterListLength; i++) {
            if (whitelisterList[i] == _adr) {
                check = true;
                break;
            }
        }
        return check;
    }

    function getMembersLength() external view returns (uint256) {
        return members.length;
    }
}
