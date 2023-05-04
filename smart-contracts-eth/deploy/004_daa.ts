import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction, Deployment } from "hardhat-deploy/types";
import { ethers } from "hardhat";
import type { Contract } from "ethers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;

  const { daoContractDeployer, firstCouncilMember } = await getNamedAccounts();

  const membership = await deployments.get("Membership");
  const timelock = await deployments.get("Timelock");

  await deploy("DAO", {
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
          args: [
            membership.address,
            timelock.address,
            "9f2984a2694119e92a301a51dcc800d1ee41bfaa29205b1a5474f2bde22ae3a6",
            "https://flatfeestack.github.io/bylaws",
          ],
        },
      },
    },
  });

  const timelockDeployed = await ethers.getContractAt(
    "Timelock",
    timelock.address
  );
  const dao = await deployments.get("DAO");

  await addDaoAsProposer(timelockDeployed, dao, daoContractDeployer);
  await revokeCouncilMemberAsAdmin(timelockDeployed, firstCouncilMember);
};

async function addDaoAsProposer(
  timelockDeployed: Contract,
  dao: Deployment,
  daoContractDeployer: string
) {
  const proposerRole = await timelockDeployed.PROPOSER_ROLE();
  const daoIsProposer = await timelockDeployed.hasRole(
    proposerRole,
    dao.address
  );

  if (!daoIsProposer) {
    console.log("Granting proposer role to DAO on timelock controller ...");
    const transaction = await timelockDeployed
      .connect(timelockDeployed.provider.getSigner(daoContractDeployer))
      .grantRole(proposerRole, dao.address);

    console.log(`Transaction hash ${transaction.hash}`);

    await transaction.wait();
  }
}

async function revokeCouncilMemberAsAdmin(
  timelockDeployed: Contract,
  contractDeployer: String
) {
  const adminRole = await timelockDeployed.TIMELOCK_ADMIN_ROLE();
  const deployerIsAdmin = await timelockDeployed.hasRole(
    adminRole,
    contractDeployer
  );

  if (deployerIsAdmin) {
    console.log("Revoking admin role for deployer on timelock controller ...");
    const transaction = await timelockDeployed
      .connect(timelockDeployed.provider.getSigner(contractDeployer))
      .revokeRole(adminRole, contractDeployer);

    console.log(`Transaction hash ${transaction.hash}`);

    await transaction.wait();
  }
}

export default func;
func.tags = ["DAO"];
