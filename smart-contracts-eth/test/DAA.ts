import { keccak256 } from "@ethersproject/keccak256";
import { toUtf8Bytes } from "@ethersproject/strings";
import { mine, time, mineUpTo } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers, upgrades } from "hardhat";
import { deployMembershipContract } from "./helpers/deployContracts";

describe("DAA", () => {
  const blocksInAMonth = 181860;
  const blocksInAWeek = 45465;
  async function deployFixture() {
    const [nonMember, representative, whitelisterOne, whitelisterTwo] =
      await ethers.getSigners();

    const { membership, wallet } = await deployMembershipContract(
      representative,
      whitelisterOne,
      whitelisterTwo
    );

    await representative.sendTransaction({
      to: wallet.address,
      value: ethers.utils.parseEther("1.0"),
    });

    const DAA = await ethers.getContractFactory("DAA");
    const daa = await upgrades.deployProxy(DAA, [membership.address]);
    await daa.deployed();

    // transfer wallet ownership
    await wallet.connect(representative).transferOwnership(daa.address);

    // create proposal slot
    const firstVotingSlot =
      (await time.latestBlock()) + blocksInAMonth + blocksInAWeek;
    await daa.connect(representative).setVotingSlot(firstVotingSlot);

    // create proposal
    const transferCalldata = [
      wallet.interface.encodeFunctionData("increaseAllowance", [
        representative.address,
        ethers.utils.parseEther("1.0"),
      ]),
    ];
    const targets = [wallet.address];
    const values = [0];
    const description = "Give me, the president, some money!";

    const transaction = await daa
      .connect(representative)
      ["propose(address[],uint256[],bytes[],string)"](
        targets,
        values,
        transferCalldata,
        description
      );
    const receipt = await transaction.wait();
    const [proposalId] = receipt.events.find(
      (event: any) => event.event === "ProposalCreated"
    ).args;

    return {
      contracts: {
        daa,
        membership,
        wallet,
      },
      entities: {
        nonMember,
        representative,
        whitelisterOne,
        whitelisterTwo,
      },
      proposal: {
        callData: transferCalldata,
        description: description,
        id: proposalId,
        targets: targets,
        values: values,
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
      ).to.emit(daa, "ProposalCreated");
    });

    it("proposals get assigned to the correct voting slots", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { representative, whitelisterOne } = fixtures.entities;
      const { firstVotingSlot } = fixtures;
      const proposalId1 = fixtures.proposal.id;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [whitelisterOne.address, ethers.utils.parseEther("1.0")]
      );

      await mineUpTo(firstVotingSlot);

      const latestBlock = await time.latestBlock();
      const secondVotingSlot = latestBlock + 2 * blocksInAMonth;
      await daa.connect(representative).setVotingSlot(secondVotingSlot);
      const thirdVotingSlot = latestBlock + 4 * blocksInAMonth;
      await daa.connect(representative).setVotingSlot(thirdVotingSlot);

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
        (event: any) => event.event === "ProposalCreated"
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
        (event: any) => event.event === "ProposalCreated"
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
        (event: any) => event.event === "ProposalCreated"
      ).args;

      expect(await daa.slots(0)).to.eq(firstVotingSlot);
      expect(await daa.slots(1)).to.eq(secondVotingSlot);
      expect(await daa.slots(2)).to.eq(thirdVotingSlot);
      expect(await daa.votingSlots(firstVotingSlot, 0)).to.eq(proposalId1);
      expect(await daa.votingSlots(secondVotingSlot, 0)).to.eq(proposalId2);
      expect(await daa.votingSlots(secondVotingSlot, 1)).to.eq(proposalId3);
      expect(await daa.votingSlots(thirdVotingSlot, 0)).to.eq(proposalId4);
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

  describe("execute", () => {
    it("successful proposal can be executed", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(representative).castVote(proposalId, 1);

      // voting period
      await mine(await daa.votingPeriod());
      await expect(
        daa
          .connect(representative)
          .execute(
            fixtures.proposal.targets,
            fixtures.proposal.values,
            fixtures.proposal.callData,
            keccak256(toUtf8Bytes(fixtures.proposal.description))
          )
      )
        .to.emit(daa, "ProposalExecuted")
        .withArgs(proposalId);
    });
  });

  describe("state", () => {
    it("unknown proposal id should revert ", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;
      const proposalId = 123;

      await expect(
        daa.connect(representative).state(proposalId)
      ).to.revertedWith("Governor: unknown proposal id");
    });

    it("executed proposal has state exectuted", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(representative).castVote(proposalId, 1);

      // voting period
      await mine(await daa.votingPeriod());
      await daa
        .connect(representative)
        .execute(
          fixtures.proposal.targets,
          fixtures.proposal.values,
          fixtures.proposal.callData,
          keccak256(toUtf8Bytes(fixtures.proposal.description))
        );

      expect(await daa.connect(representative).state(proposalId)).to.equal(7);
    });
  });

  describe("hasVoted", () => {
    it("vote should be persistent", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(representative).castVote(proposalId, 1);

      expect(
        await daa
          .connect(representative)
          .hasVoted(proposalId, representative.address)
      ).to.eq(true);
    });

    it("if user hasn't voted, he hasn't voted", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative, whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(representative).castVote(proposalId, 1);

      expect(
        await daa
          .connect(representative)
          .hasVoted(proposalId, whitelisterOne.address)
      ).to.eq(false);
    });
  });

  describe("proposalVotes", () => {
    it("should load correct proposal votes with one voting", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(representative).castVote(proposalId, 1);

      const result = await daa
        .connect(representative)
        .proposalVotes(proposalId);

      expect(result.againstVotes).to.eq(0);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(0);
    });

    it("should load correct proposal votes with three votings", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative, whitelisterOne, whitelisterTwo } =
        fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(representative).castVote(proposalId, 0);
      await daa.connect(whitelisterOne).castVote(proposalId, 1);
      await daa.connect(whitelisterTwo).castVote(proposalId, 2);

      const result = await daa
        .connect(representative)
        .proposalVotes(proposalId);

      expect(result.againstVotes).to.eq(1);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(1);
    });
  });

  describe("relay", () => {
    it("representative can not call the function because he is not the governance", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;

      await expect(
        daa.connect(representative).relay(representative.address, 5, 0xaa)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("countingMode", () => {
    it("counting mode is correct", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;

      expect(await daa.connect(representative).COUNTING_MODE()).to.eq(
        "support=bravo&quorum=for,abstain"
      );
    });
  });

  describe("name", () => {
    it("name is correct", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;

      expect(await daa.connect(representative).name()).to.eq("FlatFeeStack");
    });
  });

  describe("setVotingSlot", () => {
    // Disabled because DAA doesn't know the delegate
    // it("can not be set by non delegate", async () => {
    //   const fixtures = await deployFixture();
    //   const { daa } = fixtures.contracts;
    //   const { whitelisterOne } = fixtures.entities;
    //
    //   const slot = (await time.latestBlock()) + blocksInAMonth + 1;
    //
    //   await expect(
    //     daa.connect(whitelisterOne).setVotingSlot(slot)
    //   ).to.revertedWith("Only Governor");
    // });

    it("can not set same slot twice", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;

      const slot = (await time.latestBlock()) + 3 * blocksInAMonth;

      await daa.connect(representative).setVotingSlot(slot);

      await expect(
        daa.connect(representative).setVotingSlot(slot)
      ).to.revertedWith("Vote slot already exists");
    });

    it("can not set slot who is less than one month from now", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAWeek + 1;

      await expect(
        daa.connect(representative).setVotingSlot(slot)
      ).to.revertedWith("Must be a least a month from now");
    });

    it("emits event after setting new slot", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAMonth + 1;

      await expect(daa.connect(representative).setVotingSlot(slot))
        .to.emit(daa, "NewTimeslotSet")
        .withArgs(slot);
    });
  });
});
