// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "./Wallet.sol";

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/utils/CheckpointsUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/governance/utils/IVotesUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";

// we rely on time to track membership payments
// however, we don't care about second-level precision, as we deal with a much longer time period
// there is a good exaplanation about this on StackExchange https://ethereum.stackexchange.com/a/117874
/* solhint-disable not-rely-on-time */
contract Membership is Initializable, IVotesUpgradeable, OwnableUpgradeable {
    using CheckpointsUpgradeable for CheckpointsUpgradeable.History;

    enum MembershipStatus {
        nonMember,
        requesting,
        approvedByOne,
        isMember
    }

    Wallet private _wallet;

    mapping(address => uint256) public nextMembershipFeePayment;

    // used for IVotes
    mapping(address => CheckpointsUpgradeable.History) private _voteCheckpoints;
    CheckpointsUpgradeable.History private _totalCheckpoints;

    uint256 public minimumChairmen;
    uint256 public membershipFee;

    address[] public chairmen;
    address[] public members;

    mapping(address => MembershipStatus) internal membershipList;
    mapping(address => address) internal firstApproval;

    event ChangeInMembershipStatus(
        address indexed accountAddress,
        uint256 indexed currentStatus
    );

    event ChangeInChairman(
        address indexed concernedChairman,
        bool removedOrAdded
    );

    event ChangeInWalletAddress(
        address indexed oldWallet,
        address indexed newWallet
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

    modifier chairmenOnly() {
        require(isChairman(msg.sender) == true, "only chairmen");
        _;
    }

    function initialize(
        address _firstChairman,
        address _secondChairman,
        Wallet _walletContract
    ) public initializer {
        __Ownable_init();

        minimumChairmen = 2;
        membershipFee = 30000 wei;
        _wallet = _walletContract;
        emit ChangeInWalletAddress(address(0x0), address(_wallet));

        chairmen.push(_firstChairman);
        chairmen.push(_secondChairman);

        membershipList[_firstChairman] = MembershipStatus.isMember;
        membershipList[_secondChairman] = MembershipStatus.isMember;

        members.push(_firstChairman);
        members.push(_secondChairman);

        nextMembershipFeePayment[_firstChairman] = block.timestamp;
        nextMembershipFeePayment[_secondChairman] = block.timestamp;

        _totalCheckpoints.push(_add, 2);
        _voteCheckpoints[_firstChairman].push(_add, 1);
        _voteCheckpoints[_secondChairman].push(_add, 1);

        emit ChangeInMembershipStatus(
            _firstChairman,
            uint256(MembershipStatus.isMember)
        );

        emit ChangeInChairman(_firstChairman, true);

        emit ChangeInMembershipStatus(
            _secondChairman,
            uint256(MembershipStatus.isMember)
        );

        emit ChangeInChairman(_secondChairman, true);
    }

    function requestMembership() public nonMemberOnly returns (bool) {
        membershipList[msg.sender] = MembershipStatus.requesting;
        members.push(msg.sender);
        emit ChangeInMembershipStatus(
            msg.sender,
            uint256(MembershipStatus.requesting)
        );
        return true;
    }

    function getMembershipStatus(address _adr) public view returns (uint256) {
        return uint256(membershipList[_adr]);
    }

    function addChairman(address _adr) public onlyOwner returns (bool) {
        require(
            membershipList[_adr] == MembershipStatus.isMember,
            "A chairman must be a member"
        );
        require(isChairman(_adr) == false, "Is already chairman!");

        chairmen.push(_adr);
        emit ChangeInChairman(_adr, true);

        return true;
    }

    function removeChairman(address _adr) public returns (bool) {
        require(isChairman(_adr) == true, "Is no chairman!");
        require(
            this.getChairmenLength() > minimumChairmen,
            "Minimum chairmen not met!"
        );

        if (msg.sender != _adr) {
            _checkOwner();
        }

        uint256 i;

        for (i = 0; i < this.getChairmenLength() - 1; i++) {
            if (chairmen[i] == _adr) {
                break;
            }
        }

        if (i != this.getChairmenLength() - 1) {
            chairmen[i] = chairmen[this.getChairmenLength() - 1];
        }
        chairmen.pop();

        emit ChangeInChairman(_adr, false);

        return true;
    }

    function approveMembership(
        address _adr
    ) public chairmenOnly returns (bool) {
        require(
            membershipList[_adr] == MembershipStatus.requesting ||
                (membershipList[_adr] == MembershipStatus.approvedByOne &&
                    firstApproval[_adr] != msg.sender),
            "Invalid member status!"
        );
        if (membershipList[_adr] == MembershipStatus.requesting) {
            membershipList[_adr] = MembershipStatus.approvedByOne;
            firstApproval[_adr] = msg.sender;
            emit ChangeInMembershipStatus(
                _adr,
                uint256(MembershipStatus.approvedByOne)
            );
        } else {
            membershipList[_adr] = MembershipStatus.isMember;

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

        nextMembershipFeePayment[msg.sender] = block.timestamp + 365 days;
        _wallet.payContribution{value: msg.value}(msg.sender);

        if (nextDueDate == 0) {
            _totalCheckpoints.push(_add, 1);
            _voteCheckpoints[msg.sender].push(_add, 1);
        }
    }

    function setMembershipFee(uint256 newMembershipFee) external onlyOwner {
        membershipFee = newMembershipFee;
    }

    function setMinimumChairmen(uint256 newMinimumChairmen) external onlyOwner {
        require(newMinimumChairmen <= chairmen.length, "To few chairmen!");
        minimumChairmen = newMinimumChairmen;
    }

    function setNewWalletAddress(Wallet newWallet) external onlyOwner {
        address oldWallet = address(_wallet);
        _wallet = newWallet;
        emit ChangeInWalletAddress(oldWallet, address(newWallet));
    }

    function removeMember(address _adr) public {
        require(
            membershipList[_adr] != MembershipStatus.nonMember,
            "Address is not a member!"
        );

        if (msg.sender != _adr) {
            _checkOwner();
        }

        if (isChairman(_adr)) {
            removeChairman(_adr);
        }

        _removeMember(_adr);
    }

    function _removeMember(address _adr) private {
        if (
            membershipList[_adr] == MembershipStatus.isMember &&
            nextMembershipFeePayment[_adr] > 0
        ) {
            _totalCheckpoints.push(_subtract, 1);
            _voteCheckpoints[_adr].push(_subtract, 1);
        }

        delete firstApproval[_adr];
        membershipList[_adr] = MembershipStatus.nonMember;

        for (uint256 i = 0; i < members.length; i++) {
            if (members[i] == _adr) {
                members[i] = members[members.length - 1];
                members.pop();
                break;
            }
        }

        emit ChangeInMembershipStatus(
            _adr,
            uint256(MembershipStatus.nonMember)
        );
    }

    function removeMembersThatDidntPay() public {
        address[] memory toBeRemoved = new address[](members.length);
        uint256 toBeRemovedIndex = 0;
        for (uint256 i = 0; i < members.length; i++) {
            address member = members[i];
            uint256 nextPayment = nextMembershipFeePayment[member];
            if (nextPayment > 0 && block.timestamp > nextPayment) {
                if (!isChairman(member)) {
                    toBeRemoved[toBeRemovedIndex] = member;
                    toBeRemovedIndex++;
                }
            }
        }
        for (uint256 i = 0; i < toBeRemoved.length; i++) {
            _removeMember(toBeRemoved[i]);
        }
    }

    function getVotes(
        address account
    ) public view virtual override returns (uint256) {
        return _voteCheckpoints[account].latest();
    }

    function getPastVotes(
        address account,
        uint256 blockNumber
    ) public view virtual override returns (uint256) {
        return _voteCheckpoints[account].getAtProbablyRecentBlock(blockNumber);
    }

    function getPastTotalSupply(
        uint256 blockNumber
    ) public view virtual override returns (uint256) {
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

    function getFirstApproval(address _adr) external view returns (address) {
        return firstApproval[_adr];
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
/* solhint-enable not-rely-on-time */
