# FlatFeeStack DAA

This project contains the source code for the FlatFeeStack DAA on the Ethereum blockchain.
The project is based on [HardHat](https://hardhat.org/).

## Development

Make sure you have Node v16 up and running, best with NVM.

```shell
brew install nvm
nvm install 16
nvm use
```

Then, install the dependencies and check if the tests run.

```shell
npm i
npm run hardhat:test
```

## Deployment

We use the community-maintained `hardhat-deploy` plugin.

### To local environment

When working on frontend of DAA, which is part of the [main FlatFeeStack frontend](https://github.com/flatfeestack/frontend), we recommend to setup a local Ethereum chain with [Ganache](https://trufflesuite.com/ganache/).

This might seem silly, as `hardhat` is installed in this project and would also provide to run an Ethereum node. However, `hardhat` does not support running a persistent blockchain. Additionally, we install ganache into the global NPM environment so it's not specifically tied to the DAA project.

First, install the ganache-cli.

```shell
npm i -g ganache
```

Then, start the chain:

```shell
mkdir ganache
npx ganache --database.dbPath "./ganache" --logging.verbose --wallet.accounts "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80,100000000000000000000" --wallet.accounts "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d,100000000000000000000" --wallet.accounts "0xdbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97,100000000000000000000" --wallet.accounts "0xa267530f49f8280200edf313ee7af6b827f2a8bce2897751d06a843f644967b1,100000000000000000000" --wallet.accounts "0xdf57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e,100000000000000000000"
```

- This creates five accounts on the chain with 100 ETH each.
- The accounts are ordered as follows:
  - `0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80` is the representative of the DAA.
  - `0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d` is one whitelister.
  - `0xdbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97` is the second whitelister.
  - `0xa267530f49f8280200edf313ee7af6b827f2a8bce2897751d06a843f644967b1` is a regular member.
  - `0xdf57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e` has no privileges in the DAA.
- Those accounts can be imported into MetaMask so you can interact with the smart contract.

Deploy the smart contracts:

```shell
npm run hardhat:deploy -- --network localhost
```

Deployment is necessary each time the contracts change.

Additionally, if you run this setup the first time, you need to run a script that confirms the reserved member address:

```shell
npm run hardhat:script -- --network localhost scripts/addMember.ts
```

Now, you can export the ABIs of the smart contracts and the addresses of the proxies to the frontend:

```shell
FRONTEND_PATH="../frontend" npm run hardhat:script -- scripts/exportInterfacesToFrontend.ts
```

- The contracts' ABI will be written to `src/contracts`.
- The contracts' addresses will be written to a `.env` file. You can retrieve them via `VITE_${Contract name}_CONTRACT_ADDRESS` in the frontend.
- Always restart vite after a new interface export

# MetaMask

Make sure you are connect you wallet to localhost
