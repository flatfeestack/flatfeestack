# flatfeestack: pay-out

Service for interacting with the different blockchains. Supported are: Ethereum (ETH), NEO3 (NEO) and Tezos (XTZ).
Every blockchain has their own way so each implementation has some small pitfalls.

Every currency has implemented a `deploy`-method to deploy the smart contract on start up.
If a smart contract will be deployed depends on the `DEPLOY`-variable in the `.env` 

# ETH
The Ethereum payout uses a Go-binding which gets generated based on the `Flatfeestack.sol`.
Information about the toll can be found here:
https://geth.ethereum.org/docs/dapp/native-bindings

It can be generated with:
```
abigen --pkg main --sol Flatfeestack.sol --out ./contract.go 
```

# NEO
The NEO payout is based on two files:
- `PayoutNEO.manifest.json`
- `PayoutNeo.nef`

# XTZ
The Tezos payout happens in the `payout-nodejs`-service,
because of the missing possibility to interact with a Tezos smart-contract in go.
- (https://www.reddit.com/r/tezos/comments/qqccg7/go_library_to_interact_with_smart_contracts/)
- (https://github.com/blockwatch-cc/tzgo/issues/9)