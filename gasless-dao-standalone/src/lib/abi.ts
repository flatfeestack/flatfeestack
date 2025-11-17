import type { Abi } from "viem";

export const entryPointAbi = [
  {
    "inputs": [
      {
        "components": [
          { "internalType": "address", "name": "sender", "type": "address" },
          { "internalType": "bytes21", "name": "nonce", "type": "bytes21" },
          { "internalType": "bytes", "name": "initCode", "type": "bytes" },
          { "internalType": "bytes", "name": "callData", "type": "bytes" },
          { "internalType": "uint256", "name": "callGasLimit", "type": "uint256" },
          { "internalType": "uint256", "name": "verificationGasLimit", "type": "uint256" },
          { "internalType": "uint256", "name": "preVerificationGas", "type": "uint256" },
          { "internalType": "uint256", "name": "maxFeePerGas", "type": "uint256" },
          { "internalType": "uint256", "name": "maxPriorityFeePerGas", "type": "uint256" },
          { "internalType": "bytes", "name": "paymasterAndData", "type": "bytes" },
          { "internalType": "bytes", "name": "signature", "type": "bytes" }
        ],
        "internalType": "struct PackedUserOperation",
        "name": "userOp",
        "type": "tuple"
      }
    ],
    "name": "getUserOpHash",
    "outputs": [{ "internalType": "bytes32", "name": "", "type": "bytes32" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{ "internalType": "address", "name": "account", "type": "address" }],
    "name": "balanceOf",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  }
] as const;

export const daoAbi = [
  {
    type: "function",
    name: "propose",
    stateMutability: "nonpayable",
    inputs: [
      { name: "targets", type: "address[]" },
      { name: "values", type: "uint256[]" },
      { name: "calldatas", type: "bytes[]" },
      { name: "description", type: "string" }
    ],
    outputs: [{ name: "", type: "uint256" }]
  },
  {
    type: "function",
    name: "castVote",
    stateMutability: "nonpayable",
    inputs: [
      { name: "proposalId", type: "uint256" },
      { name: "support", type: "uint8" }
    ],
    outputs: []
  },
  {
    type: "function",
    name: "state",
    stateMutability: "view",
    inputs: [{ name: "proposalId", type: "uint256" }],
    outputs: [{ name: "", type: "uint8" }]
  },
  {
    type: "event",
    name: "ProposalCreated",
    anonymous: false,
    inputs: [
      { indexed: true, name: "proposalId", type: "uint256" },
      { indexed: true, name: "proposer", type: "address" },
      { indexed: false, name: "targets", type: "address[]" },
      { indexed: false, name: "values", type: "uint256[]" },
      { indexed: false, name: "calldatas", type: "bytes[]" },
      { indexed: false, name: "description", type: "string" }
    ]
  }
] as const;

export const FLATFEESTACK_NFT_ABI = [
  {
    "inputs": [
      { "internalType": "address", "name": "initialOwner", "type": "address" },
      { "internalType": "address", "name": "council1", "type": "address" },
      { "internalType": "address", "name": "council2", "type": "address" }
    ],
    "stateMutability": "nonpayable",
    "type": "constructor"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "internalType": "address", "name": "addr", "type": "address" },
      { "indexed": true, "internalType": "address", "name": "creator", "type": "address" }
    ],
    "name": "FlatFeeStackNFTCreated",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "internalType": "address", "name": "addr", "type": "address" },
      { "indexed": true, "internalType": "uint256", "name": "tokenId", "type": "uint256" },
      { "indexed": true, "internalType": "uint256", "name": "val", "type": "uint256" }
    ],
    "name": "MembershipPayed",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "internalType": "uint256", "name": "tokenId", "type": "uint256" },
      { "indexed": true, "internalType": "bool", "name": "status", "type": "bool" }
    ],
    "name": "CouncilSet",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "internalType": "uint256", "name": "membershipFee", "type": "uint256" },
      { "indexed": true, "internalType": "uint48", "name": "membershipPeriod", "type": "uint48" }
    ],
    "name": "MembershipSettingsSet",
    "type": "event"
  },
  {
    "inputs": [],
    "name": "MAX_UINT48",
    "outputs": [{ "internalType": "uint48", "name": "", "type": "uint48" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [
      { "internalType": "address", "name": "addr", "type": "address" },
      { "internalType": "uint256", "name": "index1", "type": "uint256" },
      { "internalType": "bytes", "name": "signature1", "type": "bytes" },
      { "internalType": "uint256", "name": "index2", "type": "uint256" },
      { "internalType": "bytes", "name": "signature2", "type": "bytes" }
    ],
    "name": "safeMint",
    "outputs": [{ "internalType": "uint256", "name": "tokenId", "type": "uint256" }],
    "stateMutability": "payable",
    "type": "function"
  },

  {
    "inputs": [{ "internalType": "uint256", "name": "tokenId", "type": "uint256" }],
    "name": "payMembership",
    "outputs": [],
    "stateMutability": "payable",
    "type": "function"
  },

  {
    "inputs": [{ "internalType": "uint256", "name": "tokenId", "type": "uint256" }],
    "name": "burn",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },

  {
    "inputs": [{ "internalType": "uint256", "name": "tokenId", "type": "uint256" }],
    "name": "tokenURI",
    "outputs": [{ "internalType": "string", "name": "", "type": "string" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [
      { "internalType": "address", "name": "council", "type": "address" },
      { "internalType": "uint256", "name": "index", "type": "uint256" }
    ],
    "name": "isCouncilIndex",
    "outputs": [{ "internalType": "bool", "name": "", "type": "bool" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [{ "internalType": "uint256", "name": "tokenId", "type": "uint256" }],
    "name": "isCouncil",
    "outputs": [{ "internalType": "bool", "name": "", "type": "bool" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [
      { "internalType": "uint256", "name": "tokenId", "type": "uint256" },
      { "internalType": "bool", "name": "status", "type": "bool" }
    ],
    "name": "setCouncil",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },

  {
    "inputs": [
      { "internalType": "uint256", "name": "_membershipFee", "type": "uint256" },
      { "internalType": "uint48", "name": "_membershipPeriod", "type": "uint48" }
    ],
    "name": "setMembershipSettings",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },

  {
    "inputs": [],
    "name": "membershipFee",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [],
    "name": "membershipPeriod",
    "outputs": [{ "internalType": "uint48", "name": "", "type": "uint48" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [],
    "name": "currentTokenId",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [],
    "name": "councilCount",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [
      { "internalType": "uint256", "name": "tokenId", "type": "uint256" }
    ],
    "name": "membershipPayed",
    "outputs": [{ "internalType": "uint48", "name": "", "type": "uint48" }],
    "stateMutability": "view",
    "type": "function"
  },

  {
    "inputs": [],
    "name": "pause",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },

  {
    "inputs": [],
    "name": "unpause",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },

  {
    "inputs": [],
    "name": "emergencyEth",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },

  {
    "inputs": [
      { "internalType": "address", "name": "token", "type": "address" }
    ],
    "name": "emergencyERC20",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },

  /* Standard ERC721 + Enumerable + Votes functions omitted for brevity */
  {
    "inputs": [],
    "name": "name",
    "outputs": [{ "internalType": "string", "name": "", "type": "string" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "symbol",
    "outputs": [{ "internalType": "string", "name": "", "type": "string" }],
    "stateMutability": "view",
    "type": "function"
  }
] as const;

export const SMART_ACCOUNT_ABI = [
  {
    type: "function",
    name: "execute",
    stateMutability: "nonpayable",
    inputs: [
      { name: "dest", type: "address" },
      { name: "value", type: "uint256" },
      { name: "func", type: "bytes" },
    ],
    outputs: [],
  },
  {
    type: "function",
    name: "nonce",
    stateMutability: "view",
    inputs: [],
    outputs: [
      { name: "", type: "uint256" }
    ],
  },
] as const satisfies Abi;