pragma solidity >=0.7.0 <0.8.0;


contract Flatfeestack {

    mapping(address => uint256) balances;
    address private _owner;

    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event PaymentReleased(address to, uint256 amount);



    /**
     * @dev Initializes the contract setting the deployer as the initial owner.
     */
    constructor () {
        _owner = msg.sender;
        emit OwnershipTransferred(address(0), msg.sender);
    }

    /**
     * @dev Store value in variable
     * @param addresses_, balances_ payouts to update
     */
    function fill(address[] memory addresses_, uint256[] memory balances_) public payable {
        require(msg.sender == _owner, "Only the owner can add new payouts");
        require(addresses_.length == balances_.length, "Addresses and balances array must have the same length");
        uint256 sum;

        for (uint i=0; i < addresses_.length; i++){
            balances[addresses_[i]] += balances_[i];
            sum = sum + balances_[i];
        }
        if(sum > msg.value) {
            revert("Sum of balances is higher than paid amount");
        }
    }

     /**
     * @dev Triggers a transfer to `account` of the amount of Ether they are owed, according to their percentage of the
     * total shares and their previous withdrawals.
     */
    function release() public virtual {
        require(balances[msg.sender] > 0, "PaymentSplitter: account has no balance");
        uint256 _balance = balances[msg.sender];
        balances[msg.sender] = 0;
        msg.sender.transfer(_balance);
        emit PaymentReleased(msg.sender, _balance );
    }

    /**
     * @dev Return balance
     * @return value of 'number'
     */
    function balanceOf(address address_) public view returns (uint256){
        return balances[address_];
    }
}