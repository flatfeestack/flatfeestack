import { readFileSync, writeFileSync } from "fs";
import { resolve } from "path";
import { ethers } from "hardhat";

async function main() {
  const pathToFrontend = "../frontend";

  ["DAO", "Membership", "PayoutEth", "PayoutERC20", "Wallet"].forEach(
    async (contractName: String) => {
      const contractAbi = resolve(
        __dirname,
        `../artifacts/contracts/${contractName}.sol/${contractName}.json`
      );

      const file = readFileSync(contractAbi, "utf8");
      const json = JSON.parse(file);
      const abi = json.abi;

      const iface = new ethers.utils.Interface(abi);
      const resultFile = resolve(
        __dirname,
        "..",
        `${pathToFrontend}/src/contracts/${contractName}.ts`
      );

      writeFileSync(
        resultFile,
        `export const ${contractName}ABI = ${JSON.stringify(
          iface.format(ethers.utils.FormatTypes.full)
        )}`
      );
    }
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
