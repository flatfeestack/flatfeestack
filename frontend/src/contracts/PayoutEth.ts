export const PayoutEthABI = [
  "event Initialized(uint8 version)",
  "event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)",
  "function getClaimableAmount(bytes32 userId, uint256 totalPayOut) view returns (uint256)",
  "function getContractBalance() view returns (uint256)",
  "function getPayedOut(bytes32 userId) view returns (uint256)",
  "function initialize()",
  "function owner() view returns (address)",
  "function payedOut(bytes32) view returns (uint256)",
  "function renounceOwnership()",
  "function sendRecover(address receiver, uint256 amount)",
  "function transferOwnership(address newOwner)",
  "function withdraw(address dev, bytes32 userId, uint256 totalPayOut, uint8 v, bytes32 r, bytes32 s)",
];
