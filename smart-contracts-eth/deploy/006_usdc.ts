import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  if (hre.network.name != "localhost") {
    return;
  }

  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { payoutUsdcDeployer } = await getNamedAccounts();

  await deploy("USDC", {
    from: payoutUsdcDeployer,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      viaAdminContract: {
        artifact: "MyProxyAdmin",
        name: "ProxyAdminPayoutUsdc",
      },
      execute: {
        init: {
          methodName: "initialize",
          args: [],
        },
      },
    },
  });
};

export default func;
func.tags = ["USDC"];
