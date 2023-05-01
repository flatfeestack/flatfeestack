import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";
import { ethers } from "hardhat";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { daoContractDeployer, firstCouncilMember, secondCouncilMember } =
    await getNamedAccounts();
  const wallet = await deployments.get("Wallet");

  const membership = await deploy("Membership", {
    from: daoContractDeployer,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [firstCouncilMember, secondCouncilMember, wallet.address],
        },
      },
    },
  });

  const walletDeployed = await ethers.getContractAt("Wallet", wallet.address);

  const isKnownSender = await walletDeployed.isKnownSender(membership.address);
  if (!isKnownSender) {
    console.log(
      "Adding membership contract as known sender to wallet contract ..."
    );

    const transaction = await walletDeployed
      .connect(walletDeployed.provider.getSigner(daoContractDeployer))
      .addKnownSender(membership.address);

    console.log(`transaction hash: ${transaction.hash}`);
    await transaction.wait();
  }
};

export default func;
func.tags = ["Membership"];
