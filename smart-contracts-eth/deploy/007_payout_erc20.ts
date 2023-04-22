import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";
import { ethers } from "hardhat";
import { BigNumber } from "ethers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  let usdcTokenAddress: string;

  if (hre.network.name != "localhost") {
    usdcTokenAddress = (await getNamedAccounts()).usdcTokenAddress;
  } else {
    const usdcToken = await deployments.get("USDC");
    usdcTokenAddress = usdcToken.address;
  }

  const tokenDeployed = await ethers.getContractAt(
    "ERC20Upgradeable",
    usdcTokenAddress
  );
  const symbol = await tokenDeployed.symbol();

  const { firstCouncilMember } = await getNamedAccounts();

  const deploymentResult = await deploy("PayoutERC20", {
    from: firstCouncilMember,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [usdcTokenAddress, symbol],
        },
      },
    },
  });

  if (hre.network.name === "localhost") {
    const usdc = await deployments.get("USDC");
    const usdcContract = await ethers.getContractAt("USDC", usdc.address);
    const tokenQuantityForPayout = await usdcContract.balanceOf(
      deploymentResult.address
    );

    if (tokenQuantityForPayout.eq(BigNumber.from(0))) {
      console.log("Transfering some USDC tokens to payout contract");

      await (
        await usdcContract.transfer(deploymentResult.address, 10000)
      ).wait();
    }
  }
};

export default func;
func.tags = ["PayoutERC20"];
