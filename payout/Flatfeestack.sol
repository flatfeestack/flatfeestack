pragma solidity ^0.8.4;

contract PayoutEth {

    // Maps the developer's addresses to their total earned amounts (teas).
    mapping(address => uint256) public teaMap;
    address public owner;

    constructor () {
        owner = msg.sender;
    }

    receive() external payable {
    }

    // Changes the contract owner.
    // Keep in mind, that changing the contract owner invalidates all previously generated signatures for the withdraw
    // method.
    function changeOwner(address newOwner) public onlyOwner() {
        owner = newOwner;
    }

    // Returns the value in the teaMap for the provided address.
    function getTea(address _dev) public view returns (uint256) {
        return teaMap[_dev];
    }

    // Sets the value for the provided address to newTea. Requires the oldTea to match and the newTea to be greater
    // than the current value.
    function setTea(address _dev, uint256 oldTea, uint256 newTea) public onlyOwner() {
        require(oldTea == teaMap[_dev], "Stored tea is not equal to the provided oldTea.");
        require(newTea > teaMap[_dev], "Cannot set a lower value due to security reasons.");
        teaMap[_dev] = newTea;
    }

    // Sets the values for the provided addresses to the corresponding newTeas.
    // If the corresponding oldTea does not match the current value, no change is applied for this address.
    function setTeas(address[] calldata _devs, uint256[] calldata oldTeas, uint256[] calldata newTeas)
        public onlyOwner()
    {
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

    // Transfers the amount _tea to address _dev, if the provided signature values match the created message and the
    // contract owner.
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

    // Updates the teaMap with the _teas for the corresponding addresses _devs and transfers the difference of _teas
    // to the current value in the teaMap to the corresponding _devs.
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
