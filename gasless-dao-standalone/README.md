# Standalone FlatFeeStack DAO (Gasless)
Using a simple Svelte Frontend without a classical backend but only ETH contracts, which have to be deployed manually.

## Development: Build / Start / Stop (Local)

Use **Docker Compose** from the project root to start the standalone
frontend locally:

``` bash
docker compose up
```

To stop:

``` bash
docker compose down
```

------------------------------------------------------------------------

## Ethereum Smart Contracts

All contracts are located in the `/contracts` directory:

-   **FlatFeeStackNFTandDAO.sol**\
    Contains both the NFT contract and the DAO governance contract. When deploying the DAO, it will internally deploy the NFT contract with the DAO contract being the owner.

-   **FlatFeeStackDAOPaymaster.sol**\
    The ERC-4337-compatible paymaster contract. It reads membership
    information from the deployed NFT contract.

### Deployment Order

1.  **Deploy `FlatFeeStackDAO`**
    -   Provide **Council 1** address
    -   Provide **Council 2** address
2.  **Deploy `FlatFeeStackDAOPaymaster`**
    -   Provide the **Entrypoint address** (from the ERC-4337 bundler)
    -   Provide the **NFT contract address** (`FlatFeeStackDAO.token()`)

### Configuration
The relevant addresses for this prototype are temporarily stored in `/src/config.ts` and have to be adjusted as soon as own contracts are deployed or the architecture changes.

### Paymaster Funding

The paymaster must be **manually funded** at the Entrypoint and **kept
above zero** to avoid failed UserOperations.\
Automatic refilling from membership fees is **not implemented yet**.

------------------------------------------------------------------------

## New Members

New membership NFTs can be minted using signatures from *both council
members*.\
In this prototype, NFT minting can be triggered using a `#debug` flag in
the URL.

After minting:

-   The new address receives an NFT and becomes **eligible for paymaster
    sponsorship**.
-   Membership is initially **inactive** until the membership fee is
    paid.
-   The member should pay the fee **gaslessly via the frontend** to
    activate their membership and gain voting rights.

------------------------------------------------------------------------

## Paymaster Logic

The paymaster sponsors gas for any EOA or smart account (SimpleAccount)
that holds a membership NFT.

Even if membership is **expired**, **inactive**, or **not yet
activated**, sponsorship still applies to simplify onboarding and reduce
friction for new members.

------------------------------------------------------------------------

## Possible Improvements
-   Automatic Paymaster Fund refill
-   Deployment Script
-   Local hardhat Tests
-   Frontend adjustments and add more functions
-   Config of Addresses