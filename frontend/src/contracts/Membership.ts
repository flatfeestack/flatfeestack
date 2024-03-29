export const MembershipABI = [
  "event ChangeInCouncilMember(address indexed concernedCouncilMember, bool removedOrAdded)",
  "event ChangeInMembershipStatus(address indexed accountAddress, uint256 indexed currentStatus)",
  "event ChangeInWalletAddress(address indexed oldWallet, address indexed newWallet)",
  "event DelegateChanged(address indexed delegator, address indexed fromDelegate, address indexed toDelegate)",
  "event DelegateVotesChanged(address indexed delegate, uint256 previousBalance, uint256 newBalance)",
  "event Initialized(uint8 version)",
  "event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)",
  "function addCouncilMember(address _adr) returns (bool)",
  "function approveMembership(address _adr) returns (bool)",
  "function councilMembers(uint256) view returns (address)",
  "function delegate(address delegatee)",
  "function delegateBySig(address delegatee, uint256 nonce, uint256 expiry, uint8 v, bytes32 r, bytes32 s)",
  "function delegates(address account) pure returns (address)",
  "function getCouncilMembersLength() view returns (uint256)",
  "function getFirstApproval(address _adr) view returns (address)",
  "function getMembersLength() view returns (uint256)",
  "function getMembershipStatus(address _adr) view returns (uint256)",
  "function getPastTotalSupply(uint256 blockNumber) view returns (uint256)",
  "function getPastVotes(address account, uint256 blockNumber) view returns (uint256)",
  "function getVotes(address account) view returns (uint256)",
  "function initialize(address _firstCouncilMember, address _secondCouncilMember, address _walletContract)",
  "function isCouncilMember(address _adr) view returns (bool)",
  "function lockMembership()",
  "function members(uint256) view returns (address)",
  "function membershipActive() view returns (bool)",
  "function membershipFee() view returns (uint256)",
  "function minimumCouncilMembers() view returns (uint256)",
  "function nextMembershipFeePayment(address) view returns (uint256)",
  "function owner() view returns (address)",
  "function payMembershipFee() payable",
  "function removeCouncilMember(address _adr) returns (bool)",
  "function removeMember(address _adr)",
  "function removeMembersThatDidntPay()",
  "function renounceOwnership()",
  "function requestMembership() returns (bool)",
  "function setMembershipFee(uint256 newMembershipFee)",
  "function setMinimumCouncilMembers(uint256 newMinimumCouncilMembers)",
  "function setNewWalletAddress(address newWallet)",
  "function transferOwnership(address newOwner)",
];
