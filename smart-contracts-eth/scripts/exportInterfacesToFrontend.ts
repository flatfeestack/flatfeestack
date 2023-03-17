import {
  readFileSync,
  writeFileSync,
  createWriteStream,
  openSync,
  closeSync,
} from "fs";
import { resolve } from "path";
import { ethers } from "hardhat";

async function main() {
  const contractAddresses: String[] = [];
  const pathToFrontend = "../frontend";

  ["DAO", "Membership", "Wallet"].forEach(async (contractName: String) => {
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

    const contractProxyDeploymentFilePath = resolve(
      __dirname,
      `../deployments/localhost/${contractName}.json`
    );
    const contractProxyDeploymentFile = readFileSync(
      contractProxyDeploymentFilePath,
      "utf8"
    );
    const contractProxyDeployment = JSON.parse(contractProxyDeploymentFile);
    contractAddresses.push(
      `VITE_${contractName.toUpperCase()}_CONTRACT_ADDRESS=${
        contractProxyDeployment.address
      }`
    );
  });

  const dotEnvFilePath = resolve(__dirname, "..", `${pathToFrontend}/.env`);

  // empty .env file
  closeSync(openSync(dotEnvFilePath, "w"));
  const dotEnvFile = createWriteStream(dotEnvFilePath);
  contractAddresses.forEach(function (v) {
    dotEnvFile.write(`${v}\n`);
  });
  dotEnvFile.end();
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
