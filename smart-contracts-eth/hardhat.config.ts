import "@nomicfoundation/hardhat-toolbox";
import "@openzeppelin/hardhat-upgrades";
import "hardhat-deploy";
import "hardhat-deploy-ethers";
import { HardhatUserConfig } from "hardhat/config";
import "solidity-coverage";

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
    firstCouncilMember: {
      1337: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
    },
    secondCouncilMember: {
      1337: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
    },
    member: {
      1337: "0xFABB0ac9d68B0B445fB7357272Ff202C5651694a",
    },
    payoutERC20Contract: {
      1337: null,
    },
  },
  networks: {
    localhost: {
      live: false,
      saveDeployments: true,
      tags: ["local"],
    },
  },
};

export default config;
