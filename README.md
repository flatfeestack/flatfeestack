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

There are different Docker compose profiles to boot up what you need to develop certain parts of the application.

* `platform`: Starts all services needed for the Flatfeestack Platform.
* `smart-contracts-eth`: Only starts the Ganache chain and the frontend, needed to develop the Flatfeestack DAO.
* `blockexplorer`: A Ethereum block explorer. Nice to have when you need to inspect what happens in the local chain.

Profiles can be used as follows:

```shell
docker compose --profile blockexplorer --profile smart-contracts-eth up --build
```

### Local development - Frontend with Hot Module Reload (HMR)
Especially while working on the frontend, HMR in local development enables you to instantly see your changes without having to rebuild the project or wait ages.

First a change in the caddy file (`caddy/Caddyfile`) is required.

```yaml
# change
reverse_proxy frontend:9085
# to
reverse_proxy host.docker.internal:9085
```

Afterwards docker can be started up again
```shell
docker compose --profile platform up --build

# Switch to a different tab / terminal and shut down the frontend
docker compose stop frontend
```

Move to your frontend folder (assuming you are in the root flatfeestack dir) and execute the new pnpm script
```shell
cd frontend
pnpm run hmr
```

Afterwards you can visit the same URL and use the frontend but this time with HMR.

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

