/*import { expect } from "chai";
import { ethers } from "hardhat";
import { IEntryPoint } from "../typechain-types";

describe("Gasless DAO via Account Abstraction", function () {
  let council1: any, council2: any, user1: any;
  let dao: any, nft: any, paymaster: any, entryPoint: any;
  let usdc: any, smartAccount: any;

  async function signDAO(signer: any, contract: any, to: any, tokenId: bigint) {
    const payloadHash = ethers.solidityPackedKeccak256(
      ["address", "string", "address", "string", "uint256"],
      [await contract.getAddress(), "safeMint", to, "#", tokenId]
    );

    const signature = await signer.signMessage(ethers.getBytes(payloadHash));
    return signature;
  }

  beforeEach(async function () {
    [council1, council2, user1] = await ethers.getSigners();

    // --- deploy DAO and NFT ---
    const DAOFactory = await ethers.getContractFactory("FlatFeeStackDAO");
    const daoDeployTx = await DAOFactory.getDeployTransaction(council1.address, council2.address);
    const tx = await council1.sendTransaction(daoDeployTx);
    const receipt = await tx.wait();
    const daoAddr = receipt?.contractAddress as string;

    // find emitted NFT address
    const eventTopic = ethers.id("FlatFeeStackNFTCreated(address,address)");
    const logs = receipt.logs.filter((l: { topics: string[]; }) => l.topics[0] === eventTopic);
    const nftAddr = ethers.getAddress("0x" + logs[0].topics[1].slice(26));

    dao = await ethers.getContractAt("FlatFeeStackDAO", daoAddr);
    nft = await ethers.getContractAt("FlatFeeStackNFT", nftAddr);

    // mock ERC-20
    const Token = await ethers.getContractFactory("USDC");
    usdc = await Token.deploy();
    await usdc.waitForDeployment();

    // --- deploy EntryPoint ---
    const EntryPointFactory = await ethers.getContractFactory(
      "@account-abstraction/contracts/core/EntryPoint.sol:EntryPoint"
    );
    entryPoint = (await EntryPointFactory.deploy()) as IEntryPoint;
    await entryPoint.waitForDeployment();

    // --- deploy Paymaster ---
    const PaymasterFactory = await ethers.getContractFactory("FlatFeeStackDAOPaymaster");
    paymaster = await PaymasterFactory.deploy(
      await entryPoint.getAddress(),
      council1.address,
      council2.address
    );
    await paymaster.waitForDeployment();

    // fund Paymaster in EntryPoint
    await entryPoint.depositTo(await paymaster.getAddress(), { value: ethers.parseEther("2") });

    const FirstAccount = await ethers.getContractFactory("FirstAccount");
    smartAccount = await FirstAccount.deploy(user1.address, entryPoint.getAddress());
    await smartAccount.waitForDeployment();

    //const smartAddr = await smartAccount.getAddress();
  });

  it("should mint NFT and renew membership gaslessly", async function () {
    //const paymasterNFTAddr = await paymaster.token();
    //const nft = await ethers.getContractAt("FlatFeeStackNFT", paymasterNFTAddr);
    const smartAddr = await smartAccount.getAddress();
    const sig1 = await signDAO(council1, nft, smartAddr, 3n);
    const sig2 = await signDAO(council2, nft, smartAddr, 3n);

    await nft.connect(user1).safeMint(smartAddr, 0, sig1, 0, sig2, {
      value: ethers.parseEther("1"),
    });

    const tokenId = await nft.tokenOfOwnerByIndex(smartAddr, 0);

    // prepare gasless UserOperation for payMembership
    const callData = ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes"],
      [
        nft.target,
        ethers.parseEther("1"),
        nft.interface.encodeFunctionData("payMembership", [tokenId])
      ]
    );

    // Fund smart Account
    await user1.sendTransaction({
      to: smartAddr,
      value: ethers.parseEther("1"),
    });

    const paymasterAndData = ethers.solidityPacked(
      ["address", "uint128", "uint128"],
      [await paymaster.getAddress(), 100_000, 100_000]
    );

    const accountGasLimits = ethers.solidityPacked(
      ["uint128", "uint128"],
      [500_000, 3_000_000]
    );

    const gasFees = ethers.solidityPacked(
      ["uint128", "uint128"],
      [
        ethers.parseUnits("1", "gwei"),
        ethers.parseUnits("5", "gwei")
      ]
    );

    const userOp = {
      sender: smartAddr,
      nonce: await entryPoint.getNonce(smartAddr, 0),

      initCode: "0x",
      callData,

      accountGasLimits,
      preVerificationGas: 100_000n,
      gasFees,

      paymasterAndData,
      signature: "0x"
    };

    /*const userOpHash = await entryPoint.getUserOpHash(userOp);
    const message = ethers.getBytes(userOpHash);

    const sigHex = await user1.signMessage(ethers.getBytes(message));

    userOp.signature = sigHex;*/
/*
    const userOpHash = await entryPoint.getUserOpHash(userOp);
    userOp.signature = await user1.signMessage(ethers.getBytes(userOpHash));
    console.log("NFT used in OP:", nft.target);
    console.log("Paymaster token():", await paymaster.token());

    await entryPoint.handleOps([userOp], council1.address);

    const renewed = await nft.membershipPayed(tokenId);
    expect(renewed).to.be.gt(0n);
  });

  /*it("should cast DAO vote gaslessly", async function () {
    // council proposes a vote
    const txPropose = await dao.connect(council1).propose(
      [await usdc.getAddress()],
      [0],
      [usdc.interface.encodeFunctionData("transfer", [user1.address, 1000n])],
      "Gasless Vote"
    );
    const receipt = await txPropose.wait();
    const topic = dao.interface.getEvent("ProposalCreated").topicHash;
    const log = receipt.logs.find((l: any) => l.topics[0] === topic);
    const proposalId = dao.interface.parseLog(log).args[0];

    const votingDelay = await dao.votingDelay();
    await timeTravel(Number(votingDelay));

    // prepare gasless vote
    const voteData = dao.interface.encodeFunctionData("castVote", [proposalId, 1]);

    const accountGasLimits = ethers.solidityPacked(
      ["uint128", "uint128"],
      [500_000, 3_000_000]
    );
    const gasFees = ethers.solidityPacked(
      ["uint128", "uint128"],
      [
        ethers.parseUnits("1", "gwei"), // priority
        ethers.parseUnits("5", "gwei"), // max
      ]
    );

    const userOp = {
      sender: await smartAccount.getAddress(),
      nonce: await entryPoint.getNonce(await smartAccount.getAddress(), 0),
      initCode: "0x",
      callData: voteData,
      accountGasLimits,
      preVerificationGas: ethers.toBeHex(100_000, 32),
      gasFees,
      paymasterAndData: await paymaster.getAddress(),
      signature: "0x",
    };

    const tx = await entryPoint.handleOps([userOp], council1.address);
    await tx.wait();

    expect(await dao.hasVoted(proposalId, await smartAccount.getAddress())).to.equal(true);
  });

  it("should revert gasless op if Paymaster deposit empty", async function () {
    // drain Paymaster deposit
    await paymaster.withdrawETH(council1.address, ethers.parseEther("2"));

    const calldata = nft.interface.encodeFunctionData("pause");

    const accountGasLimits = ethers.solidityPacked(
      ["uint128", "uint128"],
      [500_000, 3_000_000]
    );
    const gasFees = ethers.solidityPacked(
      ["uint128", "uint128"],
      [
        ethers.parseUnits("1", "gwei"), // priority
        ethers.parseUnits("5", "gwei"), // max
      ]
    );

    const userOp = {
      sender: await smartAccount.getAddress(),
      nonce: await entryPoint.getNonce(await smartAccount.getAddress(), 0),
      initCode: "0x",
      callData: calldata,
      accountGasLimits,
      preVerificationGas: ethers.toBeHex(100_000, 32),
      gasFees,
      paymasterAndData: await paymaster.getAddress(),
      signature: "0x",
    };

    await expect(entryPoint.handleOps([userOp], council1.address)).to.be.reverted;
  });
});*/
