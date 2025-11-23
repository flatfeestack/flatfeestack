require("@typechain/hardhat");

/** @type import("hardhat/config").HardhatUserConfig */
const config = {
  solidity: {
    compilers: [
      {
        version: "0.8.28",
        settings: {
          evmVersion: "cancun",
          optimizer: { enabled: true, runs: 200 },
        },
      },
    ],
  },
  typechain: {
    target: "ethers-v6",
  },
};

module.exports = config;
