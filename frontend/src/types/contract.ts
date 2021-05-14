export const ABI = [
  {
    inputs: [],
    stateMutability: "nonpayable",
    type: "constructor",
  },
  {
    anonymous: false,
    inputs: [
      {
        indexed: false,
        internalType: "address",
        name: "to",
        type: "address",
      },
      {
        indexed: false,
        internalType: "uint256",
        name: "amount",
        type: "uint256",
      },
    ],
    name: "PaymentReleased",
    type: "event",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "address_",
        type: "address",
      },
    ],
    name: "balanceOf",
    outputs: [
      {
        internalType: "uint256",
        name: "",
        type: "uint256",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address[]",
        name: "addresses_",
        type: "address[]",
      },
      {
        internalType: "uint256[]",
        name: "balances_",
        type: "uint256[]",
      },
    ],
    name: "fill",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [],
    name: "release",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
];

export const plans = [
  {
    title: "Yearly",
    price: 365 * 330000 / 1000000,
    freq: 365,
    desc: "By paying yearly <b>" + (365 * 330000 / 1000000) + " USD</b>, you help us to keep payment processing costs low and more money will reach your sponsored projects"
  },
  {
    title: "Quarterly",
    price: 90 * 330000 / 1000000,
    freq: 90,
    desc: "You want to support Open Source software with a quarterly flat fee of <b>" + (90 * 330000 / 1000000) + " USD</b>"
  },
  {
    title: "Beta",
    price: 2 * 330000 / 1000000,
    freq: 2,
    desc: "You want to support Open Source software with a quarterly flat fee of <b>" + (2 * 330000 / 1000000) + " USD</b>"
  }
];
