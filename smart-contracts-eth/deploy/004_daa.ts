import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction, Deployment } from "hardhat-deploy/types";
import { ethers } from "hardhat";
import type { Contract } from "ethers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { firstCouncilMember } = await getNamedAccounts();

  const membership = await deployments.get("Membership");
  const timelock = await deployments.get("Timelock");

  await deploy("DAO", {
    from: firstCouncilMember,
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
  const dao = await deployments.get("DAO");

  await Promise.all([
    addDaoAsProposer(timelockDeployed, dao),
    revokeCouncilMemberAsAdmin(timelockDeployed, firstCouncilMember),
  ]);
};

async function addDaoAsProposer(timelockDeployed: Contract, dao: Deployment) {
  const proposerRole = await timelockDeployed.PROPOSER_ROLE();
  const daoIsProposer = await timelockDeployed.hasRole(
    proposerRole,
    dao.address
  );

  if (!daoIsProposer) {
    console.log("Granting proposer role to DAO on timelock controller ...");
    await (await timelockDeployed.grantRole(proposerRole, dao.address)).wait();
  }
}

async function revokeCouncilMemberAsAdmin(
  timelockDeployed: Contract,
  councilMember: String
) {
  const adminRole = await timelockDeployed.TIMELOCK_ADMIN_ROLE();
  const councilMemberIsAdmin = await timelockDeployed.hasRole(
    adminRole,
    councilMember
  );

  if (councilMemberIsAdmin) {
    console.log(
      "Revoking admin role for council member on timelock controller ..."
    );
    await (await timelockDeployed.revokeRole(adminRole, councilMember)).wait();
  }
}

export default func;
func.tags = ["DAO"];
