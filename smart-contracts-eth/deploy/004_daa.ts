import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction, Deployment } from "hardhat-deploy/types";
import { ethers } from "hardhat";
import type { Contract } from "ethers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { firstChairman } = await getNamedAccounts();

  const membership = await deployments.get("Membership");
  const timelock = await deployments.get("Timelock");

  await deploy("DAA", {
    from: firstChairman,
    log: true,
    proxy: {
      proxyContract: "OpenZeppelinTransparentProxy",
      execute: {
        init: {
          methodName: "initialize",
          args: [
            membership.address,
            timelock.address,
            "9f2984a2694119e92a301a51dcc800d1ee41bfaa29205b1a5474f2bde22ae3a6",
          ],
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
    revokeChairmanAsAdmin(timelockDeployed, firstChairman),
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
