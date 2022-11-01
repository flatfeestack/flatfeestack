import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { representative } = await getNamedAccounts();
  const membership = await deployments.get("Membership");

  await deploy("DAA", {
    from: representative,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [membership.address],
        },
      },
    },
  });
};

export default func;
func.tags = ["DAA"];
