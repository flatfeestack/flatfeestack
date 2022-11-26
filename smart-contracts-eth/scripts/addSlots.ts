import { ethers } from "hardhat";

async function main() {
  const hre = require("hardhat");
  const { deployments } = hre;

  const daa = await deployments.get("DAA");
  const daaDeployed = await ethers.getContractAt("DAA", daa.address);

  const wallet = await deployments.get("Wallet");
  const walletDeployed = await ethers.getContractAt("Wallet", daa.address);

  const [chairman] = await ethers.getSigners();

  const blocksInAMonth = 201600;
  const latestBlock = (await hre.ethers.provider.getBlock("latest")).number;
  const slot1 = latestBlock + blocksInAMonth + 1;
  const slot2 = latestBlock + 2 * blocksInAMonth + 1;
  const slot3 = latestBlock + 3 * blocksInAMonth + 1;

  console.log("Creating voting slots ...");
  const [firstVotingSlot, secondVotingSlot, thirdVotingSlot] =
    await Promise.all([
      daaDeployed.connect(chairman).setVotingSlot(slot1),
      daaDeployed.connect(chairman).setVotingSlot(slot2),
      daaDeployed.connect(chairman).setVotingSlot(slot3),
    ]);

  await Promise.all([
    firstVotingSlot.wait(),
    secondVotingSlot.wait(),
    thirdVotingSlot.wait(),
  ]);

  const transferCalldata = [
    walletDeployed.interface.encodeFunctionData("increaseAllowance", [
      chairman.address,
      ethers.utils.parseEther("1.0"),
    ]),
  ];

  console.log("Creating proposal ...");
  await daaDeployed
    .connect(chairman)
    ["propose(address[],uint256[],bytes[],string)"](
      [wallet.address],
      [0],
      transferCalldata,
      "Give me, the president, some money!"
    );
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
