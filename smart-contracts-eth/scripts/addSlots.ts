import { ethers } from "hardhat";

async function main() {
  const hre = require("hardhat");
  const { deployments } = hre;

  const daa = await deployments.get("DAA");
  const daaDeployed = await ethers.getContractAt("DAA", daa.address);

  const wallet = await deployments.get("Wallet");
  const walletDeployed = await ethers.getContractAt("Wallet", daa.address);

  const [firstChairman] = await ethers.getSigners();

  const blocksInAMonth = 201600;
  const latestBlock = (await hre.ethers.provider.getBlock("latest")).number;
  const slot1 = latestBlock + blocksInAMonth + 1;
  const slot2 = latestBlock + 2 * blocksInAMonth + 1;
  const slot3 = latestBlock + 3 * blocksInAMonth + 1;

  console.log("Create first voting slot ...");
  const firstVotingSlot = await daaDeployed
    .connect(firstChairman)
    .setVotingSlot(slot1);
  await firstVotingSlot.wait();

  console.log("Create second voting slot ...");
  const secondVotingSlot = await daaDeployed
    .connect(firstChairman)
    .setVotingSlot(slot2);
  await secondVotingSlot.wait();

  console.log("Create third voting slot ...");
  const thirdVotingSlot = await daaDeployed
    .connect(firstChairman)
    .setVotingSlot(slot3);
  await thirdVotingSlot.wait();

  const transferCalldata = [
    walletDeployed.interface.encodeFunctionData("increaseAllowance", [
      firstChairman.address,
      ethers.utils.parseEther("1.0"),
    ]),
  ];

  console.log("Creating proposal ...");
  await daaDeployed
    .connect(firstChairman)
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
