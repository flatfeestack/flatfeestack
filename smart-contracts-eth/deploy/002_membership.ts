import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";
import { ethers } from "hardhat";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { representative, whitelisterOne, whitelisterTwo } =
    await getNamedAccounts();
  const wallet = await deployments.get("Wallet");

  const membership = await deploy("Membership", {
    from: representative,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [
            representative,
            whitelisterOne,
            whitelisterTwo,
            wallet.address,
          ],
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
    await (await walletDeployed.addKnownSender(membership.address)).wait();
  }
};

export default func;
func.tags = ["Membership"];
