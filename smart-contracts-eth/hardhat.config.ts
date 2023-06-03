require("dotenv").config();

import "@nomicfoundation/hardhat-toolbox";
import "@openzeppelin/hardhat-upgrades";
import "hardhat-deploy";
import "hardhat-deploy-ethers";
import { HardhatUserConfig, subtask } from "hardhat/config";
import "solidity-coverage";
import { TASK_COMPILE_SOLIDITY_GET_SOURCE_PATHS } from "hardhat/builtin-tasks/task-names";
import path from "path";

subtask(
  TASK_COMPILE_SOLIDITY_GET_SOURCE_PATHS,
  async (_, { config }, runSuper) => {
    const paths = await runSuper();

    return paths.filter((solidityFilePath: any) => {
      const relativePath = path.relative(
        config.paths.sources,
        solidityFilePath
      );

      return relativePath !== "DAO2.sol" && relativePath !== "SBT.sol";
    });
  }
);

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.17",
    settings: {
      optimizer: {
        enabled: true,
        runs: 5,
      },
    },
  },
  namedAccounts: {
    daoContractDeployer: {
      goerli: "0xaC37Eb0d57f261AB95D3c65B8E8D93a60c128F50",
      1337: "0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199",
    },
    payoutEthDeployer: {
      goerli: "0xDba01b34D04789241B2a4B98295ad10ACA0C1339",
      1337: "0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199",
    },
    payoutUsdcDeployer: {
      goerli: "0x77a60DBD8605381b725b41993CE15e21AFA96b2a",
      main: "0x3AfDb92ec96b2F8a689B1b4B49F1d2a6AB16d4Dd",
      1337: "0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199",
    },
    firstCouncilMember: {
      goerli: "0xa879cA79d2702Df9BC51fc111d656ebd5342b067",
      1337: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
    },
    secondCouncilMember: {
      goerli: "0x30afB07D4e2c44ac362cCc89965B5d329Cabc4a0",
      1337: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
    },
    member: {
      1337: "0xFABB0ac9d68B0B445fB7357272Ff202C5651694a",
    },
    usdcTokenAddress: {
      goerli: "0x07865c6E87B9F70255377e024ace6630C1Eaa37F",
      main: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
      1337: null,
    },
  },
  networks: {
    goerli: {
      url: `https://eth-goerli.alchemyapi.io/v2/${process.env.ALCHEMY_GOERLI_API_KEY}`,
      accounts: [
        process.env.GOERLI_DAO_DEPLOYER_KEY ||
          "df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e",
        process.env.GOERLI_PAYOUT_ETH_DEPLOYER_KEY ||
          "df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e",
        process.env.GOERLI_PAYOUT_USDC_DEPLOYER_KEY ||
          "df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e",
      ],
    },
    localhost: {
      live: false,
      saveDeployments: true,
      tags: ["local"],
    },
    main: {
      url: `https://eth-mainnet.g.alchemy.com/v2/${process.env.ALCHEMY_MAIN_API_KEY}`,
      accounts: [
        process.env.MAIN_PAYOUT_USDC_DEPLOYER_KEY ||
          "df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e",
      ],
    },
  },
};

export default config;
