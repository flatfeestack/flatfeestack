import { ethers } from "hardhat";

async function main() {
  const hre = require("hardhat");
  const { deployments } = hre;

  const membership = await deployments.get("Membership");
  const membershipDeployed = await ethers.getContractAt(
    "Membership",
    membership.address
  );

  const [chairman, whitelisterOne, whitelisterTwo, member] =
    await ethers.getSigners();

  console.log("Requesting membership ...");
  await (await membershipDeployed.connect(member).requestMembership()).wait();

  console.log("Approving membership with first account ...");
  await (
    await membershipDeployed
      .connect(whitelisterOne)
      .whitelistMember(member.address)
  ).wait();

  console.log("Approving membership with second account ...");
  await (
    await membershipDeployed
      .connect(whitelisterTwo)
      .whitelistMember(member.address)
  ).wait();
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
