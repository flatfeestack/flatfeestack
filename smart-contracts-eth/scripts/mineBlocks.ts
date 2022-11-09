async function main() {
  const hre = require("hardhat");
  const blocksInADay = 6495;
  const blocksInAWeek = 45465;
  const blocksIn4Weeks = 181860;
  const blocksInAYear = 9456720;
  await hre.ethers.provider.send("evm_mine", [{ blocks: blocksInADay }]);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
