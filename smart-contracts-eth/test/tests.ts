import { expect } from "chai";
import { ethers } from "hardhat";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import { Contract, ContractTransactionResponse, TransactionResponse, EventLog } from "ethers";

const ERR_PAYOUT_INVALID_SIG = "Invalid signature";

let council1: HardhatEthersSigner, council2: HardhatEthersSigner, user1: HardhatEthersSigner, user2: HardhatEthersSigner, user3: HardhatEthersSigner, user4: HardhatEthersSigner;
let contractNFT: any, contractDAO: any, contractUSDC: any, contractPayoutEth: any, contractPayoutUSDC: any;
 
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
  const totalFundingUSDC = dollar(1) * BigInt(usdc);
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

function dollar(nr:number) {
  return BigInt(nr) * BigInt(1000000);
}

beforeEach("Deploy All Contracts", async function() {
  [council1, council2, user1, user2, user3, user4] = await ethers.getSigners();
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
  await contractUSDC.waitForDeployment();
  const addressUSDC = await contractUSDC.getAddress();

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

    //test wrong user to withdraw
    await withdraw(contractPayoutEth, user2, payOutEth, signature, ERR_PAYOUT_INVALID_SIG);
    //correct user, good path
    await withdraw(contractPayoutEth, user1, payOutEth, signature, undefined);
    //correct user, but no funds left
    await withdraw(contractPayoutEth, user1, payOutEth, signature, "Nothing to withdraw");
  });

  it("not enough funds on contract in ETH", async function () {
    await fundingEth(contractPayoutEth, 10);

    const payOutEth = ethers.parseEther("11") as bigint;  
    const signature = await sign(council1, ethers.sha256(user1.address), payOutEth, contractPayoutEth);

    //correct user, good path, but not enough funds
    await withdraw(contractPayoutEth, user1, payOutEth, signature, "ETH Insufficient Balance");
  });

  it("not enough funds on contract in USDC", async function () {
    await fundingUSDC(contractPayoutUSDC, 10);

    const payOutUSDC = dollar(11);
    const signature = await sign(council1, ethers.sha256(user1.address), payOutUSDC, contractPayoutUSDC);

    //correct user, good path, but not enough funds
    await withdraw(contractPayoutUSDC, user1, payOutUSDC, signature, "ERC20InsufficientBalance");
  });

  it("should allow owner to withdraw correct amount in USDC", async function () {
    await fundingUSDC(contractPayoutUSDC, 10);
    
    const payOutUSDC = dollar(1);
    const signature = await sign(council1, ethers.sha256(user1.address), payOutUSDC, contractPayoutUSDC);

    //test wrong user to withdraw
    await withdraw(contractPayoutUSDC, user2, payOutUSDC, signature, ERR_PAYOUT_INVALID_SIG);
    //correct user, good path
    await withdraw(contractPayoutUSDC, user1, payOutUSDC, signature, undefined);
    //correct user, but no funds left
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

    await withdraw(contractPayoutEth, user1, payOutEth, invalidSignature, ERR_PAYOUT_INVALID_SIG);
  });

  it("should fail for invalid signature in Eth", async function () {
    await fundingUSDC(contractPayoutUSDC, 10);
    
    const invalidSignature = "0x" + "00".repeat(65);
    const payOutUSDC = dollar(1);

    await withdraw(contractPayoutUSDC, user1, payOutUSDC, invalidSignature, ERR_PAYOUT_INVALID_SIG);
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
    const recoveryAmountUSDC = dollar(1);
    const user2Balance = await contractPayoutUSDC.getBalance(user2.address);

    const totalFundingUSDC = await fundingUSDC(contractPayoutUSDC, 9);
    await contractPayoutUSDC.sendRecoverToken(contractUSDC.target, user2.address, recoveryAmountUSDC);

    expect(await contractPayoutUSDC.getBalance(user2.address)).to.equal(user2Balance + recoveryAmountUSDC);
    expect(await contractPayoutUSDC.getBalance(contractPayoutUSDC.target)).to.equal(totalFundingUSDC - recoveryAmountUSDC);
  });

  it("fail to recover funds in USDC", async function () {
    const recoveryAmountUSDC = dollar(10);
    await fundingUSDC(contractPayoutUSDC, 9);
    try {
      await contractPayoutUSDC.sendRecoverToken(contractUSDC.target, user2.address, recoveryAmountUSDC);
    } catch (error:any) {
      expect(error.message).to.include("ERC20InsufficientBalance");
    }
  });

  it("fail to recover funds in Eth and USDC as not owner", async function () {
    const recoveryAmountUSDC = dollar(1);
    const recoveryAmountEth = ethers.parseEther("1");
    await fundingUSDC(contractPayoutUSDC, 9);
    await fundingEth(contractPayoutEth, 9);

    try {
      await connect(contractPayoutUSDC, user1).sendRecoverToken(contractUSDC.target, user2.address, recoveryAmountUSDC);
    } catch (error:any) {
      expect(error.message).to.include("OwnableUnauthorizedAccount");
    }

    try {
      await connect(contractPayoutEth, user1).sendRecover(user2.address, recoveryAmountEth);
    } catch (error:any) {
      expect(error.message).to.include("OwnableUnauthorizedAccount");
    }

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

  it("Wrong signature1", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("2"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    try {
      await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")});
    } catch (error:any) {
      expect(error.message).to.include("Signature err");
    }
  });

  it("Wrong signature2", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("2"));
    try {
      await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")});
    } catch (error:any) {
      expect(error.message).to.include("Signature err");
    }
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
        [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
        "Test Vote");
      expect.fail("The transaction should have failed");
    } catch (error: any) {
      // Check if the error message is as expected
      expect(error.message).to.include("GovernorInsufficientProposerVotes");
    }
  });

  it("Initiate Vote - too early", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;

    //too early
    await timeTravel(Number(votingDelay) - 1);
    try {
      await connect(contractDAO, user1).castVote(proposalId, 1);
    } catch(error:any) {
      expect(error.message).to.include("GovernorUnexpectedProposalState");
    }
  });

  it("Initiate Vote - too late", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];

    //too late
    const votingDelay = await contractDAO.votingDelay() as bigint;
    await timeTravel(Number(votingDelay) * 50);
    try {
      await connect(contractDAO, user1).castVote(proposalId, 1);
    } catch(error:any) {
      expect(error.message).to.include("GovernorUnexpectedProposalState");
    }
  });

  it("Initiate Vote - queue too early", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;
    await timeTravel(Number(votingDelay));

    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 0);
    await connect(contractDAO, council2).castVote(proposalId, 1);
    
    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));
    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    //1s too early
    await timeTravel(Number(votingPeriod) - 1);
    
    try {
      await contractDAO.queue(
        [contractUSDC.target],
        [BigInt(0)],
        [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
        descriptionHash);
    } catch(error:any) {
      expect(error.message).to.include("GovernorUnexpectedProposalState");
    }
  });

  it("Initiate Vote - execute too early", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;
    await timeTravel(Number(votingDelay));

    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 0);
    await connect(contractDAO, council2).castVote(proposalId, 1);
    
    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));
    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    await timeTravel(Number(votingPeriod));
    
    await contractDAO.queue(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      descriptionHash);

    //1s too early
    await timeTravel(Number(votingPeriod) - 1);

    try {
      await contractDAO.execute(
        [contractUSDC.target],
        [BigInt(0)],
        [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
        descriptionHash);
    } catch(error:any) {
      expect(error.message).to.include("GovernorUnexpectedProposalState");
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
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;
    
    await timeTravel(Number(votingDelay));

    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 0);
    await connect(contractDAO, council2).castVote(proposalId, 1);

    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    await timeTravel(Number(votingPeriod));

    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));
    
    const txQueue = await contractDAO.queue(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      descriptionHash);

    const logProposalQueued = await eventLogs(txQueue, contractDAO, "ProposalQueued");
    const testProposalId1 = logProposalQueued[0].args[0];
    expect(testProposalId1).to.equal(proposalId);

    await timeTravel(Number(votingPeriod));    

    const txExecute = await contractDAO.execute(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      descriptionHash);

    const logProposalExecuted = await eventLogs(txExecute, contractDAO, "ProposalExecuted");
    const testProposalId2 = logProposalExecuted[0].args[0];
    expect(testProposalId2).to.equal(proposalId);

    //check the USDC amount
    expect(await contractUSDC.balanceOf(user2.address)).to.equal(dollar(1));
  });

  it("Initiate Vote Reject", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;
    
    await timeTravel(Number(votingDelay));

    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 0);
    await connect(contractDAO, council2).castVote(proposalId, 0);

    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    await timeTravel(Number(votingPeriod));

    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));

    try {
      const txQueue = await contractDAO.queue(
        [contractUSDC.target],
        [BigInt(0)],
        [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
        descriptionHash);
    } catch (error:any) {
      expect(error.message).to.include("GovernorUnexpectedProposalState");
    }
  });

  it("Initiate Vote Success - no queue", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));
    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;
    
    await timeTravel(Number(votingDelay));

    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 0);
    await connect(contractDAO, council2).castVote(proposalId, 1);

    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    await timeTravel(Number(votingPeriod));
    await timeTravel(Number(votingPeriod));
    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));

    const txExecute = await contractDAO.execute(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      descriptionHash);

    const logProposalExecuted = await eventLogs(txExecute, contractDAO, "ProposalExecuted");
    const testProposalId2 = logProposalExecuted[0].args[0];
    expect(testProposalId2).to.equal(proposalId);

    //check the USDC amount
    expect(await contractUSDC.balanceOf(user2.address)).to.equal(dollar(1));
  });

  it("Initiate Vote Failed - no quorum", async function () {
    const signature1 = signDAO(council1, contractNFT, user1, BigInt("3"));
    const signature2 = signDAO(council2, contractNFT, user1, BigInt("3"));

    const signature3 = signDAO(council1, contractNFT, user2, BigInt("4"));
    const signature4 = signDAO(council2, contractNFT, user2, BigInt("4"));

    const signature5 = signDAO(council1, contractNFT, user3, BigInt("5"));
    const signature6 = signDAO(council2, contractNFT, user3, BigInt("5"));

    const signature7 = signDAO(council1, contractNFT, user4, BigInt("6"));
    const signature8 = signDAO(council2, contractNFT, user4, BigInt("6"));

    await connect(contractNFT, user1).safeMint(user1.address, 0, signature1, 0, signature2, {value: ethers.parseEther("1")})
    await connect(contractNFT, user2).safeMint(user2.address, 0, signature3, 0, signature4, {value: ethers.parseEther("1")})
    await connect(contractNFT, user3).safeMint(user3.address, 0, signature5, 0, signature6, {value: ethers.parseEther("1")})
    await connect(contractNFT, user4).safeMint(user4.address, 0, signature7, 0, signature8, {value: ethers.parseEther("1")})

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, user1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;
    
    await timeTravel(Number(votingDelay));
    await connect(contractDAO, user1).castVote(proposalId, 1);

    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    await timeTravel(Number(votingPeriod));
    await timeTravel(Number(votingPeriod));
    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));

    try {
      const txExecute = await contractDAO.execute(
        [contractUSDC.target],
        [BigInt(0)],
        [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
        descriptionHash);
    } catch (error:any) {
      expect(error.message).to.include("GovernorUnexpectedProposalState");
    }
  });

  it("Initiate Vote Success - only one vote with 2 councils", async function () {

    try {
    await connect(contractNFT, council2).burn(2);
    } catch(error:any) {
      expect(error.message).to.include("Is council");
    }

    //send 1 USDC to the contract
    await fundingUSDC(contractDAO, 1);

    //propose to send 1 USDC to user2 from user1
    const txPropose = await (connect(contractDAO, council1)).propose(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      "Test Vote") as ContractTransactionResponse;
    
    const logProposalCreated = await eventLogs(txPropose, contractDAO, "ProposalCreated");
    const proposalId = logProposalCreated[0].args[0];
    const votingDelay = await contractDAO.votingDelay() as bigint;
    
    await timeTravel(Number(votingDelay));

    //user1 can vote, but weigth will be 0
    await connect(contractDAO, user1).castVote(proposalId, 1);
    await connect(contractDAO, council1).castVote(proposalId, 1);

    const now = await contractDAO.clock();

    const votingPeriod = await contractDAO.votingPeriod() as bigint;
    await timeTravel(Number(votingPeriod));
    await timeTravel(Number(votingPeriod));
    const descriptionHash =  ethers.keccak256(ethers.toUtf8Bytes("Test Vote"));

    const q = await contractDAO.quorum(now - BigInt(1));
    
    const txExecute = await contractDAO.execute(
      [contractUSDC.target],
      [BigInt(0)],
      [contractUSDC.interface.encodeFunctionData("transfer", [user2.address, dollar(1)])],
      descriptionHash);
  });
});