import { ethers } from "hardhat";

async function main() {
  const hre = require("hardhat");
  const { deployments } = hre;

  const membership = await deployments.get("Membership");
  const membershipDeployed = await ethers.getContractAt(
    "Membership",
    membership.address
  );

  const [firstCouncilMember, secondCouncilMember, member] =
    await ethers.getSigners();

  console.log("Requesting membership ...");
  await (await membershipDeployed.connect(member).requestMembership()).wait();

  console.log("Approving membership with first account ...");
  await (
    await membershipDeployed
      .connect(firstCouncilMember)
      .approveMembership(member.address)
  ).wait();

  console.log("Approving membership with second account ...");
  await (
    await membershipDeployed
      .connect(secondCouncilMember)
      .approveMembership(member.address)
  ).wait();

  console.log("Pay membership fees ...");
  const toBePaid = ethers.utils.parseUnits("3", 4); // exactly 30k wei

  await membershipDeployed.connect(firstCouncilMember).payMembershipFee({
    value: toBePaid,
  });
  await membershipDeployed.connect(secondCouncilMember).payMembershipFee({
    value: toBePaid,
  });
  await membershipDeployed.connect(member).payMembershipFee({
    value: toBePaid,
  });
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
