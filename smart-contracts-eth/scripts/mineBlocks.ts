async function main() {
  const hre = require("hardhat");
  const blocksInADay = 7200;
  const blocksInAWeek = 50400;
  const blocksIn4Weeks = 201600;
  const blocksInAYear = 2620800;
  await hre.ethers.provider.send("evm_mine", [{ blocks: blocksInADay }]);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
