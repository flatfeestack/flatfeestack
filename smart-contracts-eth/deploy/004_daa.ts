import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction, Deployment } from "hardhat-deploy/types";
import { ethers } from "hardhat";
import type { Contract } from "ethers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { representative } = await getNamedAccounts();

  const membership = await deployments.get("Membership");
  const timelock = await deployments.get("Timelock");

  await deploy("DAA", {
    from: representative,
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
    revokeRepresentativeAsAdmin(timelockDeployed, representative),
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

async function revokeRepresentativeAsAdmin(
  timelockDeployed: Contract,
  representative: String
) {
  const adminRole = await timelockDeployed.TIMELOCK_ADMIN_ROLE();
  const representativeIsAdmin = await timelockDeployed.hasRole(
    adminRole,
    representative
  );

  if (representativeIsAdmin) {
    console.log(
      "Revoking admin role for representative on timelock controller ..."
    );
    await (await timelockDeployed.revokeRole(adminRole, representative)).wait();
  }
}

export default func;
func.tags = ["DAA"];
