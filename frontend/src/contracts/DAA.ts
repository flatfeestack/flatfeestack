export const DAAABI = [
  "error Empty()",
  "event Initialized(uint8 version)",
  "event NewTimeslotSet(uint256 timeslot)",
  "event ProposalCanceled(uint256 proposalId)",
  "event ProposalCreated(uint256 proposalId, address proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 startBlock, uint256 endBlock, string description)",
  "event ProposalExecuted(uint256 proposalId)",
  "event QuorumNumeratorUpdated(uint256 oldQuorumNumerator, uint256 newQuorumNumerator)",
  "event VoteCast(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason)",
  "event VoteCastWithParams(address indexed voter, uint256 proposalId, uint8 support, uint256 weight, string reason, bytes params)",
  "function BALLOT_TYPEHASH() view returns (bytes32)",
  "function COUNTING_MODE() pure returns (string)",
  "function EXTENDED_BALLOT_TYPEHASH() view returns (bytes32)",
  "function castVote(uint256 proposalId, uint8 support) returns (uint256)",
  "function castVoteBySig(uint256 proposalId, uint8 support, uint8 v, bytes32 r, bytes32 s) returns (uint256)",
  "function castVoteWithReason(uint256 proposalId, uint8 support, string reason) returns (uint256)",
  "function castVoteWithReasonAndParams(uint256 proposalId, uint8 support, string reason, bytes params) returns (uint256)",
  "function castVoteWithReasonAndParamsBySig(uint256 proposalId, uint8 support, string reason, bytes params, uint8 v, bytes32 r, bytes32 s) returns (uint256)",
  "function execute(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) payable returns (uint256)",
  "function getNumberOfProposalsInVotingSlot(uint256 slotNumber) view returns (uint256)",
  "function getSlotsLength() view returns (uint256)",
  "function getVotes(address account, uint256 blockNumber) view returns (uint256)",
  "function getVotesWithParams(address account, uint256 blockNumber, bytes params) view returns (uint256)",
  "function hasVoted(uint256 proposalId, address account) view returns (bool)",
  "function hashProposal(address[] targets, uint256[] values, bytes[] calldatas, bytes32 descriptionHash) pure returns (uint256)",
  "function initialize(address _membership)",
  "function membershipContract() view returns (address)",
  "function name() view returns (string)",
  "function proposalDeadline(uint256 proposalId) view returns (uint256)",
  "function proposalSnapshot(uint256 proposalId) view returns (uint256)",
  "function proposalThreshold() pure returns (uint256)",
  "function proposalVotes(uint256 proposalId) view returns (uint256 againstVotes, uint256 forVotes, uint256 abstainVotes)",
  "function propose(address[] targets, uint256[] values, bytes[] calldatas, string description) returns (uint256)",
  "function quorum(uint256 blockNumber) view returns (uint256)",
  "function quorumDenominator() view returns (uint256)",
  "function quorumNumerator(uint256 blockNumber) view returns (uint256)",
  "function quorumNumerator() view returns (uint256)",
  "function relay(address target, uint256 value, bytes data) payable",
  "function setVotingSlot(uint256 blockNumber) returns (uint256)",
  "function slotCloseTime() view returns (uint256)",
  "function slots(uint256) view returns (uint256)",
  "function state(uint256 proposalId) view returns (uint8)",
  "function supportsInterface(bytes4 interfaceId) view returns (bool)",
  "function token() view returns (address)",
  "function updateQuorumNumerator(uint256 newQuorumNumerator)",
  "function version() view returns (string)",
  "function votingDelay() pure returns (uint256)",
  "function votingPeriod() pure returns (uint256)",
  "function votingSlots(uint256, uint256) view returns (uint256)",
];
