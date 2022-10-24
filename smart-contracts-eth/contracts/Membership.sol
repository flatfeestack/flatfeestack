// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "./Wallet.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/utils/CheckpointsUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/utils/IVotesUpgradeable.sol";

// we rely on time to track membership payments
// however, we don't care about second-level precision, as we deal with a much longer time period
// there is a good exaplanation about this on StackExchange https://ethereum.stackexchange.com/a/117874
/* solhint-disable not-rely-on-time */
contract Membership is Initializable, IVotesUpgradeable {
    using CheckpointsUpgradeable for CheckpointsUpgradeable.History;

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

    Wallet private _wallet;

    mapping(uint256 => address) public whitelisterList;
    mapping(address => MembershipStatus) internal membershipList;
    mapping(address => address) internal firstWhiteLister;

    mapping(address => uint256) public nextMembershipFeePayment;

    // used for IVotes
    mapping(address => CheckpointsUpgradeable.History) private _voteCheckpoints;
    CheckpointsUpgradeable.History private _totalCheckpoints;

    event ChangeInMembershipStatus(
        address indexed accountAddress,
        uint256 currentStatus
    );

    event ChangeInWhiteLister(
        address indexed concernedWhitelister,
        bool removedOrAdded
    );

    event ChangeInRepresentative(
        address indexed concernedRepresentative,
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

    function initialize(
        address _representative,
        address _whitelisterOne,
        address _whitelisterTwo,
        Wallet _walletContract
    ) public initializer {
        minimumWhitelister = 2;
        whitelisterListLength = 2;
        representative = _representative;
        membershipFee = 30000 wei;
        _wallet = _walletContract;

        whitelisterList[0] = _whitelisterOne;
        whitelisterList[1] = _whitelisterTwo;

        membershipList[_representative] = MembershipStatus.isMember;
        membershipList[_whitelisterOne] = MembershipStatus.isMember;
        membershipList[_whitelisterTwo] = MembershipStatus.isMember;

        nextMembershipFeePayment[_representative] = block.timestamp;
        nextMembershipFeePayment[_whitelisterOne] = block.timestamp;
        nextMembershipFeePayment[_whitelisterTwo] = block.timestamp;

        _totalCheckpoints.push(_add, 3);
        _voteCheckpoints[_representative].push(_add, 1);
        _voteCheckpoints[_whitelisterOne].push(_add, 1);
        _voteCheckpoints[_whitelisterTwo].push(_add, 1);

        emit ChangeInMembershipStatus(
            representative,
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

    function addWhitelister(address _adr)
        public
        representativeOnly
        returns (bool)
    {
        require(representative != _adr, "Can't become whitelister!");
        require(
            membershipList[_adr] == MembershipStatus.isMember,
            "A whitelister must be a member"
        );
        require(isWhitelister(_adr) == false, "Is already whitelister!");
        whitelisterList[whitelisterListLength] = _adr;
        whitelisterListLength++;
        emit ChangeInWhiteLister(_adr, true);
        return true;
    }

    function _removeWhitelister(address _adr)
        internal
        representativeOrWhitelisterOnly
        returns (bool)
    {
        require(isWhitelister(_adr) == true, "Is no whitelister!");
        require(
            whitelisterListLength > minimumWhitelister,
            "Minimum whitelister not met!"
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

    function removeWhitelister(address _adr)
        public
        representativeOnly
        returns (bool)
    {
        require(isWhitelister(_adr) == true, "Is no whitelister!");

        return _removeWhitelister(_adr);
    }

    function whitelistMember(address _adr)
        public
        whitelisterOnly
        returns (bool)
    {
        require(
            membershipList[_adr] == MembershipStatus.requesting ||
                (membershipList[_adr] == MembershipStatus.whitelistedByOne &&
                    firstWhiteLister[_adr] != msg.sender),
            "Invalid member status!"
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

            _totalCheckpoints.push(_add, 1);
            _voteCheckpoints[_adr].push(_add, 1);

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
        require(msg.value >= membershipFee, "Membership fee not covered!");

        nextMembershipFeePayment[msg.sender] = nextDueDate + 365 days;
        _wallet.payContribution{value: msg.value}(msg.sender);
    }

    function setMembershipFee(uint256 newMembershipFee)
        public
        representativeOnly
    {
        membershipFee = newMembershipFee;
    }

    function setRepresentative(address _adr) public returns (bool) {
        // TODO: require oder modifier einbauen, dass der sender vom verwalter der proposals kommt
        require(
            membershipList[_adr] == MembershipStatus.isMember,
            "Address is not a member!"
        );
        require(representative != _adr, "Address is the representative!");
        address oldRepresentative = representative;
        representative = _adr;
        emit ChangeInRepresentative(oldRepresentative, false);
        emit ChangeInRepresentative(representative, true);
        return true;
    }

    function removeMember(address _adr) public {
        require(
            membershipList[_adr] != MembershipStatus.nonMember,
            "Address is not a member!"
        );

        require(representative != _adr, "Representative cannot leave!");

        if (msg.sender != _adr) {
            require(
                msg.sender == representative,
                "Restricted to representative!"
            );
        }

        if (isWhitelister(_adr)) {
            _removeWhitelister(_adr);
        }

        if (membershipList[_adr] == MembershipStatus.isMember) {
            _totalCheckpoints.push(_subtract, 1);
            _voteCheckpoints[_adr].push(_subtract, 1);
        }

        membershipList[_adr] = MembershipStatus.nonMember;
        emit ChangeInMembershipStatus(
            _adr,
            uint256(MembershipStatus.nonMember)
        );
    }

    function getVotes(address account)
        public
        view
        virtual
        override
        returns (uint256)
    {
        return _voteCheckpoints[account].latest();
    }

    function getPastVotes(address account, uint256 blockNumber)
        public
        view
        virtual
        override
        returns (uint256)
    {
        return _voteCheckpoints[account].getAtProbablyRecentBlock(blockNumber);
    }

    function getPastTotalSupply(uint256 blockNumber)
        public
        view
        virtual
        override
        returns (uint256)
    {
        require(blockNumber < block.number, "Votes: block not yet mined");
        return _totalCheckpoints.getAtProbablyRecentBlock(blockNumber);
    }

    /* solhint-disable no-empty-blocks */
    function delegate(address delegatee) public virtual override {
        // doesnt need to anything
    }

    function delegateBySig(
        address delegatee,
        uint256 nonce,
        uint256 expiry,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external {
        // is fine to be empty
    }

    /* solhint-enable no-empty-blocks */

    function delegates(address account) external pure returns (address) {
        return account;
    }

    function _add(uint256 a, uint256 b) private pure returns (uint256) {
        return a + b;
    }

    function _subtract(uint256 a, uint256 b) private pure returns (uint256) {
        return a - b;
    }
}
/* solhint-enable not-rely-on-time */
