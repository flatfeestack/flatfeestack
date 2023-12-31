# FlatFeeStack DAO and Payout Smart Contracts

This repository contains the following Ethereum smart contracts: ```FlatFeeStackNFT``` (Non-Fungible Token) and ```FlatFeeStackDAO``` (Decentralized Autonomous Organization), written in Solidity.

## FlatFeeStackNFT

The ```FlatFeeStackNFT``` contract is based on OpenZeppelin's ERC721 token standard and includes additional features like enumerability, pausability, burnability, and voting rights via the ERC721Votes extension. The NFTs are minted with a flat membership fee for a specified period. Every owner of an NFT can renew their membership. If a member doesn't pay, its NFT can be burned. The member can cancel within the membership period, their NFT will be burned and become non-member and can't vote in DAO decisions until they rejoin with the required payment with a new NFT. The contract also includes two council members who have special privileges: they can mint new tokens (to be paid by the new owners). The owner of the contract is the DAO, and the DAO can change the membership settings (fee or period), and execute certain DAO decisions from proposal after a voting delay.

## FlatFeeStackDAO

The ```FlatFeeStackDAO``` contract is based on OpenZeppelin's Governor Bravo standard which implements on-chain governance via proposals and votes. The DAO uses the same two council members. Every member can propose and vote on decisions (like changing bylaws or membership settings) or execute any smart contracts. It also includes a quorum function to ensure that at least 20% of all tokens holders participate in voting for a proposal to be valid.

## Installation

For running the contracts, install bun (at least: 1.0.20)

```
pnpm install
npx hardhat compile
npx hardhat test
```
## Deploy Contracts

Deployming can be done via remix. Copy/paste the smart contracts, compile and deploy.
