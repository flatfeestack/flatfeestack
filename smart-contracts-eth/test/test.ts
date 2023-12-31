import { expect } from "chai";
import { ethers } from "hardhat";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import { Contract, ContractTransactionResponse, EventLog } from "ethers";

let council1: HardhatEthersSigner, council2: HardhatEthersSigner, user1: HardhatEthersSigner, user2: HardhatEthersSigner;
let addressNFT:string, addressDAO:string, addressUSDC:string, addressPayoutEth:string, addressPayoutUSDC:string;
let contractNFT: Contract, contractDAO: Contract, contractUSDC: Contract, contractPayoutEth:Contract, contractPayoutUSDC:Contract;
 
async function sign(signer:HardhatEthersSigner, userId:string, payOut:bigint, contract:Contract) : Promise<string> {
  const message = ethers.solidityPackedKeccak256(
    ["address", "string", "uint256", "string", "uint256"], 
    [contract.target as string, "calculateWithdraw", userId, "#", payOut]);
    return await signer.signMessage(ethers.getBytes(message));
}

async function withdraw(contract:Contract, user:HardhatEthersSigner, payOut:bigint, signature:string, errorMsg:string|undefined) {
  const userBalance = await contract.getBalance(user.address) as bigint;
  if(!errorMsg) {
    // Withdraw funds
    await contract.withdraw(user.address, ethers.sha256(user.address), payOut, signature);
    // Check balance of user
    expect(await contract.getBalance(user.address)).to.equal(payOut + userBalance);
  } else {
    try {
      await contract.withdraw(user.address, ethers.sha256(user.address), payOut, signature);
      expect.fail("The transaction should have failed");
    } catch (error: any) {
      // Check if the error message is as expected
      expect(error.message).to.include(errorMsg);
    }
  }
}

async function fundingEth(to: Contract, eth:number): Promise<bigint> {
  const totalFundingEth = ethers.parseEther("" + eth);
  await council1.sendTransaction({to: to.target as string, value: totalFundingEth});
  return totalFundingEth;
}

async function fundingUSDC(to: Contract, usdc:number): Promise<bigint> {
  const totalFundingUSDC = BigInt("1000000") * BigInt(usdc);
  await contractUSDC.transfer(to.target as string, totalFundingUSDC, {signer: council1});
  return totalFundingUSDC;
}

async function signDAO(signer:HardhatEthersSigner, contract:Contract, to:HardhatEthersSigner, tokenId:bigint) : Promise<string> {
  const message = ethers.solidityPackedKeccak256(
    ["address", "string", "address", "string", "uint256"], 
    [contract.target as string, "safeMint", to.address, "#", tokenId]);
    return await signer.signMessage(ethers.getBytes(message));
}

async function timeTravel(seconds:number) {
  await ethers.provider.send("evm_increaseTime", [seconds]);
  await ethers.provider.send("evm_mine");
}

function connect(input:Contract, signer:HardhatEthersSigner):Contract {
  return input.connect(signer) as Contract;
}

beforeEach("Deploy All Contracts", async function() {
  [council1, council2, user1, user2] = await ethers.getSigners();
  const FlatFeeStackDAO = ethers.getContractFactory("FlatFeeStackDAO");
  
  const tx = (await FlatFeeStackDAO).getDeployTransaction(council1, council2);
  const sentTx = await council1.sendTransaction(await tx);
  const receipt = await sentTx.wait();

  const eventName = "FlatFeeStackNFTCreated(address,address)";
  const eventTopic = ethers.id(eventName);

  const filteredLogs = receipt?.logs.filter(log => log.topics[0] === eventTopic);
  const addressNFTStr = filteredLogs ? filteredLogs[0].topics[1] : "0x0";
  addressNFT = ethers.getAddress(addressNFTStr.replace("000000000000000000000000", ""));
  addressDAO = receipt?.contractAddress as string;

  contractUSDC = await ethers.deployContract("USDC", {signer: council1});
  addressUSDC = contractUSDC.target as string;

  contractNFT = await ethers.getContractAt("FlatFeeStackNFT", addressNFT);
  contractDAO = await ethers.getContractAt("FlatFeeStackDAO", addressDAO);
  
  contractPayoutEth = await ethers.deployContract("PayoutEth", {signer: council1});
  contractPayoutUSDC = await ethers.deployContract("PayoutERC20", [addressUSDC], {signer: council1});
})

describe("Withdraw Functionality", function () {
  it("should allow owner to withdraw correct amount in ETH", async function () {
    await fundingEth(contractPayoutEth, 10);

    const payOutEth = ethers.parseEther("1") as bigint;  
    const signature = await sign(council1, ethers.sha256(user1.address), payOutEth, contractPayoutEth);

    await withdraw(contractPayoutEth, user1, payOutEth, signature, undefined);
    await withdraw(contractPayoutEth, user1, payOutEth, signature, "Nothing to withdraw");
  });

  it("should allow owner to withdraw correct amount in USDC", async function () {
    await fundingUSDC(contractPayoutUSDC, 10);
    
    const payOutUSDC = BigInt("1000000"); 
    const signature = await sign(council1, ethers.sha256(user1.address), payOutUSDC, contractPayoutUSDC);

    await withdraw(contractPayoutUSDC, user1, payOutUSDC, signature, undefined);
    await withdraw(contractPayoutUSDC, user1, payOutUSDC, signature, "Nothing to withdraw");
  });

  it("should fail as no funds can be withdrawn in ETH and USDC", async function () {
    const payOut = BigInt("0");
    const signatureEth = await sign(council1, ethers.sha256(user1.address), payOut, contractPayoutEth);
    await withdraw(contractPayoutEth, user1, payOut, signatureEth, "Nothing to withdraw");

    const signatureUSDC = await sign(council1, ethers.sha256(user1.address), payOut, contractPayoutUSDC);
    await withdraw(contractPayoutEth, user1, payOut, signatureUSDC, "Nothing to withdraw");
  });

  it("should fail for invalid signature in Eth", async function () {
    await fundingEth(contractPayoutEth, 10);
    
    const invalidSignature = "0x" + "00".repeat(65);
    const payOutEth = ethers.parseEther("2") as bigint;

    await withdraw(contractPayoutEth, user1, payOutEth, invalidSignature, "ECDSAInvalidSignature");
  });

  it("should fail for invalid signature in Eth", async function () {
    await fundingUSDC(contractPayoutUSDC, 10);
    
    const invalidSignature = "0x" + "00".repeat(65);
    const payOutUSDC = BigInt("1000000"); 

    await withdraw(contractPayoutUSDC, user1, payOutUSDC, invalidSignature, "ECDSAInvalidSignature");
  });
});

describe("Balance and Recovery Functionality", function () {
  it("should return correct contract balance in Eth", async function () {
    const totalFundingEth = await fundingEth(contractPayoutEth, 9);
    expect(await contractPayoutEth.getBalance(contractPayoutEth.target)).to.equal(totalFundingEth);
  });

  it("should return correct contract balance in USDC", async function () {
    const totalFundingUSDC = await fundingUSDC(contractPayoutUSDC, 9);
    expect(await contractPayoutUSDC.getBalance(contractPayoutUSDC.target)).to.equal(totalFundingUSDC);
  });

  it("should allow owner to recover funds in Eth", async function () {
    const recoveryAmountEth = ethers.parseEther("1");
    const user2Balance = await contractPayoutEth.getBalance(user2.address);

    const totalFundingEth = await fundingEth(contractPayoutEth, 9);
    await contractPayoutEth.sendRecover(user2.address, recoveryAmountEth, {signer: council1});

    expect(await contractPayoutEth.getBalance(user2.address)).to.equal(user2Balance + recoveryAmountEth);
    expect(await contractPayoutEth.getBalance(contractPayoutEth.target)).to.equal(totalFundingEth - recoveryAmountEth);
  });

  it("should allow owner to recover funds in USDC", async function () {
    const recoveryAmountUSDC = BigInt("1000000")
    const user2Balance = await contractPayoutUSDC.getBalance(user2.address);

    const totalFundingUSDC = await fundingUSDC(contractPayoutUSDC, 9);
    await contractPayoutUSDC.sendRecover(user2.address, recoveryAmountUSDC, {signer: council1});

    expect(await contractPayoutUSDC.getBalance(user2.address)).to.equal(user2Balance + recoveryAmountUSDC);
    expect(await contractPayoutUSDC.getBalance(contractPayoutUSDC.target)).to.equal(totalFundingUSDC - recoveryAmountUSDC);
  });
});






describe("DAO Testing", function () {
  it("Check if council is setup", async function () {
    expect(await contractNFT.isCouncilIndex(council1.address, 0)).to.equal(true);
    expect(await contractNFT.isCouncilIndex(council2.address, 0)).to.equal(true);
    expect(await contractNFT.isCouncilIndex(user1.address, 0)).to.equal(false);

    expect(await contractNFT.balanceOf(council1.address)).to.equal(BigInt("1"));
    expect(await contractNFT.balanceOf(council2.address)).to.equal(BigInt("1"));
    expect(await contractNFT.balanceOf(user1.address)).to.equal(BigInt("0"));
  });

  it("Add user1 to DAO", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await contractNFT.safeMint(user1.address, 0, signature1, 0, signature2, {signer: user1, value: ethers.parseEther("1")})

    expect(await contractNFT.balanceOf(council1.address)).to.equal(BigInt("1"));
    expect(await contractNFT.balanceOf(council2.address)).to.equal(BigInt("1"));
    expect(await contractNFT.balanceOf(user1.address)).to.equal(BigInt("1"));
    //The 1eth goes to the DAO
    expect(await ethers.provider.getBalance(contractDAO.target)).to.equal(ethers.parseEther("1"));
    expect(await ethers.provider.getBalance(contractNFT.target)).to.equal(ethers.parseEther("0"));
    
    // make sure we have the votes for user1
    const timestampAfter = await contractDAO.clock();
    expect(await contractDAO.getVotes(council1.address, BigInt(0))).to.equal(BigInt(0));
    expect(await contractDAO.getVotes(council1.address, timestampAfter - BigInt(1))).to.equal(BigInt(1));
    //expect(await contractDAO.getVotes(user1.address, 1)).to.equal(BigInt(1));
  });

  it("Initiate Vote Failed", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await contractNFT.safeMint(user1.address, 0, signature1, 0, signature2, {signer: user1, value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user2
    //this will fail, as user2 cannot vote
    
    try {
      await connect(contractDAO, user2).propose(
        [contractUSDC.target],
        [BigInt(0)],
        [ethers.solidityPacked(
          ["string","address","uint256"],
          ["transfer",user2.address, BigInt(1000000)])],
        "Test Vote");
      expect.fail("The transaction should have failed");
    } catch (error: any) {
      // Check if the error message is as expected
      expect(error.message).to.include("GovernorInsufficientProposerVotes");
    }
  });

  it("Initiate Vote Success", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await contractNFT.safeMint(user1.address, 0, signature1, 0, signature2, {signer: user1, value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //contractDAO.castVote(,{signer: user1});
    //await timeTravel(2);
    //propose to send 1 USDC to user2 from user1
    const tx = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [ethers.solidityPacked(
        ["string","address","uint256"],
        ["transfer",user2.address, BigInt(1000000)])],
      "Test Vote") as ContractTransactionResponse;

    const receipt = await tx.wait();
    const eventName = "ProposalCreated(uint256,address,address[],uint256[],string[],bytes[],uint256,uint256,string)";
    const eventTopic = ethers.id(eventName);
    const filteredLogs = receipt?.logs.filter(log => log.topics[0] === eventTopic) as EventLog[];
    const proposalId = filteredLogs[0].args[0];

    const now = await contractDAO.clock();
    const votingDelay = await contractDAO.votingDelay() as bigint;
    console.log("votingDelay in days:", Number(votingDelay) / (24 * 60 * 60));

    //too early TODO: testing
    //await connect(contractDAO, user1).castVote(proposalId, 1);

    await timeTravel(Number(votingDelay));

    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 0);
    await connect(contractDAO, council2).castVote(proposalId, 0);

    //too late TODO: testing
    //await timeTravel(Number(votingDelay) * 50);
    

  });
});