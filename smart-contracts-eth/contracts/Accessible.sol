// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

contract Accessible {
    enum MembershipStatus {
        nonMember,
        requesting,
        approvedByOne,
        isMember
    }

    uint256 public minimumChairmen;
    uint256 public membershipFee;

    address[] public chairmen;
    address[] public members;

    mapping(address => MembershipStatus) internal membershipList;
    mapping(address => address) internal firstApproval;

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

    modifier chairmenOnly() {
        require(isChairman(msg.sender) == true, "only chairmen");
        _;
    }

    function isChairman(address _adr) public view returns (bool) {
        bool check = false;

        for (uint256 i = 0; i < this.getChairmenLength(); i++) {
            if (chairmen[i] == _adr) {
                check = true;
                break;
            }
        }

        return check;
    }

    function getChairmenLength() external view returns (uint256) {
        return chairmen.length;
    }

    function getMembersLength() external view returns (uint256) {
        return members.length;
    }
}
