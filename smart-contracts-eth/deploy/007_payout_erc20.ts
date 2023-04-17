import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";
import { ethers } from "hardhat";

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

  await deploy("PayoutERC20", {
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
};

export default func;
func.tags = ["PayoutERC20"];
