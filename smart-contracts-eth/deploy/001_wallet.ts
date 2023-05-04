import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { daoContractDeployer } = await getNamedAccounts();

  await deploy("Wallet", {
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
          args: [],
        },
      },
    },
  });
};

export default func;
func.tags = ["Wallet"];
