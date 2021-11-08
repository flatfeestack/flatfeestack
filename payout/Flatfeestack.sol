pragma solidity ^0.8.4;

contract PayoutEthEval {

    mapping(address => uint256) private teaMap;
    address public owner;

    constructor () {
        owner = msg.sender;
    }

    function deposit(uint256 amount) public payable {
        require(msg.value == amount, "Message value must match the provided parameter value.");
    }

    function changeOwner(address newOwner) public {
        require(msg.sender == owner, "No authorization.");
        owner = newOwner;
    }

    function getTotalEarnedAmount(address _dev) public view returns (uint256) {
        return teaMap[_dev];
    }

    function withdrawNotHashed(address payable _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) public {
        require(_tea > teaMap[_dev], "These funds have already been withdrawn.");
        bytes memory concatDevTea = abi.encodePacked(_dev, _tea);
        address signer = recoverSigner(concatDevTea, _v, _r, _s);
        if (signer == owner) {
            uint256 oldTea = teaMap[_dev];
            teaMap[_dev] = _tea;
            bool transfer = _dev.send(_tea - oldTea);
            require(transfer, "Transfer was not successful.");
        }
    }

    function withdrawHashedRequire(address payable _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) public {
        require(_tea > teaMap[_dev], "These funds have already been withdrawn.");
        bytes32 concatDevTeaHash = keccak256(abi.encodePacked(_dev, _tea));
        address signer = recoverSignerHash(concatDevTeaHash, _v, _r, _s);
        require(signer == owner, "Signature does not match owner and provided parameters.");
        uint256 oldTea = teaMap[_dev];
        teaMap[_dev] = _tea;
        bool transfer = _dev.send(_tea - oldTea);
        require(transfer, "Transfer was not successful.");
    }

    function withdrawHashed(address payable _dev, uint256 _tea, uint8 _v, bytes32 _r, bytes32 _s) public {
        require(_tea > teaMap[_dev], "These funds have already been withdrawn.");
        bytes32 concatDevTeaHash = keccak256(abi.encodePacked(_dev, _tea));
        address signer = recoverSignerHash(concatDevTeaHash, _v, _r, _s);
        if (signer == owner) {
            uint256 oldTea = teaMap[_dev];
            teaMap[_dev] = _tea;
            bool transfer = _dev.send(_tea - oldTea);
            require(transfer, "Transfer was not successful.");
        }
    }

    function batchPayout(address payable[] memory _devs, uint256[] memory _teas) public {
        require(msg.sender == owner, "No authorization.");
        require(_devs.length == _teas.length, "Arrays must have same length.");
        for (uint256 i = 0; i < _devs.length; i++) {
            uint256 oldTea = teaMap[_devs[i]];
            if (oldTea >= _teas[i]) {
                continue;
            }
            teaMap[_devs[i]] = _teas[i];
            _devs[i].send(_teas[i] - oldTea);
        }
    }

    function batchPayoutServiceFeeWithPayout(address payable[] memory _devs, uint256[] memory _teasForPayout, uint256 serviceFee) public {
        require(msg.sender == owner, "No authorization.");
        require(_devs.length == _teasForPayout.length, "Arrays must have same length.");
        for (uint256 i = 0; i < _devs.length; i++) {
            uint256 oldTea = teaMap[_devs[i]];
            if (oldTea >= _teasForPayout[i]) {
                continue;
            }
            teaMap[_devs[i]] = _teasForPayout[i] + serviceFee;
            _devs[i].send(_teasForPayout[i] - oldTea);
        }
    }

    function batchPayoutServiceFeeWithStore(address payable[] memory _devs, uint256[] memory _teasToStore, uint256 serviceFee) public {
        require(msg.sender == owner, "No authorization.");
        require(_devs.length == _teasToStore.length, "Arrays must have same length.");
        for (uint256 i = 0; i < _devs.length; i++) {
            uint256 oldTea = teaMap[_devs[i]];
            if (oldTea >= _teasToStore[i] - serviceFee) {
                continue;
            }
            teaMap[_devs[i]] = _teasToStore[i];
            _devs[i].send(_teasToStore[i] - oldTea - serviceFee);
        }
    }

    // helper methods

    function getConcat(address _dev, uint256 _tea) public pure returns (bytes memory) {
        return abi.encodePacked(_dev, _tea);
    }

    // todo: Consider inlining in withdraw method.
    function recoverSigner(bytes memory concatDevTea, uint8 v, bytes32 r, bytes32 s) public pure returns (address) {
        /*
        Signature is produced by signing a keccak256 hash with the following format:
        "\x19Ethereum Signed Message\n" + len(msg) + msg
        */
        bytes memory prefix = "\x19Ethereum Signed Message:\n106";
        bytes32 prefixedHash = keccak256(abi.encodePacked(prefix, concatDevTea));
        return ecrecover(prefixedHash, v, r, s);
    }

    function getConcatHash(address _dev, uint256 _tea) public pure returns (bytes32) {
        return keccak256(getConcat(_dev, _tea));
    }

    function recoverSignerHash(bytes32 concatDevTeaHash, uint8 v, bytes32 r, bytes32 s) public pure returns (address) {
        /*
        Signature is produced by signing a keccak256 hash with the following format:
        "\x19Ethereum Signed Message\n" + len(msg) + msg
        */
        bytes memory prefix = "\x19Ethereum Signed Message:\n66";
        bytes32 prefixedHash = keccak256(abi.encodePacked(prefix, concatDevTeaHash));
        return ecrecover(prefixedHash, v, r, s);
    }

}