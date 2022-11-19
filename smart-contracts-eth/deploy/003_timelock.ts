import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { representative } = await getNamedAccounts();

  await deploy("Timelock", {
    from: representative,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [representative],
        },
      },
    },
  });

  const timelock = await deployments.get("Timelock");
  const wallet = await deployments.get("Wallet");
  const walletDeployed = await ethers.getContractAt("Wallet", wallet.address);

  const isTimelockWalletOwner =
    (await walletDeployed.owner()) === timelock.address;

  if (!isTimelockWalletOwner) {
    console.log("Assigning wallet ownership to timelock controller ...");
    await (await walletDeployed.transferOwnership(timelock.address)).wait();
  }
};

export default func;
func.tags = ["Wallet"];
