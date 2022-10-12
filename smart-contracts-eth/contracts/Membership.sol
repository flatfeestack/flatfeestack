// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract Membership is Initializable {
    enum membershipStatus {
        nonMember,
        requesting,
        whitelisted,
        isMember
    }

    address public delegate;
    uint256 MINIMUM_WHITELISTER;
    uint256 public whitelisterListLength;

    mapping(uint256 => address) public whitelisterList;
    mapping(address => membershipStatus) internal membershipList;

    event ChangeInMembershipStatus(
        address indexed accountAddress,
        uint256 currentStatus
    );

    modifier nonMemberOnly() {
        require(
            membershipList[msg.sender] == membershipStatus.nonMember,
            "This function can only be called by non-members"
        );
        _;
    }

    function initialize(
        address _delegate,
        address _whitelisterOne,
        address _whitelisterTwo
    ) public initializer {
        MINIMUM_WHITELISTER = 2;
        whitelisterListLength = 2;
        delegate = _delegate;
        whitelisterList[0] = _whitelisterOne;
        whitelisterList[1] = _whitelisterTwo;
        membershipList[_delegate] = membershipStatus.isMember;
        membershipList[_whitelisterOne] = membershipStatus.isMember;
        membershipList[_whitelisterTwo] = membershipStatus.isMember;
        emit ChangeInMembershipStatus(
            delegate,
            uint256(membershipStatus.isMember)
        );
        emit ChangeInMembershipStatus(
            _whitelisterOne,
            uint256(membershipStatus.isMember)
        );
        emit ChangeInMembershipStatus(
            _whitelisterTwo,
            uint256(membershipStatus.isMember)
        );
    }

    function requestMembership() public nonMemberOnly returns (bool) {
        membershipList[msg.sender] = membershipStatus.requesting;
        emit ChangeInMembershipStatus(
            msg.sender,
            uint256(membershipStatus.requesting)
        );
        return true;
    }

    function getMembershipStatus(address _adr) public view returns (uint256) {
        return uint256(membershipList[_adr]);
    }
}
