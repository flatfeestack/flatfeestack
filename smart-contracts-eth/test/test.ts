import { expect } from "chai";
import { ethers } from "hardhat";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import { Contract, ContractTransactionResponse, TransactionResponse, EventLog, FormatType, FunctionFragment, AbiCoder } from "ethers";
import { Func } from "mocha";

let council1: HardhatEthersSigner, council2: HardhatEthersSigner, user1: HardhatEthersSigner, user2: HardhatEthersSigner;
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
  await contractUSDC.transfer(to.target as string, totalFundingUSDC);
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

async function eventLogs(tx: TransactionResponse, contract: Contract, eventName:string): Promise<EventLog[]> {
  const eventSig = contract.interface.getEvent(eventName)?.format();
  if(!eventSig) {
    throw "cannot find event " + eventName + " in contract";
  }
  const receipt = await tx.wait();
  const eventTopic = ethers.id(eventSig);
  return receipt?.logs.filter(log => log.topics[0] === eventTopic) as EventLog[];
}

async function eventLogsRaw(tx: TransactionResponse, eventSig:string): Promise<EventLog[]> {
  const receipt = await tx.wait();
  const eventTopic = ethers.id(eventSig);
  return receipt?.logs.filter(log => log.topics[0] === eventTopic) as EventLog[];
}

function connect(input:Contract, signer:HardhatEthersSigner):Contract {
  return input.connect(signer) as Contract;
}

beforeEach("Deploy All Contracts", async function() {
  [council1, council2, user1, user2] = await ethers.getSigners();
  const FlatFeeStackDAO = await ethers.getContractFactory("FlatFeeStackDAO");
  
  const deployTx = await FlatFeeStackDAO.getDeployTransaction(council1, council2);
  const tx = await council1.sendTransaction(deployTx);
  const receipt = await tx.wait();

  //we need to use eventLogsRaw, as we don't have the NFT contract yet
  const filteredLogs = await eventLogsRaw(tx, "FlatFeeStackNFTCreated(address,address)");
  const addressNFTStr = filteredLogs[0].topics[1];
  const addressNFT = ethers.getAddress(addressNFTStr.replace("000000000000000000000000", ""));
  const addressDAO = receipt?.contractAddress as string;

  contractUSDC = await ethers.deployContract("USDC");
  const addressUSDC = contractUSDC.target as string;

  contractNFT = await ethers.getContractAt("FlatFeeStackNFT", addressNFT);
  contractDAO = await ethers.getContractAt("FlatFeeStackDAO", addressDAO);
  
  contractPayoutEth = await ethers.deployContract("PayoutEth");
  contractPayoutUSDC = await ethers.deployContract("PayoutERC20", [addressUSDC]);

  //console.log("addressUSDC", addressUSDC);
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
    await contractPayoutEth.sendRecover(user2.address, recoveryAmountEth);

    expect(await contractPayoutEth.getBalance(user2.address)).to.equal(user2Balance + recoveryAmountEth);
    expect(await contractPayoutEth.getBalance(contractPayoutEth.target)).to.equal(totalFundingEth - recoveryAmountEth);
  });

  it("should allow owner to recover funds in USDC", async function () {
    const recoveryAmountUSDC = BigInt("1000000")
    const user2Balance = await contractPayoutUSDC.getBalance(user2.address);

    const totalFundingUSDC = await fundingUSDC(contractPayoutUSDC, 9);
    await contractPayoutUSDC.sendRecover(user2.address, recoveryAmountUSDC);

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
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

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
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

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
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, BigInt(1000000)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];

    const votingDelay = await contractDAO.votingDelay() as bigint;
    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    console.log("votingDelay in days:", Number(votingDelay) / (24 * 60 * 60));

    //too early TODO: testing
    //await connect(contractDAO, user1).castVote(proposalId, 1);

    await timeTravel(Number(votingDelay));

    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 0);
    await connect(contractDAO, council2).castVote(proposalId, 1);

    //too late TODO: testing
    //await timeTravel(Number(votingDelay) * 50);

    await timeTravel(Number(votingPeriod));

    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));
    
    const txQueue = await contractDAO.queue(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, BigInt(1000000)])],
      descriptionHash);

    const logProposalQueued = await eventLogs(txQueue, contractDAO, "ProposalQueued");
    const testProposalId1 = logProposalQueued[0].args[0];
    expect(testProposalId1).to.equal(proposalId);

    await timeTravel(Number(votingPeriod));    

    const txExecute = await contractDAO.execute(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, BigInt(1000000)])],
      descriptionHash);

    const logProposalExecuted = await eventLogs(txExecute, contractDAO, "ProposalExecuted");
    const testProposalId2 = logProposalExecuted[0].args[0];
    expect(testProposalId2).to.equal(proposalId);

    //check the USDC amount
    expect(await contractUSDC.balanceOf(user2.address)).to.equal(BigInt(1000000));

  });
});