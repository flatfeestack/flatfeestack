import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { representative, whitelisterOne, whitelisterTwo } =
    await getNamedAccounts();
  const wallet = await deployments.get("Wallet");

  await deploy("Membership", {
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
};

export default func;
func.tags = ["Membership"];
