# FlatFeeStack DAO and Payout Smart Contracts

This repository contains the following Ethereum smart contracts: ```FlatFeeStackNFT``` (Non-Fungible Token), ```FlatFeeStackDAO``` (Decentralized Autonomous Organization), ```Payout``` (tracking paid amount), and ```USDC_test``` for testcases written in Solidity.

## FlatFeeStackNFT

The ```FlatFeeStackNFT``` contract is based on OpenZeppelin's ERC721 token standard and includes additional features like enumerability, pausability, burnability, and voting rights via the ERC721Votes extension. The NFTs are minted with a flat membership fee for a specified period. Every owner of an NFT can renew their membership. If a member doesn't pay, its NFT can be burned. The member can cancel within the membership period, their NFT will be burned and become non-member and can't vote in DAO decisions until they rejoin with the required payment with a new NFT. The contract also includes two council members who have special privileges: they can mint new tokens (to be paid by the new owners). The owner of the contract is the DAO, and the DAO can change the membership settings (fee or period), and execute certain DAO decisions from proposal after a voting delay.

## FlatFeeStackDAO

The ```FlatFeeStackDAO``` contract is based on OpenZeppelin's Governor Bravo standard which implements on-chain governance via proposals and votes. The DAO uses the same two council members. Every member can propose and vote on decisions (like changing bylaws or membership settings) or execute any smart contracts. It also includes a quorum function to ensure that at least 20% of all tokens holders participate in voting for a proposal to be valid.

## Payout

Payout.sol defines two Ethereum smart contracts, PayoutEth and PayoutERC20, both extending from a base contract. The base contract is designed to handle payouts, specifically tracking and managing the amounts paid out to various user IDs.

In the Base contract, a mapping named payedOut is used to track the amount already paid to each user ID. The contract provides functions to get the paid and claimable amounts for a user, based on their total payout and what has already been paid. The calculateWithdraw function is used to calculate the amount to be withdrawn. It requires a valid signature from the contract owner to proceed, ensuring security and owner authorization for withdrawals. The sendRecover and sendRecoverToken functions allow the contract owner to recover Ether or ERC20 tokens from the contract in exceptional circumstances.

The PayoutEth contract is specifically for handling payouts in Ether (ETH). The PayoutERC20 contract functions similarly but is designed for handling ERC20 token payouts. Both contracts requiring owner signatures for withdrawals and using safe transfer methods for both Ether and ERC20 tokens.

## Installation

For running the contracts, install bun (at least: 1.0.20)

```
pnpm install
npx hardhat compile
npx hardhat test
npx hardhat coverage
```
## Deploy Contracts

Deployming can be done via remix. Copy/paste the smart contracts, compile and deploy.
