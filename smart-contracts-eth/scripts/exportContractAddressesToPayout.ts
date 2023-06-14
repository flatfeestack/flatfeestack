import { readFileSync, writeFileSync } from "fs";
import { resolve } from "path";
import dotenv from "dotenv";
import { stringify } from "envfile";
import { HardhatRuntimeEnvironment } from "hardhat/types";

async function main() {
  const hre: HardhatRuntimeEnvironment = require("hardhat");
  const pathToPayout = process.env.PAYOUT_PATH ?? "../../payout";

  const dotEnvFilePath = resolve(__dirname, `${pathToPayout}/.env`);
  let parseDotEnvFile: any = {};
  try {
    parseDotEnvFile = dotenv.parse(
      readFileSync(dotEnvFilePath, { encoding: "utf8" })
    );
  } catch (err: any) {
    if (err.code !== "ENOENT") {
      console.error(
        `There was a problem processing the .env file (${dotEnvFilePath})`,
        err
      );
    }
  }

  ["DAO", "Membership", "Wallet"].forEach(async (contractName: String) => {
    parseDotEnvFile[`DAO_${contractName.toUpperCase()}_CONTRACT`] =
      getProxyContractAddress(hre, contractName);
  });

  // export the payout contracts
  parseDotEnvFile[`ETH_CONTRACT`] = getProxyContractAddress(hre, "PayoutEth");
  parseDotEnvFile[`USDC_CONTRACT`] = getProxyContractAddress(hre, "PayoutUsdc");

  writeFileSync(dotEnvFilePath, stringify(parseDotEnvFile));
}

function getProxyContractAddress(
  hre: HardhatRuntimeEnvironment,
  contractName: String
) {
  const contractProxyDeploymentFilePath = resolve(
    __dirname,
    `../deployments/${hre.network.name}/${contractName}.json`
  );
  const contractProxyDeploymentFile = readFileSync(
    contractProxyDeploymentFilePath,
    "utf8"
  );
  return JSON.parse(contractProxyDeploymentFile).address;
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
