// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract Membership is Initializable {
    enum membershipStatus {
        nonMember,
        requesting,
        whitelistedByOne,
        isMember
    }

    address public delegate;
    uint256 MINIMUM_WHITELISTER;
    uint256 public whitelisterListLength;

    mapping(uint256 => address) public whitelisterList;
    mapping(address => membershipStatus) internal membershipList;
    mapping(address => address) internal firstWhiteLister;

    event ChangeInMembershipStatus(
        address indexed accountAddress,
        uint256 currentStatus
    );

    event ChangeInWhiteLister(
        address indexed concernedWhitelister,
        bool removedOrAdded
    );

    modifier nonMemberOnly() {
        require(
            membershipList[msg.sender] == membershipStatus.nonMember,
            "This function can only be called by non-members"
        );
        _;
    }

    modifier delegateOnly() {
        require(
            msg.sender == delegate,
            "This function can only be called by a delegate"
        );
        _;
    }

    modifier whitelisterOnly() {
        require(isWhitelister(msg.sender) == true);
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

    function addWhitelister(address _adr) public delegateOnly returns (bool) {
        require(delegate != _adr, "The delegate can't become a whitelister");
        require(
            membershipList[_adr] == membershipStatus.isMember,
            "A whitelister must be a member"
        );
        require(
            isWhitelister(_adr) == false,
            "This address is already a whitelister"
        );
        whitelisterList[whitelisterListLength] = _adr;
        whitelisterListLength++;
        emit ChangeInWhiteLister(_adr, true);
        return true;
    }

    function removeWhitelister(address _adr)
        public
        delegateOnly
        returns (bool)
    {
        require(
            isWhitelister(_adr) == true,
            "This address is not a whitelister"
        );
        require(
            whitelisterListLength > MINIMUM_WHITELISTER,
            "Can't remove because there is a minimum of 2 whitelisters"
        );
        uint256 i;
        for (i = 0; i < whitelisterListLength - 1; i++) {
            if (whitelisterList[i] == _adr) {
                break;
            }
        }
        if (i != whitelisterListLength - 1) {
            whitelisterList[i] = whitelisterList[whitelisterListLength - 1];
        }
        whitelisterListLength--;
        emit ChangeInWhiteLister(_adr, false);
        return true;
    }

    function whitelistMember(address _adr)
        public
        whitelisterOnly
        returns (bool)
    {
        require(
            membershipList[_adr] == membershipStatus.requesting ||
                (membershipList[_adr] == membershipStatus.whitelistedByOne &&
                    firstWhiteLister[_adr] != msg.sender)
        );
        if (membershipList[_adr] == membershipStatus.requesting) {
            membershipList[_adr] = membershipStatus.whitelistedByOne;
            firstWhiteLister[_adr] = msg.sender;
            emit ChangeInMembershipStatus(
                _adr,
                uint256(membershipStatus.whitelistedByOne)
            );
        } else {
            membershipList[_adr] = membershipStatus.isMember;
            emit ChangeInMembershipStatus(
                _adr,
                uint256(membershipStatus.isMember)
            );
        }
        return true;
    }
}
