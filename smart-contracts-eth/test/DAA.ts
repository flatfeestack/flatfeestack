import { keccak256 } from "@ethersproject/keccak256";
import { toUtf8Bytes } from "@ethersproject/strings";
import { mine, mineUpTo, time } from "@nomicfoundation/hardhat-network-helpers";
import type { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import type { Contract } from "ethers";
import { ethers, upgrades } from "hardhat";
import { deployMembershipContract } from "./helpers/deployContracts";

describe("DAA", () => {
  const blocksInAMonth = 201600;
  const blocksInAWeek = 50400;

  async function deployFixture() {
    const [nonMember, chairman, whitelisterOne, whitelisterTwo] =
      await ethers.getSigners();

    const { membership, wallet } = await deployMembershipContract(
      chairman,
      whitelisterOne,
      whitelisterTwo
    );

    await chairman.sendTransaction({
      to: wallet.address,
      value: ethers.utils.parseEther("1.0"),
    });

    // deploy timelock controller
    const Timelock = await ethers.getContractFactory("Timelock");
    const timelock = await upgrades.deployProxy(Timelock, [chairman.address]);
    await timelock.deployed();

    // move wallet contract ownership to timelock
    await wallet.connect(chairman).transferOwnership(timelock.address);

    // deploy DAA
    const DAA = await ethers.getContractFactory("DAA");
    const daa = await upgrades.deployProxy(DAA, [
      membership.address,
      timelock.address,
    ]);
    await daa.deployed();

    // set proper permissions on timelock controller
    const proposerRole = await timelock.PROPOSER_ROLE();
    await timelock.connect(chairman).grantRole(proposerRole, daa.address);

    const adminRole = await timelock.TIMELOCK_ADMIN_ROLE();
    await timelock.connect(chairman).revokeRole(adminRole, chairman.address);

    // create proposal slot
    const firstVotingSlot =
      (await time.latestBlock()) + blocksInAMonth + blocksInAWeek;
    await daa.connect(chairman).setVotingSlot(firstVotingSlot);

    // create proposal
    const transferCalldata = [
      wallet.interface.encodeFunctionData("increaseAllowance", [
        chairman.address,
        ethers.utils.parseEther("1.0"),
      ]),
    ];
    const targets = [wallet.address];
    const values = [0];
    const description = "Give me, the president, some money!";

    const transaction = await daa
      .connect(chairman)
      ["propose(address[],uint256[],bytes[],string)"](
        targets,
        values,
        transferCalldata,
        description
      );
    const receipt = await transaction.wait();
    const [proposalId] = receipt.events.find(
      (event: any) => event.event === "DAAProposalCreated"
    ).args;

    return {
      contracts: {
        daa,
        membership,
        timelock,
        wallet,
      },
      entities: {
        nonMember,
        chairman,
        whitelisterOne,
        whitelisterTwo,
      },
      proposal: {
        callData: transferCalldata,
        description: description,
        id: proposalId,
        targets: targets,
        values: values,
        proposalArgs: [
          targets,
          values,
          transferCalldata,
          keccak256(toUtf8Bytes(description)),
        ],
      },
      firstVotingSlot,
    };
  }

  describe("propose", () => {
    it("cannot create a proposal if they don't have any votes", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { nonMember, whitelisterOne } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [whitelisterOne.address, ethers.utils.parseEther("1.0")]
      );

      await expect(
        daa
          .connect(nonMember)
          ["propose(address[],uint256[],bytes[],string)"](
            [wallet.address],
            [0],
            [transferCalldata],
            "I would like to have some money to expand my island in Animal crossing."
          )
      ).to.revertedWith("Proposer votes below threshold");
    });

    it("can propose a proposal", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [whitelisterOne.address, ethers.utils.parseEther("1.0")]
      );

      await expect(
        daa
          .connect(whitelisterOne)
          ["propose(address[],uint256[],bytes[],string)"](
            [wallet.address],
            [0],
            [transferCalldata],
            "I would like to have some money to expand my island in Animal crossing."
          )
      )
        .to.emit(daa, "ProposalCreated")
        .and.to.emit(daa, "DAAProposalCreated");
    });

    it("can propose a proposal with a category", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [whitelisterOne.address, ethers.utils.parseEther("1.0")]
      );

      await expect(
        daa
          .connect(whitelisterOne)
          ["propose(address[],uint256[],bytes[],string,uint8)"](
            [wallet.address],
            [0],
            [transferCalldata],
            "I would like to have an ExtraordinaryVote.",
            1
          )
      )
        .to.emit(daa, "ProposalCreated")
        .and.to.emit(daa, "DAAProposalCreated");
    });

    it("proposal events emits correct data", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [whitelisterOne.address, ethers.utils.parseEther("1.0")]
      );

      let description = "I would like to have an ExtraordinaryVote.";
      let calldata = [transferCalldata];
      let values = [0];
      let targets = [wallet.address];
      let category = 1;
      const proposal = await daa
        .connect(whitelisterOne)
        ["propose(address[],uint256[],bytes[],string,uint8)"](
          targets,
          values,
          calldata,
          description,
          category
        );

      const receiptProposal = await proposal.wait();
      const event = receiptProposal.events.find(
        (event: any) => event.event === "DAAProposalCreated"
      ).args;

      expect(event[1]).to.eq(whitelisterOne.address);
      expect(event[2]).to.deep.eq(targets, "Targets don't match");
      expect(event[3]).to.deep.eq(values, "Values don't match");
      expect(event[5]).to.deep.eq(calldata, "Calldata don't match");
      expect(event[8]).to.eq(description);
      expect(event[9]).to.eq(category);
    });

    it("proposals get assigned to the correct voting slots", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { chairman, whitelisterOne } = fixtures.entities;
      const { firstVotingSlot } = fixtures;
      const proposalId1 = fixtures.proposal.id;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [whitelisterOne.address, ethers.utils.parseEther("1.0")]
      );

      await mineUpTo(firstVotingSlot);

      const latestBlock = await time.latestBlock();
      const secondVotingSlot = latestBlock + 2 * blocksInAMonth;
      await daa.connect(chairman).setVotingSlot(secondVotingSlot);
      const thirdVotingSlot = latestBlock + 4 * blocksInAMonth;
      await daa.connect(chairman).setVotingSlot(thirdVotingSlot);

      const proposal2 = await daa
        .connect(whitelisterOne)
        ["propose(address[],uint256[],bytes[],string)"](
          [wallet.address],
          [0],
          [transferCalldata],
          "I would like to have some money to buy new plants for the office."
        );
      const receiptProposal2 = await proposal2.wait();
      const [proposalId2] = receiptProposal2.events.find(
        (event: any) => event.event === "DAAProposalCreated"
      ).args;

      const proposal3 = await daa
        .connect(whitelisterOne)
        ["propose(address[],uint256[],bytes[],string)"](
          [wallet.address],
          [0],
          [transferCalldata],
          "I would like to have some money to buy me a better pc."
        );
      const receiptProposal3 = await proposal3.wait();
      const [proposalId3] = receiptProposal3.events.find(
        (event: any) => event.event === "DAAProposalCreated"
      ).args;

      await mineUpTo(secondVotingSlot);

      const proposal4 = await daa
        .connect(whitelisterOne)
        ["propose(address[],uint256[],bytes[],string)"](
          [wallet.address],
          [0],
          [transferCalldata],
          "I would like to have some money to build a new feature."
        );
      const receiptProposal4 = await proposal4.wait();
      const [proposalId4] = receiptProposal4.events.find(
        (event: any) => event.event === "DAAProposalCreated"
      ).args;

      expect(await daa.slots(0)).to.eq(firstVotingSlot);
      expect(await daa.slots(1)).to.eq(secondVotingSlot);
      expect(await daa.slots(2)).to.eq(thirdVotingSlot);
      expect(await daa.getSlotsLength()).to.eq(3);
      expect(await daa.votingSlots(firstVotingSlot, 0)).to.eq(proposalId1);
      expect(await daa.getNumberOfProposalsInVotingSlot(firstVotingSlot)).to.eq(
        1
      );
      expect(await daa.votingSlots(secondVotingSlot, 0)).to.eq(proposalId2);
      expect(await daa.votingSlots(secondVotingSlot, 1)).to.eq(proposalId3);
      expect(
        await daa.getNumberOfProposalsInVotingSlot(secondVotingSlot)
      ).to.eq(2);
      expect(await daa.votingSlots(thirdVotingSlot, 0)).to.eq(proposalId4);
      expect(await daa.getNumberOfProposalsInVotingSlot(thirdVotingSlot)).to.eq(
        1
      );
    });

    it("can not propose if there is no voting slot announced", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [whitelisterOne.address, ethers.utils.parseEther("1.0")]
      );

      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(whitelisterOne)
          ["propose(address[],uint256[],bytes[],string)"](
            [wallet.address],
            [0],
            [transferCalldata],
            "I would like to have some money to expand my island in Animal crossing."
          )
      ).to.revertedWith("No voting slot found");
    });
  });

  describe("castVote", () => {
    it("member can cast vote without reason", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(daa.connect(whitelisterOne).castVote(proposalId, 0))
        .to.emit(daa, "VoteCast")
        .withArgs(whitelisterOne.address, proposalId, 0, 1, "");
    });
  });

  describe("getVotes", () => {
    it("account has votes", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const blockNumber = await time.latestBlock();

      await mineUpTo(firstVotingSlot);

      expect(
        await daa
          .connect(whitelisterOne)
          .getVotes(whitelisterOne.address, blockNumber)
      ).to.equal(1);
    });
  });

  describe("getVotesWithParams", () => {
    it("account has votes", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const blockNumber = await time.latestBlock();

      await mineUpTo(firstVotingSlot);

      expect(
        await daa
          .connect(whitelisterOne)
          .getVotesWithParams(whitelisterOne.address, blockNumber, 0xaa)
      ).to.equal(1);
    });
  });

  describe("castVoteBySig", () => {
    it("member can not cast vote by a signature", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteBySig(
            proposalId,
            0,
            5,
            keccak256(toUtf8Bytes("abc")),
            keccak256(toUtf8Bytes("def"))
          )
      ).to.revertedWith("not possible");
    });
  });

  describe("castVoteWithReasonAndParamsBySig", () => {
    it("member can not cast vote with reason by a signature", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteWithReasonAndParamsBySig(
            proposalId,
            0,
            "reason",
            0xaa,
            5,
            keccak256(toUtf8Bytes("abc")),
            keccak256(toUtf8Bytes("def"))
          )
      ).to.revertedWith("not possible");
    });
  });

  describe("castVoteWithReason", () => {
    it("member cannot cast vote before voting starts", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteWithReason(proposalId, 0, "No power to the president!")
      ).to.revertedWith("Vote not currently active");
    });

    it("member can cast vote with reason", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const reason =
        "I think it's good that we pay the president a fair share.";

      // votingDelay
      await mineUpTo(firstVotingSlot);

      await expect(
        daa.connect(whitelisterOne).castVoteWithReason(proposalId, 0, reason)
      )
        .to.emit(daa, "VoteCast")
        .withArgs(whitelisterOne.address, proposalId, 0, 1, reason);
    });

    it("member cannot cast vote after voting ends", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const delay = firstVotingSlot + (await daa.votingPeriod()) + 1;
      mineUpTo(delay);

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteWithReason(proposalId, 0, "No power to the president!")
      ).to.revertedWith("Vote not currently active");
    });
  });

  describe("castVoteWithReasonAndParams", () => {
    it("member cannot cast vote with params before voting starts", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteWithReasonAndParams(
            proposalId,
            0,
            "No power to the president!",
            0xaa
          )
      ).to.revertedWith("Vote not currently active");
    });

    it("member can cast vote with reason and params", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const reason =
        "I think it's good that we pay the president a fair share.";

      const param = 0xaa;

      // votingDelay
      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteWithReasonAndParams(proposalId, 0, reason, param)
      ).to.emit(daa, "VoteCastWithParams");
    });

    it("member cannot cast vote with params after voting ends", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const delay = firstVotingSlot + (await daa.votingPeriod()) + 1;
      mineUpTo(delay);

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteWithReasonAndParams(
            proposalId,
            0,
            "No power to the president!",
            0xaa
          )
      ).to.revertedWith("Vote not currently active");
    });
  });

  describe("queue", () => {
    it("successful proposal can be queued", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);

      // voting period
      await mine(await daa.votingPeriod());
      await expect(daa.queue(...fixtures.proposal.proposalArgs)).to.emit(
        daa,
        "ProposalQueued"
      );
    });

    it("cannot re-queue proposal", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      expect(await daa.connect(chairman).state(proposalId)).to.equal(5);

      await expect(
        queueProposal(daa, fixtures.proposal.proposalArgs)
      ).to.revertedWith("Proposal not successful");
    });
  });

  describe("state", () => {
    it("unknown proposal id should revert ", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = 123;

      await expect(daa.connect(chairman).state(proposalId)).to.revertedWith(
        "Governor: unknown proposal id"
      );
    });

    it("executed proposal has state executed", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await daa.connect(chairman).execute(...fixtures.proposal.proposalArgs);

      expect(await daa.connect(chairman).state(proposalId)).to.equal(7);
    });
  });

  describe("hasVoted", () => {
    it("vote should be persistent", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);

      expect(
        await daa.connect(chairman).hasVoted(proposalId, chairman.address)
      ).to.eq(true);
    });

    it("if user hasn't voted, he hasn't voted", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman, whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);

      expect(
        await daa.connect(chairman).hasVoted(proposalId, whitelisterOne.address)
      ).to.eq(false);
    });
  });

  describe("proposalVotes", () => {
    it("should load correct proposal votes with one voting", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);

      const result = await daa.connect(chairman).proposalVotes(proposalId);

      expect(result.againstVotes).to.eq(0);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(0);
    });

    it("should load correct proposal votes with three votings", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman, whitelisterOne, whitelisterTwo } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(chairman).castVote(proposalId, 0);
      await daa.connect(whitelisterOne).castVote(proposalId, 1);
      await daa.connect(whitelisterTwo).castVote(proposalId, 2);

      const result = await daa.connect(chairman).proposalVotes(proposalId);

      expect(result.againstVotes).to.eq(1);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(1);
    });
  });

  describe("relay", () => {
    it("chairman can not call the function because he is not the governance", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;

      await expect(
        daa.connect(chairman).relay(chairman.address, 5, 0xaa)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("countingMode", () => {
    it("counting mode is correct", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;

      expect(await daa.connect(chairman).COUNTING_MODE()).to.eq(
        "support=bravo&quorum=for,abstain"
      );
    });
  });

  describe("name", () => {
    it("name is correct", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;

      expect(await daa.connect(chairman).name()).to.eq("FlatFeeStack");
    });
  });

  describe("setVotingSlot", () => {
    it("can not be set by non delegate", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAMonth + 1;

      await expect(
        daa.connect(whitelisterOne).setVotingSlot(slot)
      ).to.revertedWith("only chairman");
    });

    it("can not set same slot twice", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;

      const slot = (await time.latestBlock()) + 3 * blocksInAMonth;

      await daa.connect(chairman).setVotingSlot(slot);

      await expect(daa.connect(chairman).setVotingSlot(slot)).to.revertedWith(
        "Vote slot already exists"
      );
    });

    it("can not set slot who is less than one month from now", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAWeek + 1;

      await expect(daa.connect(chairman).setVotingSlot(slot)).to.revertedWith(
        "Must be a least a month from now"
      );
    });

    it("emits event after setting new slot", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAMonth + 1;

      await expect(daa.connect(chairman).setVotingSlot(slot))
        .to.emit(daa, "NewTimeslotSet")
        .withArgs(slot);
    });
  });

  describe("execute", () => {
    it("proposal can be executed by other than the proposer", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { chairman, whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await expect(
        daa.connect(whitelisterOne).execute(...fixtures.proposal.proposalArgs)
      )
        .to.emit(daa, "ProposalExecuted")
        .withArgs(proposalId);

      expect(await daa.connect(chairman).state(proposalId)).to.equal(7);
    });

    it("cannot execute without queueing", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);

      await mine(await daa.votingPeriod());
      await expect(
        daa.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("TimelockController: operation is not ready");

      expect(await daa.state(proposalId)).to.equal(4);
    });

    it("cannot execute proposal too early", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await expect(
        daa.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("TimelockController: operation is not ready");

      expect(await daa.state(proposalId)).to.equal(5);
    });

    it("cannot re-execute proposal", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { chairman, whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, chairman, proposalId);
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await expect(
        daa.connect(whitelisterOne).execute(...fixtures.proposal.proposalArgs)
      )
        .to.emit(daa, "ProposalExecuted")
        .withArgs(proposalId);

      expect(await daa.connect(chairman).state(proposalId)).to.equal(7);

      await expect(
        daa.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("Proposal not successful");
    });
  });

  describe("updateTimelock", () => {
    it("is protected", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;

      await expect(daa.updateTimelock(timelock.address)).to.revertedWith(
        "Governor: onlyGovernance"
      );
    });
  });

  describe("timelock", () => {
    it("returns address of timelock controller", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;

      expect(await daa.timelock()).to.eq(timelock.address);
    });
  });

  describe("proposalEta", () => {
    it("returns timestamp when proposal can be executed", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await castVote(firstVotingSlot, daa, chairman, proposalId);
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      expect((await daa.proposalEta(proposalId)).toNumber()).to.eq(
        (await time.latest()) + 86400
      );
    });

    it("returns 0 if proposal is unknown", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const proposalId = fixtures.proposal.id;

      expect((await daa.proposalEta(proposalId)).toNumber()).to.eq(0);
    });
  });

  describe("cancelVotingSlot", () => {
    it("reverts if sender is not the chairman", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.daa
          .connect(fixtures.entities.whitelisterOne)
          .cancelVotingSlot(1234)
      ).to.revertedWith("only chairman");
    });

    it("cannot cancel too late", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.daa
          .connect(fixtures.entities.chairman)
          .cancelVotingSlot(await time.latestBlock())
      ).to.revertedWith("Must be a day before slot!");
    });

    it("cannot cancel non-existent voting slot", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.daa
          .connect(fixtures.entities.chairman)
          .cancelVotingSlot((await time.latestBlock()) + 10000)
      ).to.revertedWith("Voting slot does not exist!");
    });

    it("cancels voting slots and moves proposals", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { chairman } = fixtures.entities;

      // create a new voting slot
      const secondVotingSlot = (await time.latestBlock()) + 2 * blocksInAMonth;
      await daa.connect(chairman).setVotingSlot(secondVotingSlot);
      expect(await daa.getSlotsLength()).to.eq(2);

      await expect(
        daa.connect(chairman).cancelVotingSlot(fixtures.firstVotingSlot)
      )
        .to.emit(daa, "VotingSlotCancelled")
        .withArgs(fixtures.firstVotingSlot)
        .and.to.emit(daa, "ProposalVotingTimeChanged")
        .withArgs(
          fixtures.proposal.id,
          fixtures.firstVotingSlot,
          secondVotingSlot
        );

      expect(await daa.getSlotsLength()).to.eq(1);
    });
  });
});

async function castVote(
  firstVotingSlot: number,
  daa: Contract,
  chairman: SignerWithAddress,
  proposalId: string
) {
  await mineUpTo(firstVotingSlot);
  await daa.connect(chairman).castVote(proposalId, 1);
}

async function queueProposal(daa: Contract, proposalArgs: any[]) {
  await mine(await daa.votingPeriod());
  await daa.queue(...proposalArgs);
}
