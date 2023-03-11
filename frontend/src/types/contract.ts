export const ABI = `[
	{
		"inputs": [],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint192",
				"name": "amount",
				"type": "uint192"
			},
			{
				"indexed": false,
				"internalType": "uint64",
				"name": "time",
				"type": "uint64"
			}
		],
		"name": "PaymentReleased",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "addr",
				"type": "address"
			}
		],
		"name": "balanceOf",
		"outputs": [
			{
				"internalType": "uint192",
				"name": "",
				"type": "uint192"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address[]",
				"name": "addresses",
				"type": "address[]"
			},
			{
				"internalType": "uint192[]",
				"name": "balancesWei",
				"type": "uint192[]"
			}
		],
		"name": "fill",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "release",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "addr",
				"type": "address"
			}
		],
		"name": "timeOf",
		"outputs": [
			{
				"internalType": "uint64",
				"name": "",
				"type": "uint64"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address[]",
				"name": "addresses",
				"type": "address[]"
			}
		],
		"name": "unclaimed",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`;
