# Flatfeestack Installation
This repo combines all Flatfeestack packages using `docker-compose`.

## Build / Start / Stop (local development)

```shell script
# Create example .env files
cp analyzer/.example.env analyzer/.env
cp backend/.example.env backend/.env
cp forum/.example.env forum/.env
cp fastauth/.example.env fastauth/.env
cp payout/.example.env payout/.env
echo "HOST=http://localhost:8080" >> caddy/.env

echo "POSTGRES_PASSWORD=password" > db/.env
echo "POSTGRES_USER=postgres" >> db/.env
echo "POSTGRES_DB=flatfeestack" >> db/.env
```

## Register a user for the platform

To register a user:

1. Open `localhost:8080` and register a new user.
2. Connect to the database: `docker exec -it flatfeestack-db-1 psql -d flatfeestack -U postgres`
3. Show content of the `auth` table: `TABLE auth;`.
4. Get the token for the user you just registered.
5. Exit the PSQL session and confirm the user registration: `curl "http://localhost:9081/confirm/signup/{email}/{token}"`

If you want to stop, and clean everything up:

```shell script
docker-compose down -v
```

## ETH smart contracts

The development environment for the smart contracts needs a separate setup.

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

If you start the docker compose file with the `smart-contracts-eth` profile, a local Ganache chain gets started with the following properties.

- The four accounts on the chain have 100 ETH each.
- The accounts are ordered as follows:
  - `0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80` is the first council member of the DAO.
  - `0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d` is the second council member.
  - `0xa267530f49f8280200edf313ee7af6b827f2a8bce2897751d06a843f644967b1` is a regular member.
  - `0xdf57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e` has no privileges in the DAO, but is the deployer of all smart contracts.
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

Then you can run this script to add voting slots and proposals to the chain

```shell
npm run hardhat:script -- --network localhost scripts/addSlots.ts
```

If you want to fast-forward your chain can run this script.

```shell
npm run hardhat:script -- --network localhost scripts/mineBlocks.ts
```

Now, you can export the ABIs of the smart contracts to the frontend:

```shell
npm run hardhat:script -- --network localhost scripts/exportAbisToFrontend.ts
```

- The contracts' ABI will be written to `src/contracts`.
- When you access the frontend, make sure you connect MetaMask to the `localhost` network.

Also export the contract addresses to the payout service.

```shell
npm run hardhat:script -- --network localhost scripts/exportContractAddressesToPayout.ts
```

### Payout contracts and payout service

The payout service retrieves the current balance from the payout contracts to expose it to Prometheus. In case you update the contracts, you need to regenerate the contract code in the project.

You need a tool called `abigen` to generate Go code from Solidity ABIs. Make sure you have `solc` and `protoc` installed globally before building the tool:

```shell
git clone https://github.com/ethereum/go-ethereum.git
cd go-ethereum
make devtools
```

Next, you need the ABI for the `PayoutBase` contract. After compiling the contracts with `npm run hardhat:compile`, within `smart-contracts-eth/artifacts/contracts/PayoutEth.sol/PayoutEth.json`, there is a section with `abi`. Copy only this part to a separate file (the example below copies it into the same directory named `PayoutBase.abi`). Then navigate to the `payout` service directory and execute:

```shell
abigen --abi ../smart-contracts-eth/artifacts/contracts/PayoutBase.sol/PayoutBase.abi --pkg contracts --type PayoutBase --out contracts/PayoutBase.go
```

## NEO smart contract

There is a version of the payout contract for NEO. To start development, you need Java 8 and Docker to run the tests.

Running tests works with Gradle:

```shell
./gradlew test
```

To get a file that can be deployed to a NEO blockchain, execute the compile command:

```shell
./gradlew neow3jCompile
```

The resulting NEF file and manifest can be found in `smart-contracts-neo/build/neow3j`.

This NEF file can be deployed to a local NEO blockchain. Easiest way is to install VS Code and the [NEO blockchain toolkit](https://marketplace.visualstudio.com/items?itemName=ngd-seattle.neo-blockchain-toolkit). The NEO blockchain toolkit requires to have the [.NET SDK v6 installed](https://dotnet.microsoft.com/en-us/).

The NEO blockchain toolkit adds a new tab to your VS Code, where you have a button to start a NEO express chain. There is a [video tutorial available](https://ngdenterprise.com/neo-tutorials/quickstart1.html) that explains the functions of the toolkit.

Transfer some assets from the genesis block to Alice's wallet, either from the UI or with the CLI:

```shell
neoxp transfer 100 GAS genesis alice
```

Then, execute the file `NeoExpressDeployment.java`, ideally using IntelliJ (no clue how to do this from the terminal ...).

The script will yield you a `Deployment Transaction Hash`. Copy this hash and convert into a wallet import format (WIF). There is a website available to do this. Just copy the hash and press `Priv → WIF`. Place the result in your `.env` for the payout service at `NEO_CONTRACT`.

## Networking

This repo includes a caddy server to create reverse proxies to the different packages:

**/** --> Frontend

**/auth/*** --> Authentication Service

**/analyzer/*** --> Analysis Engine

## Env

Sample .env

```
POSTGRES_PASSWORD=password
POSTGRES_USER=postgres
POSTGRES_DB=flatfeestack
```

## Tezos sandbox

Get wallet information on private blockchain
```
docker exec -it flatfeestack-flextesa flobox info
```
# Monitoring

Everything who is related to monitoring is stored in the folder `monitoring`.

For grafana to be able to read from the database, a user with the following permissions needs to be created:
```
CREATE USER grafanareader WITH PASSWORD 'password';
GRANT USAGE ON SCHEMA public TO grafanareader;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO grafanareader;
```
## Local

To start the monitoring stack, run the following command:
```shell
docker compose -f monitoring/docker-compose.yml up -d
```

## Prod
On the productive environment, the monitoring stack is deployed on a DigitalOcean droplet.
See `.github/workflows/deploy-monitoring.yml` for the deployment script.

# PROD deployment
PROD deployment can be made with the `deploy-to-production.yaml`.
Deployment is made to the DigitalOcean App Platform.
The app spec for this is in the `app-spec.yaml`

