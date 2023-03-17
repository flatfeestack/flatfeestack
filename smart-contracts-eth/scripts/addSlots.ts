import { ethers } from "hardhat";

async function main() {
  const hre = require("hardhat");
  const { deployments } = hre;

  const dao = await deployments.get("DAO");
  const daoDeployed = await ethers.getContractAt("DAO", dao.address);

  const wallet = await deployments.get("Wallet");
  const walletDeployed = await ethers.getContractAt("Wallet", dao.address);

  const [firstCouncilMember] = await ethers.getSigners();

  const blocksInAMonth = 201600;
  const latestBlock = (await hre.ethers.provider.getBlock("latest")).number;
  const slot1 = latestBlock + blocksInAMonth + 1;
  const slot2 = latestBlock + 2 * blocksInAMonth + 1;
  const slot3 = latestBlock + 3 * blocksInAMonth + 1;

  console.log("Create first voting slot ...");
  const firstVotingSlot = await daoDeployed
    .connect(firstCouncilMember)
    .setVotingSlot(slot1);
  await firstVotingSlot.wait();

  console.log("Create second voting slot ...");
  const secondVotingSlot = await daoDeployed
    .connect(firstCouncilMember)
    .setVotingSlot(slot2);
  await secondVotingSlot.wait();

  console.log("Create third voting slot ...");
  const thirdVotingSlot = await daoDeployed
    .connect(firstCouncilMember)
    .setVotingSlot(slot3);
  await thirdVotingSlot.wait();

  const transferCalldata = [
    walletDeployed.interface.encodeFunctionData("increaseAllowance", [
      firstCouncilMember.address,
      ethers.utils.parseEther("1.0"),
    ]),
  ];

  console.log("Creating proposal ...");
  await daoDeployed
    .connect(firstCouncilMember)
    ["propose(address[],uint256[],bytes[],string)"](
      [wallet.address],
      [0],
      transferCalldata,
      "Give me, the president, some money!"
    );
}

main().catch((error) => {
  console.log(error.data);
  console.dir(error);
  process.exitCode = 1;
});
