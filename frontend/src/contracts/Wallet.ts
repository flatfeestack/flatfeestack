export const WalletABI = [
  "event AcceptPayment(address indexed account, uint256 amount)",
  "event IncreaseAllowance(address indexed account, uint256 amount)",
  "event Initialized(uint8 version)",
  "event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)",
  "event WithdrawFunds(address indexed account, uint256 amount)",
  "function addKnownSender(address _adr)",
  "function allowance(address) view returns (uint256)",
  "function increaseAllowance(address _adr, uint256 _amount) returns (bool)",
  "function individualContribution(address) view returns (uint256)",
  "function initialize()",
  "function isKnownSender(address _adr) view returns (bool)",
  "function owner() view returns (address)",
  "function payContribution(address _adr) payable returns (bool)",
  "function removeKnownSender(address _adr)",
  "function renounceOwnership()",
  "function totalAllowance() view returns (uint256)",
  "function totalBalance() view returns (uint256)",
  "function transferOwnership(address newOwner)",
  "function withdrawMoney(address _adr) returns (bool)",
  "function withdrawingAllowance(address) view returns (uint256)",
];
