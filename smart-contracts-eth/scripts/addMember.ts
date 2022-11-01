import { ethers } from "hardhat";

async function main() {
  const hre = require("hardhat");
  const { deployments } = hre;

  const membership = await deployments.get("Membership");
  const membershipDeployed = await ethers.getContractAt(
    "Membership",
    membership.address
  );

  const [representative, whitelisterOne, whitelisterTwo, member] =
    await ethers.getSigners();

  await membershipDeployed.connect(member).requestMembership();
  await membershipDeployed
    .connect(whitelisterOne)
    .whitelistMember(member.address);
  await membershipDeployed
    .connect(whitelisterTwo)
    .whitelistMember(member.address);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
