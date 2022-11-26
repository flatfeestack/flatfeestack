import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction, Deployment } from "hardhat-deploy/types";
import { ethers } from "hardhat";
import type { Contract } from "ethers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { chairman } = await getNamedAccounts();

  const membership = await deployments.get("Membership");
  const timelock = await deployments.get("Timelock");

  await deploy("DAA", {
    from: chairman,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [membership.address, timelock.address],
        },
      },
    },
    gasPrice: "3000000000",
  });

  const timelockDeployed = await ethers.getContractAt(
    "Timelock",
    timelock.address
  );
  const daa = await deployments.get("DAA");

  await Promise.all([
    addDaaAsProposer(timelockDeployed, daa),
    revokeChairmanAsAdmin(timelockDeployed, chairman),
  ]);
};

async function addDaaAsProposer(timelockDeployed: Contract, daa: Deployment) {
  const proposerRole = await timelockDeployed.PROPOSER_ROLE();
  const daaIsProposer = await timelockDeployed.hasRole(
    proposerRole,
    daa.address
  );

  if (!daaIsProposer) {
    console.log("Granting proposer role to DAA on timelock controller ...");
    await (await timelockDeployed.grantRole(proposerRole, daa.address)).wait();
  }
}

async function revokeChairmanAsAdmin(
  timelockDeployed: Contract,
  chairman: String
) {
  const adminRole = await timelockDeployed.TIMELOCK_ADMIN_ROLE();
  const chairmanIsAdmin = await timelockDeployed.hasRole(adminRole, chairman);

  if (chairmanIsAdmin) {
    console.log("Revoking admin role for chairman on timelock controller ...");
    await (await timelockDeployed.revokeRole(adminRole, chairman)).wait();
  }
}

export default func;
func.tags = ["DAA"];
