import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import type { Contract } from "ethers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  if (hre.network.name == "main") {
    console.log("Timelock wont be deployed for now on main net.");
    return;
  }

  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { daoContractDeployer } = await getNamedAccounts();

  await deploy("Timelock", {
    from: daoContractDeployer,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      viaAdminContract: {
        artifact: "MyProxyAdmin",
        name: "DefaultProxyAdmin",
      },
      execute: {
        init: {
          methodName: "initialize",
          args: [daoContractDeployer],
        },
      },
    },
  });

  const timelock = await deployments.get("Timelock");

  const wallet = await deployments.get("Wallet");
  const walletDeployed = await ethers.getContractAt("Wallet", wallet.address);

  const membership = await deployments.get("Membership");
  const membershipDeployed = await ethers.getContractAt(
    "Membership",
    membership.address
  );

  await Promise.all([
    assignContractOwnershipToTimeLock(
      membershipDeployed,
      timelock.address,
      daoContractDeployer
    ),
    assignContractOwnershipToTimeLock(
      walletDeployed,
      timelock.address,
      daoContractDeployer
    ),
  ]);
};

async function assignContractOwnershipToTimeLock(
  contract: Contract,
  timelockAddress: string,
  daoDeployer: string
) {
  const isTimelockContractOwner = (await contract.owner()) === timelockAddress;

  if (!isTimelockContractOwner) {
    console.log(
      `Assigning ${contract.address} ownership to timelock controller ...`
    );
    const transaction = await contract
      .connect(contract.provider.getSigner(daoDeployer))
      .transferOwnership(timelockAddress);

    console.log(`Transaction hash ${transaction.hash}`);

    await transaction.wait();
  }
}

export default func;
func.tags = ["Wallet"];
