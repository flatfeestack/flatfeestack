# flatfeestack: pay-out

Service for interacting with the different blockchains. Supported are: Ethereum (ETH), NEO3 (NEO) and Tezos (XTZ).
Every blockchain has their own way so each implementation has some small pitfalls.

Every currency has implemented a `deploy`-method to deploy the smart contract on start up.
If a smart contract will be deployed depends on the `[CURRENCY]_DEPLOY`-variable in the `.env` 

**Please fill out the `[CURRENCY]_PRIVATE_KEY` and the `[CURRENCY]_CONTRACT` for each currency before you start.**

# ETH
The Ethereum payout uses a Go-binding which gets generated based on the `Flatfeestack.sol`.
Information about the tool can be found here:
https://geth.ethereum.org/docs/dapp/native-bindings

It can be generated with:
```
abigen --pkg main --sol Flatfeestack.sol --out ./contract.go 
```

For the smart contract development have a look at:
https://github.com/flatfeestack/payout-eth-contracts
# NEO
The NEO payout is based on two files:
- `PayoutNEO.manifest.json`
- `PayoutNeo.nef`

These are generated from the smart contract development repository.
https://github.com/flatfeestack/payout-neo-contracts

## Useful links:
- [Contract address to NEO-Adddress Converter](https://neo.org/converter/index)
- [Blockchain Explorer](https://neo3.testnet.neotube.io/home)

# XTZ
The Tezos payout happens in the `payout-nodejs`-service,
because of the missing possibility to interact with a Tezos smart-contract in go.
- (https://www.reddit.com/r/tezos/comments/qqccg7/go_library_to_interact_with_smart_contracts/)
- (https://github.com/blockwatch-cc/tzgo/issues/9)