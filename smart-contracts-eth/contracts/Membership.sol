// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "./Wallet.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract Membership is Initializable {
    enum MembershipStatus {
        nonMember,
        requesting,
        whitelistedByOne,
        isMember
    }

    address public delegate;
    uint256 public minimumWhitelister;
    uint256 public whitelisterListLength;
    uint256 public membershipFee;

    Wallet private _wallet;

    mapping(uint256 => address) public whitelisterList;
    mapping(address => MembershipStatus) internal membershipList;
    mapping(address => address) internal firstWhiteLister;

    mapping(address => uint256) public nextMembershipFeePayment;

    event ChangeInMembershipStatus(
        address indexed accountAddress,
        uint256 currentStatus
    );

    event ChangeInWhiteLister(
        address indexed concernedWhitelister,
        bool removedOrAdded
    );

    event ChangeInDelegate(
        address indexed concernedDelegate,
        bool removedOrAdded
    );

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

    modifier delegateOnly() {
        require(msg.sender == delegate, "only delegate");
        _;
    }

    modifier whitelisterOnly() {
        require(isWhitelister(msg.sender) == true, "only whitelisters");
        _;
    }

    function initialize(
        address _delegate,
        address _whitelisterOne,
        address _whitelisterTwo,
        Wallet _walletContract
    ) public initializer {
        minimumWhitelister = 2;
        whitelisterListLength = 2;
        delegate = _delegate;
        membershipFee = 30000 wei;
        _wallet = _walletContract;

        whitelisterList[0] = _whitelisterOne;
        whitelisterList[1] = _whitelisterTwo;

        membershipList[_delegate] = MembershipStatus.isMember;
        membershipList[_whitelisterOne] = MembershipStatus.isMember;
        membershipList[_whitelisterTwo] = MembershipStatus.isMember;

        nextMembershipFeePayment[_delegate] = block.timestamp;
        nextMembershipFeePayment[_whitelisterOne] = block.timestamp;
        nextMembershipFeePayment[_whitelisterTwo] = block.timestamp;

        emit ChangeInMembershipStatus(
            delegate,
            uint256(MembershipStatus.isMember)
        );

        emit ChangeInMembershipStatus(
            _whitelisterOne,
            uint256(MembershipStatus.isMember)
        );

        emit ChangeInMembershipStatus(
            _whitelisterTwo,
            uint256(MembershipStatus.isMember)
        );
    }

    function requestMembership() public nonMemberOnly returns (bool) {
        membershipList[msg.sender] = MembershipStatus.requesting;
        emit ChangeInMembershipStatus(
            msg.sender,
            uint256(MembershipStatus.requesting)
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
            membershipList[_adr] == MembershipStatus.isMember,
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
            whitelisterListLength > minimumWhitelister,
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
            membershipList[_adr] == MembershipStatus.requesting ||
                (membershipList[_adr] == MembershipStatus.whitelistedByOne &&
                    firstWhiteLister[_adr] != msg.sender)
        );
        if (membershipList[_adr] == MembershipStatus.requesting) {
            membershipList[_adr] = MembershipStatus.whitelistedByOne;
            firstWhiteLister[_adr] = msg.sender;
            emit ChangeInMembershipStatus(
                _adr,
                uint256(MembershipStatus.whitelistedByOne)
            );
        } else {
            membershipList[_adr] = MembershipStatus.isMember;
            nextMembershipFeePayment[_adr] = block.timestamp;
            emit ChangeInMembershipStatus(
                _adr,
                uint256(MembershipStatus.isMember)
            );
        }
        return true;
    }

    function payMembershipFee() public payable memberOnly {
        uint256 nextDueDate = nextMembershipFeePayment[msg.sender];
        require(nextDueDate <= block.timestamp, "Membership fee not due yet.");
        // we don't say "no" if somebody pays more than they should :)
        require(
            msg.value >= membershipFee,
            "Membership fee not fully covered."
        );

        nextMembershipFeePayment[msg.sender] = nextDueDate + 365 days;
        _wallet.payContribution{value: msg.value}(msg.sender);
    }

    function setMembershipFee(uint256 newMembershipFee) public delegateOnly {
        membershipFee = newMembershipFee;
    }

    function setDelegate(address _adr) public returns (bool) {
        // TODO: require oder modifier einbauen, dass der sender vom verwalter der proposals kommt
        require(
            membershipList[_adr] == MembershipStatus.isMember,
            "Has to be member to become delegate"
        );
        require(
            delegate != _adr,
            "Can't set the delegate to the same delegate again"
        );
        address oldDelegate = delegate;
        delegate = _adr;
        emit ChangeInDelegate(oldDelegate, false);
        emit ChangeInDelegate(delegate, true);
        return true;
    }
}
