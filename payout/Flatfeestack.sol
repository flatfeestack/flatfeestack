pragma solidity ^0.8.4;

contract PayoutEth {

    /**
    * @dev Maps each address to its current total earned amount (tea).
    */
    mapping(address => uint256) public teaMap;

    /**
    * @dev The contract owner
    */
    address public owner;

    constructor () {
        owner = msg.sender;
    }

    receive() external payable {
    }

    /**
    * @dev Changes the owner of this contract.
    */
    function changeOwner(address newOwner) public onlyOwner() {
        owner = newOwner;
    }

    /**
    * @dev Gets the tea for the provided address.
    */
    function getTea(address _dev) public view returns (uint256) {
        return teaMap[_dev];
    }

    /**
    * @dev Sets the tea for the provided address. The oldTea must match the currently stored value, in order to verify
    * that no immediate withdrawal took place before this is executed.
    */
    function setTea(address _dev, uint256 oldTea, uint256 newTea) public onlyOwner() {
        require(oldTea == teaMap[_dev], "Stored tea is not equal to the provided oldTea.");
        require(newTea > teaMap[_dev], "Cannot set a lower value due to security reasons.");
        teaMap[_dev] = newTea;
    }

    /**
    * @dev Sets the teas for the provided addresses. The oldTeas must match the currently stored value, in order to
    * verify that no immediate withdrawal took place before this is executed. In case one value does not match, no
    * update is executed for that address.
    */
    function setTeas(address[] calldata _devs, uint256[] calldata oldTeas, uint256[] calldata newTeas) public onlyOwner() {
        require(_devs.length == newTeas.length, "Parameters must have same length.");
        for (uint256 i = 0; i < _devs.length; i++) {
            address dev = _devs[i];
            uint256 storedTea = teaMap[dev];
            uint256 newTea = newTeas[i];
            if ((oldTeas[i] == storedTea) && (newTea > storedTea)) {
                teaMap[dev] = newTea;
            }
        }
    }

    /**
    * @dev Withdraws the earned amount. The signature has to be created by the contract owner and the signed message
    * is the hash of the concatenation of the account and tea.
    *
    * @param _dev The address to withdraw to.
    * @param _tea The amount to withdraw.
    * @param _v The recovery byte of the signature.
    * @param _r The r value of the signature.
    * @param _s The s value of the signature.
    */
    function withdraw(address payable _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) public {
        require(_tea > teaMap[_dev], "These funds have already been withdrawn.");
        require(ecrecover(keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n66",
            keccak256(abi.encodePacked(_dev, _tea)))), _v, _r, _s) == owner,
            "Signature does not match owner and provided parameters.");
        uint256 oldTea = teaMap[_dev];
        teaMap[_dev] = _tea;
        // transfer reverts transaction if not successful.
        _dev.transfer(_tea - oldTea);
    }

    /**
    * @dev Pays out the earned amount for multiple addresses. The payment amount for each address is equal to the
    * difference of the provided new tea in the parameter and the currently stored value in the teaMap. After
    * calculating the payment amount, the value of the account in the teaMap is updated with the new tea.
    *
    * @param _devs The addresses to pay out to.
    * @param _teas The teas for the corresponding addresses.
    */
    function batchPayout(address payable[] memory _devs, uint256[] memory _teas) public onlyOwner() {
        require(_devs.length == _teas.length, "Arrays must have same length.");
        for (uint256 i = 0; i < _devs.length; i++) {
            address payable dev = _devs[i];
            uint256 oldTea = teaMap[dev];
            uint256 tea = _teas[i];
            if (tea <= oldTea) {
                continue;
            }
            teaMap[dev] = tea;
            // transfer reverts transaction if not successful.
            dev.transfer(tea - oldTea);
        }
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "No authorization.");
        _;
    }

}
