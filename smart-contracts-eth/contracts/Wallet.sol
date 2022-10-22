// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract Wallet is Initializable, OwnableUpgradeable {
    uint256 public totalBalance;
    uint256 public totalAllowance;

    mapping(address => uint256) public individualContribution;
    mapping(address => uint256) public allowance;
    mapping(address => uint256) public withdrawingAllowance;

    event IncreaseAllowance(
        address indexed Account,
        uint256 Amount,
        uint256 Timestamp
    );
    event AcceptPayment(
        address indexed Account,
        uint256 Amount,
        uint256 Timestamp
    );
    event WithdrawFunds(
        address indexed Account,
        uint256 Amount,
        uint256 Timestamp
    );

    function initialize() public initializer {
        __Ownable_init();
    }

    receive() external payable {
        totalBalance += msg.value;
        individualContribution[msg.sender] += msg.value;
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
        emit IncreaseAllowance(_adr, _amount, block.timestamp);
        return true;
    }

    function payContribution(address _adr)
        public
        payable
        onlyOwner
        returns (bool)
    {
        uint256 _amount = msg.value;
        totalBalance += _amount;
        individualContribution[_adr] += _amount;

        emit AcceptPayment(_adr, _amount, block.timestamp);

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
            emit WithdrawFunds(_adr, totalAllowance, block.timestamp);
            return true;
        } else {
            allowance[_adr] += withdrawingAllowance[_adr];
            withdrawingAllowance[_adr] = 0;
            return false;
        }
    }
}
