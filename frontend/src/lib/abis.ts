export const nftAbi = [
  "function balanceOf(address owner) view returns (uint256)",
  "function tokenOfOwnerByIndex(address owner, uint256 index) view returns (uint256)",
  "function membershipPayed(uint256 tokenId) view returns (uint48)",
  "function isCouncil(uint256 tokenId) view returns (bool)",
];

export const daoAbi = [
  "function propose(address[] calldata targets, uint256[] calldata values, bytes[] calldata calldatas, string calldata description) returns (uint256)",
  "function state(uint256 proposalId) view returns (uint8)",
  "function castVote(uint256 proposalId, uint8 support) returns (uint256)",
  "function votingDelay() view returns (uint256)",
  "function votingPeriod() view returns (uint256)"
];

export const paymasterAbi = [
  "function ping() view returns (string)",
  "function deposit() payable",
  "function hasDeposit() view returns (uint256)"
];

export const entryPointAbi = [
  "function balanceOf(address) view returns (uint256)"
];