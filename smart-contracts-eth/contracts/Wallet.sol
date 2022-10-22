// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract Wallet is Initializable, OwnableUpgradeable {
    uint256 public totalBalance;
    uint256 public totalAllowance;

    uint256 private _knownSenderLength;
    address[] private _knownSender;

    mapping(address => uint256) public individualContribution;
    mapping(address => uint256) public allowance;
    mapping(address => uint256) public withdrawingAllowance;

    event IncreaseAllowance(address indexed Account, uint256 Amount);
    event AcceptPayment(address indexed Account, uint256 Amount);
    event WithdrawFunds(address indexed Account, uint256 Amount);

    modifier knownSender() {
        require(isKnownSender(msg.sender) == true, "only known senders");
        _;
    }

    function initialize() public initializer {
        __Ownable_init();
        addKnownSender(msg.sender);
        _knownSenderLength = 1;
    }

    receive() external payable {
        totalBalance += msg.value;
        individualContribution[msg.sender] += msg.value;
    }

    function addKnownSender(address _adr) public onlyOwner {
        if (isKnownSender(_adr) == false) {
            _knownSender.push(_adr);
            _knownSenderLength++;
        }
    }

    function isKnownSender(address _adr) public view returns (bool) {
        bool check = false;
        for (uint256 i = 0; i < _knownSenderLength; i++) {
            if (_knownSender[i] == _adr) {
                check = true;
                break;
            }
        }
        return check;
    }

    function removeKnownSender(address _adr) public onlyOwner {
        require(_adr != owner(), "Owner cannot be removed from known senders!");

        uint256 i;

        for (i = 0; i < _knownSenderLength - 1; i++) {
            if (_knownSender[i] == _adr) {
                break;
            }
        }

        if (i != _knownSenderLength - 1) {
            _knownSender[i] = _knownSender[_knownSenderLength - 1];
        }

        _knownSenderLength--;
    }

    function increaseAllowance(address _adr, uint256 _amount)
        public
        onlyOwner
        returns (bool)
    {
        require(
            (totalAllowance + _amount) <= totalBalance,
            "Cannot increase allowance over total balance of wallet!"
        );
        allowance[_adr] += _amount;
        totalAllowance += _amount;
        emit IncreaseAllowance(_adr, _amount);
        return true;
    }

    function payContribution(address _adr)
        public
        payable
        knownSender
        returns (bool)
    {
        uint256 _amount = msg.value;
        totalBalance += _amount;
        individualContribution[_adr] += _amount;

        emit AcceptPayment(_adr, _amount);

        return true;
    }

    function withdrawMoney(address payable _adr)
        public
        onlyOwner
        returns (bool)
    {
        require(allowance[_adr] > 0, "cannot withdraw without any allowance!");

        uint256 operatingAmount = allowance[_adr];
        withdrawingAllowance[_adr] = operatingAmount;
        allowance[_adr] -= operatingAmount;

        if (_adr.send(withdrawingAllowance[_adr]) == true) {
            totalAllowance -= withdrawingAllowance[_adr];
            totalBalance -= withdrawingAllowance[_adr];
            withdrawingAllowance[_adr] = 0;
            emit WithdrawFunds(_adr, totalAllowance);
            return true;
        } else {
            allowance[_adr] += withdrawingAllowance[_adr];
            withdrawingAllowance[_adr] = 0;
            return false;
        }
    }
}
