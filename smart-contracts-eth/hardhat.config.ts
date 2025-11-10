import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "solidity-coverage";

const config: HardhatUserConfig = {
  solidity: {
    compilers: [
      {
        version: "0.8.25",
        settings: {
          evmVersion: "cancun",
          optimizer: { enabled: true, runs: 1000 },
        },
      },
      {
        version: "0.8.28",
        settings: {
          evmVersion: "cancun",
          optimizer: { enabled: true, runs: 1000 },
        },
      },
    ],
  },
  typechain: {
    outDir: "typechain-types",
    target: "ethers-v6",
    externalArtifacts: [
      "node_modules/@account-abstraction/contracts/core/*.json",
      "node_modules/@account-abstraction/contracts/accounts/*.json",
    ],
  },
  networks: {
    hardhat: { hardfork: 'cancun' },
  },
};

export default config;
