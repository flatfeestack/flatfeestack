// SPDX-License-Identifier: MIT

pragma solidity >=0.7.0 <0.8.0;

library SafeMath64 {
    /**
     * @dev Returns the addition of two unsigned integers, reverting on
     * overflow.
     *
     * Counterpart to Solidity's `+` operator.
     *
     * Requirements:
     *
     * - Addition cannot overflow.
     */
    function add(uint192 a, uint192 b) internal pure returns (uint192) {
        uint192 c = a + b;
        require(c >= a, "SafeMath: addition overflow");

        return c;
    }
}

contract Flatfeestack {
    using SafeMath64 for uint192;
    mapping(address => Balance) private balances;
    address private owner;

    event PaymentReleased(address to, uint192 amount, uint64 time);

    struct Balance {
        uint192 balanceWei;
        uint64 time;
    }

    constructor () {
        owner = msg.sender;
    }

    /**
     * @dev Fill contract with array of balances. This needs to be optimized to cost as less gas as possible
     * Currently, for the input, it costs 104684 gas
     *  - ["0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2","0xDA0bab807633f07f013f94DD0E6A4F96F8742B53","0x9D7f74d0C41E726EC95884E0e97Fa6129e3b5E99"], [1234,1235,1236] wei: 3705
     * calling hash: 158c29c7
     * @param addresses, balancesWei payouts to update
     */
    function fill(address[] memory addresses, uint192[] memory balancesWei) public payable {
        require(msg.sender == owner, "Only the owner can add new payouts");
        require(addresses.length == balancesWei.length, "Addresses and balances array must have the same length");

        uint256 sumWei;
        for (uint16 i=0; i < addresses.length; i++) { //unlikely to iterate over more than 65536 addresses
            balances[addresses[i]].balanceWei = balances[addresses[i]].balanceWei.add(balancesWei[i]);
            if(balances[addresses[i]].time == 0) {
                balances[addresses[i]].time = uint64(block.timestamp);
            }
            //impossible to overflow, no SafeMath here
            sumWei += uint256(balancesWei[i]);
        }
        if(sumWei != msg.value) {
            revert("Sum of balances is higher than paid amount");
        }
    }

     /**
     * @dev Triggers a transfer of the assigned funds.
     * total shares and their previous withdrawals.
     */
    function release() public payable {
        require(balances[msg.sender].balanceWei > 0, "PaymentSplitter: account has no balance");

        uint192 balanceWei = balances[msg.sender].balanceWei;
        uint64 time = balances[msg.sender].time;
        balances[msg.sender].balanceWei = 0;
        balances[msg.sender].time = 0;

        msg.sender.transfer(balanceWei);
        emit PaymentReleased(msg.sender, balanceWei, time);
    }

    /**
     * @dev Return balance
     * @return balance and the time it was first added
     */
    function balanceOf(address addr) public view returns (uint192, uint64) {
        return (balances[addr].balanceWei, balances[addr].time);
    }

    function unclaimed(address[] memory addresses) public {
        require(msg.sender == owner, "Only the owner can collect unclaimed payouts");
        for (uint16 i=0; i < addresses.length; i++) { //unlikely to iterate over more than 65536 addresses
            if (balances[addresses[i]].time + 365 days < block.timestamp) {
                balances[msg.sender].balanceWei = balances[msg.sender].balanceWei.add(balances[addresses[i]].balanceWei);
                balances[addresses[i]].balanceWei = 0;
            }
        }
    }
}
