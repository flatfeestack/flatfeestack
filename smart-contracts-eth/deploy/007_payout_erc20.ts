import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  let tokenAddress: string;

  if (hre.network.name != "localhost") {
    tokenAddress = (await getNamedAccounts()).payoutERC20Contract;
  } else {
    const ffsToken = await deployments.get("FlatFeeStackToken");
    tokenAddress = ffsToken.address;
  }

  const { firstCouncilMember } = await getNamedAccounts();

  await deploy("PayoutERC20", {
    from: firstCouncilMember,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [tokenAddress],
        },
      },
    },
  });
};

export default func;
func.tags = ["PayoutERC20"];
