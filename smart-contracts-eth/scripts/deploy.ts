import { ethers } from "hardhat";
import { EventLog } from "ethers";

async function main() {
    //deploy and fund contract
    const contractPayoutEth = await ethers.deployContract("PayoutEth");
    await contractPayoutEth.waitForDeployment();

    //top up payout contract with 1 eth
    const [council1, council2] = await ethers.getSigners();
    await council1.sendTransaction({
        to: contractPayoutEth.target,
        value: ethers.parseEther("1"), // Sends exactly 1.0 ether
    });

    const FlatFeeStackDAO = await ethers.getContractFactory("FlatFeeStackDAO");
    const deployTx = await FlatFeeStackDAO.getDeployTransaction(council1, council2);
    const tx = await council1.sendTransaction(deployTx);
    const receipt = await tx.wait();
    //we need to use eventLogsRaw, as we don't have the NFT contract yet
    const eventTopic = ethers.id("FlatFeeStackNFTCreated(address,address)");
    const filteredLogs = receipt?.logs.filter(log => log.topics[0] === eventTopic) as EventLog[];
    const addressNFTStr = filteredLogs[0].topics[1];
    const addressNFT = ethers.getAddress(addressNFTStr.replace("000000000000000000000000", ""));
    const addressDAO = receipt?.contractAddress as string;
    console.log("*************************");
    console.log("PayoutEth contract      : ", contractPayoutEth.target);
    console.log("FlatFeeStackDAO contract: ", addressDAO);
    console.log("FlatFeeStackNFT contract: ", addressNFT);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});