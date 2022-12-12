import { keccak256 } from "@ethersproject/keccak256";
import { toUtf8Bytes } from "@ethersproject/strings";
import { mine, mineUpTo, time } from "@nomicfoundation/hardhat-network-helpers";
import type { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import type { Contract } from "ethers";
import { ethers, upgrades } from "hardhat";
import { deployMembershipContract } from "./helpers/deployContracts";
import { addNewMember } from "./helpers/membershipHelpers";

describe("DAA", () => {
  const blocksInAMonth = 201600;
  const blocksInAWeek = 50400;

  async function deployFixture() {
    const [nonMember, firstChairman, secondChairman, regularMember] =
      await ethers.getSigners();

    const { membership, wallet } = await deployMembershipContract(
      firstChairman,
      secondChairman,
      regularMember
    );

    await firstChairman.sendTransaction({
      to: wallet.address,
      value: ethers.utils.parseEther("1.0"),
    });

    // deploy timelock controller
    const Timelock = await ethers.getContractFactory("Timelock");
    const timelock = await upgrades.deployProxy(Timelock, [
      firstChairman.address,
    ]);
    await timelock.deployed();

    // move wallet contract ownership to timelock
    await wallet.connect(firstChairman).transferOwnership(timelock.address);

    // deploy DAA
    const DAA = await ethers.getContractFactory("DAA");
    const bylawsHash =
      "3d3cb723c544b48169a908737027aadfdc56540a7b9121e6bf90695e214e209c";
    const daa = await upgrades.deployProxy(DAA, [
      membership.address,
      timelock.address,
      bylawsHash,
    ]);
    await daa.deployed();

    // set proper permissions on timelock controller
    const proposerRole = await timelock.PROPOSER_ROLE();
    await timelock.connect(firstChairman).grantRole(proposerRole, daa.address);

    const adminRole = await timelock.TIMELOCK_ADMIN_ROLE();
    await timelock
      .connect(firstChairman)
      .revokeRole(adminRole, firstChairman.address);

    // Approve founding proposal
    const initialVotingSlot = await daa.slots(0);
    const events = await daa.queryFilter(
      daa.filters.DAAProposalCreated(
        null,
        null,
        null,
        null,
        null,
        null,
        null,
        null,
        null,
        null
      )
    );

    // @ts-ignore
    const [
      initialProposalId,
      sender,
      initialTargets,
      initialValues,
      targetsLength,
      initialCalldatas,
      start,
      end,
      initialDescription,
    ] = events.find((event: any) => event.event === "DAAProposalCreated")?.args;
    await castVote(initialVotingSlot, daa, firstChairman, initialProposalId, 1);
    await castVote(
      initialVotingSlot,
      daa,
      secondChairman,
      initialProposalId,
      1
    );
    await castVote(initialVotingSlot, daa, regularMember, initialProposalId, 1);
    await mine(await daa.votingPeriod());
    const initialProposalArgs = [
      initialTargets,
      initialValues,
      initialCalldatas,
      keccak256(toUtf8Bytes(initialDescription)),
    ];
    await queueProposal(daa, initialProposalArgs);

    await mine(await timelock.getMinDelay());
    await daa.connect(firstChairman).execute(...initialProposalArgs);

    expect(await daa.bylawsHash()).to.eq(bylawsHash);

    // create proposal slot
    const firstVotingSlot =
      (await time.latestBlock()) + blocksInAMonth + blocksInAWeek;
    await daa.connect(firstChairman).setVotingSlot(firstVotingSlot);

    // create proposal
    const transferCalldata = [
      wallet.interface.encodeFunctionData("increaseAllowance", [
        firstChairman.address,
        ethers.utils.parseEther("1.0"),
      ]),
    ];
    const targets = [wallet.address];
    const values = [0];
    const description = "Give me, the president, some money!";

    const transaction = await daa
      .connect(firstChairman)
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
        firstChairman,
        secondChairman,
        regularMember,
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
      const { nonMember } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [nonMember.address, ethers.utils.parseEther("1.0")]
      );

      await expect(
        daa
          .connect(nonMember)
          .propose(
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
      const { secondChairman } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [secondChairman.address, ethers.utils.parseEther("1.0")]
      );

      await expect(
        daa
          .connect(secondChairman)
          .propose(
            [wallet.address],
            [0],
            [transferCalldata],
            "I would like to have some money to expand my island in Animal crossing."
          )
      )
        .to.emit(daa, "ProposalCreated")
        .and.to.emit(daa, "DAAProposalCreated");
    });

    it("can propose an extraordinary assembly", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { secondChairman } = fixtures.entities;

      const transferCalldata = daa.interface.encodeFunctionData(
        "setVotingSlot",
        [987654321]
      );

      await expect(
        daa
          .connect(secondChairman)
          .propose(
            [daa.address],
            [0],
            [transferCalldata],
            "I would like to propose an extraordinary assembly to vote stuff."
          )
      )
        .to.emit(daa, "ProposalCreated")
        .and.to.emit(daa, "ExtraOrdinaryAssemblyRequested");
    });

    it("proposal events emits correct data", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { secondChairman } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [secondChairman.address, ethers.utils.parseEther("1.0")]
      );

      let description = "I would like to have an ExtraordinaryVote.";
      let calldata = [transferCalldata];
      let values = [0];
      let targets = [wallet.address];

      const proposal = await daa
        .connect(secondChairman)
        .propose(targets, values, calldata, description);

      const receiptProposal = await proposal.wait();
      const event = receiptProposal.events.find(
        (event: any) => event.event === "DAAProposalCreated"
      ).args;

      expect(event[1]).to.eq(secondChairman.address);
      expect(event[2]).to.deep.eq(targets, "Targets don't match");
      expect(event[3]).to.deep.eq(values, "Values don't match");
      expect(event[5]).to.deep.eq(calldata, "Calldata don't match");
      expect(event[8]).to.eq(description);
      expect(event[9]).to.eq(0);
    });

    it("proposals get assigned to the correct voting slots", async () => {
      const fixtures = await deployFixture();
      const { daa, wallet } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const { firstVotingSlot } = fixtures;
      const proposalId1 = fixtures.proposal.id;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [firstChairman.address, ethers.utils.parseEther("1.0")]
      );

      await mineUpTo(firstVotingSlot);

      const latestBlock = await time.latestBlock();
      const secondVotingSlot = latestBlock + 2 * blocksInAMonth;
      await daa.connect(firstChairman).setVotingSlot(secondVotingSlot);
      const thirdVotingSlot = latestBlock + 4 * blocksInAMonth;
      await daa.connect(firstChairman).setVotingSlot(thirdVotingSlot);

      const proposal2 = await daa
        .connect(firstChairman)
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
        .connect(firstChairman)
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
        .connect(firstChairman)
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

      expect(await daa.slots(1)).to.eq(firstVotingSlot);
      expect(await daa.slots(2)).to.eq(secondVotingSlot);
      expect(await daa.slots(3)).to.eq(thirdVotingSlot);
      expect(await daa.getSlotsLength()).to.eq(4);
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
      const { firstChairman } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [firstChairman.address, ethers.utils.parseEther("1.0")]
      );

      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(firstChairman)
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
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(daa.connect(firstChairman).castVote(proposalId, 0))
        .to.emit(daa, "VoteCast")
        .withArgs(firstChairman.address, proposalId, 0, 1, "");
    });
  });

  describe("getVotes", () => {
    it("account has votes", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const blockNumber = await time.latestBlock();

      await mineUpTo(firstVotingSlot);

      expect(
        await daa
          .connect(firstChairman)
          .getVotes(firstChairman.address, blockNumber)
      ).to.equal(1);
    });
  });

  describe("getVotesWithParams", () => {
    it("account has votes", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const blockNumber = await time.latestBlock();

      await mineUpTo(firstVotingSlot);

      expect(
        await daa
          .connect(firstChairman)
          .getVotesWithParams(firstChairman.address, blockNumber, 0xaa)
      ).to.equal(1);
    });
  });

  describe("castVoteBySig", () => {
    it("member can not cast vote by a signature", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(firstChairman)
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
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(firstChairman)
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
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      await expect(
        daa
          .connect(firstChairman)
          .castVoteWithReason(proposalId, 0, "No power to the president!")
      ).to.revertedWith("Vote not currently active");
    });

    it("member can cast vote with reason", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const reason =
        "I think it's good that we pay the president a fair share.";

      // votingDelay
      await mineUpTo(firstVotingSlot);

      await expect(
        daa.connect(firstChairman).castVoteWithReason(proposalId, 0, reason)
      )
        .to.emit(daa, "VoteCast")
        .withArgs(firstChairman.address, proposalId, 0, 1, reason);
    });

    it("member cannot cast vote after voting ends", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const delay = firstVotingSlot + (await daa.votingPeriod()) + 1;
      mineUpTo(delay);

      await expect(
        daa
          .connect(firstChairman)
          .castVoteWithReason(proposalId, 0, "No power to the president!")
      ).to.revertedWith("Vote not currently active");
    });
  });

  describe("castVoteWithReasonAndParams", () => {
    it("member cannot cast vote with params before voting starts", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      await expect(
        daa
          .connect(firstChairman)
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
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const reason =
        "I think it's good that we pay the president a fair share.";

      const param = 0xaa;

      // votingDelay
      await mineUpTo(firstVotingSlot);

      await expect(
        daa
          .connect(firstChairman)
          .castVoteWithReasonAndParams(proposalId, 0, reason, param)
      ).to.emit(daa, "VoteCastWithParams");
    });

    it("member cannot cast vote with params after voting ends", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const delay = firstVotingSlot + (await daa.votingPeriod()) + 1;
      mineUpTo(delay);

      await expect(
        daa
          .connect(firstChairman)
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
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);

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
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);
      await mine(await daa.votingPeriod());
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      expect(await daa.connect(firstChairman).state(proposalId)).to.equal(5);

      await expect(
        queueProposal(daa, fixtures.proposal.proposalArgs)
      ).to.revertedWith("Proposal not successful");
    });
  });

  describe("state", () => {
    it("unknown proposal id should revert ", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = 123;

      await expect(
        daa.connect(firstChairman).state(proposalId)
      ).to.revertedWith("Governor: unknown proposal id");
    });

    it("extra ordinary assembly: proposal is defeated if quorum is not reached", async () => {
      const fixtures = await deployFixture();
      const { daa, membership } = fixtures.contracts;
      const { firstChairman, secondChairman } = fixtures.entities;
      const membershipFee = ethers.utils.parseUnits("3", 4);

      // our fixtures ships with three confirmed members, therefore one vote is already 33%
      // add 3 additional members, so one vote is equal to a quorum of 17%
      let previousBlock = await ethers.provider.getBlockNumber();
      await mine(1);
      expect(await membership.getPastTotalSupply(previousBlock)).to.eq(3);

      const [
        _nonMember,
        _firstChairman,
        _secondChairman,
        _regularMember,
        newMember1,
        newMember2,
        newMember3,
      ] = await ethers.getSigners();

      await addNewMember(newMember1, firstChairman, secondChairman, membership);
      await membership.connect(newMember1).payMembershipFee({
        value: membershipFee,
      });

      await addNewMember(newMember2, firstChairman, secondChairman, membership);
      await membership.connect(newMember2).payMembershipFee({
        value: membershipFee,
      });

      await addNewMember(newMember3, firstChairman, secondChairman, membership);
      await membership.connect(newMember3).payMembershipFee({
        value: membershipFee,
      });

      previousBlock = await ethers.provider.getBlockNumber();
      await mine(1);
      expect(await membership.getPastTotalSupply(previousBlock)).to.eq(6);

      const transferCalldata = daa.interface.encodeFunctionData(
        "setVotingSlot",
        [987654321]
      );

      const transaction = await daa
        .connect(secondChairman)
        .propose(
          [daa.address],
          [0],
          [transferCalldata],
          "I would like to propose an extraordinary assembly to vote stuff."
        );
      const receipt = await transaction.wait();
      const [proposalId] = receipt.events.find(
        (event: any) => event.event === "ExtraOrdinaryAssemblyRequested"
      ).args;

      expect(await daa.state(proposalId)).to.eq(0);

      const currentTime = await time.latestBlock();

      castVote(currentTime, daa, firstChairman, proposalId, 0);
      castVote(currentTime, daa, secondChairman, proposalId, 1);

      await mine(await daa.extraOrdinaryAssemblyVotingPeriod());

      expect(await daa.connect(firstChairman).state(proposalId)).to.eq(3);
    });

    it("executed proposal has state executed", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);
      await mine(await daa.votingPeriod());
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await daa
        .connect(firstChairman)
        .execute(...fixtures.proposal.proposalArgs);

      expect(await daa.connect(firstChairman).state(proposalId)).to.equal(7);
    });
  });

  describe("hasVoted", () => {
    it("vote should be persistent", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);

      expect(
        await daa
          .connect(firstChairman)
          .hasVoted(proposalId, firstChairman.address)
      ).to.eq(true);
    });

    it("if user hasn't voted, he hasn't voted", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman, secondChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);

      expect(
        await daa
          .connect(firstChairman)
          .hasVoted(proposalId, secondChairman.address)
      ).to.eq(false);
    });
  });

  describe("proposalVotes", () => {
    it("should load correct proposal votes with one voting", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);

      const result = await daa.connect(firstChairman).proposalVotes(proposalId);

      expect(result.againstVotes).to.eq(0);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(0);
    });

    it("should load correct proposal votes with two votes", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman, secondChairman, regularMember } =
        fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await daa.connect(firstChairman).castVote(proposalId, 0);
      await daa.connect(secondChairman).castVote(proposalId, 1);
      await daa.connect(regularMember).castVote(proposalId, 2);

      const result = await daa.proposalVotes(proposalId);
      expect(result.againstVotes).to.eq(1);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(1);
    });
  });

  describe("relay", () => {
    it("chairman can not call the function because he is not the governance", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      await expect(
        daa.connect(firstChairman).relay(firstChairman.address, 5, 0xaa)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("countingMode", () => {
    it("counting mode is correct", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;

      expect(await daa.COUNTING_MODE()).to.eq(
        "support=bravo&quorum=for,abstain"
      );
    });
  });

  describe("name", () => {
    it("name is correct", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;

      expect(await daa.name()).to.eq("FlatFeeStack");
    });
  });

  describe("setVotingSlot", () => {
    it("can not be set by regular member", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { regularMember } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAMonth + 1;

      await expect(
        daa.connect(regularMember).setVotingSlot(slot)
      ).to.revertedWith("only chairman or governor");
    });

    it("can not set same slot twice", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      const slot = (await time.latestBlock()) + 3 * blocksInAMonth;

      await daa.connect(firstChairman).setVotingSlot(slot);

      await expect(
        daa.connect(firstChairman).setVotingSlot(slot)
      ).to.revertedWith("Vote slot already exists");
    });

    it("can not set slot who is less than one month from now", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAWeek + 1;

      await expect(
        daa.connect(firstChairman).setVotingSlot(slot)
      ).to.revertedWith("Must be a least a month from now");
    });

    it("emits event after setting new slot", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAMonth + 1;

      await expect(daa.connect(firstChairman).setVotingSlot(slot))
        .to.emit(daa, "NewTimeslotSet")
        .withArgs(slot);
    });

    it("can be requested using a proposal", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { regularMember, firstChairman, secondChairman } =
        fixtures.entities;

      const proposedVotingSlotTime = 987654321;
      const transferCalldatas = [
        daa.interface.encodeFunctionData("setVotingSlot", [
          proposedVotingSlotTime,
        ]),
      ];

      const proposalArgs = await createQueueAndVoteProposal(
        daa,
        regularMember,
        await time.latestBlock(),
        [firstChairman, secondChairman],
        [],
        [],
        transferCalldatas,
        [daa.address],
        [0],
        "I would like to have an extraordinary voting slot",
        daa.extraOrdinaryAssemblyVotingPeriod
      );

      await mine(await timelock.getMinDelay());

      await expect(daa.execute(...proposalArgs))
        .to.emit(daa, "NewTimeslotSet")
        .withArgs(proposedVotingSlotTime);
    });

    it("keeps the slots in sorted order", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      const slotInThreeMonths =
        (await time.latestBlock()) + 3 * blocksInAMonth + 1;
      await daa.connect(firstChairman).setVotingSlot(slotInThreeMonths);

      const slotInAMonth = (await time.latestBlock()) + 1 * blocksInAMonth + 1;
      await daa.connect(firstChairman).setVotingSlot(slotInAMonth);

      const slotInTwoMonths =
        (await time.latestBlock()) + 2 * blocksInAMonth + 1;
      await daa.connect(firstChairman).setVotingSlot(slotInTwoMonths);

      expect(await daa.slots(1)).to.eq(slotInAMonth);
      expect(await daa.slots(2)).to.eq(fixtures.firstVotingSlot);
      expect(await daa.slots(3)).to.eq(slotInTwoMonths);
      expect(await daa.slots(4)).to.eq(slotInThreeMonths);
    });
  });

  describe("execute", () => {
    it("proposal can be executed by other than the proposer", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { firstChairman, regularMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, regularMember, proposalId, 1);
      await mine(await daa.votingPeriod());
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await expect(
        daa.connect(firstChairman).execute(...fixtures.proposal.proposalArgs)
      )
        .to.emit(daa, "ProposalExecuted")
        .withArgs(proposalId);

      expect(await daa.state(proposalId)).to.equal(7);
    });

    it("cannot execute without queueing", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);

      await mine(await daa.votingPeriod());
      await expect(
        daa.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("TimelockController: operation is not ready");

      expect(await daa.state(proposalId)).to.equal(4);
    });

    it("cannot execute proposal too early", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);
      await mine(await daa.votingPeriod());
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await expect(
        daa.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("TimelockController: operation is not ready");

      expect(await daa.state(proposalId)).to.equal(5);
    });

    it("cannot re-execute proposal", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { firstChairman, secondChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);
      await mine(await daa.votingPeriod());
      await queueProposal(daa, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await expect(
        daa.connect(secondChairman).execute(...fixtures.proposal.proposalArgs)
      )
        .to.emit(daa, "ProposalExecuted")
        .withArgs(proposalId);

      expect(await daa.state(proposalId)).to.equal(7);

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
      const { firstChairman } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await castVote(firstVotingSlot, daa, firstChairman, proposalId, 1);
      await mine(await daa.votingPeriod());
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
          .connect(fixtures.entities.regularMember)
          .cancelVotingSlot(1234, "")
      ).to.revertedWith("only chairman");
    });

    it("cannot cancel too late", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.daa
          .connect(fixtures.entities.firstChairman)
          .cancelVotingSlot(await time.latestBlock(), "")
      ).to.revertedWith("Must be a day before slot!");
    });

    it("cannot cancel non-existent voting slot", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.daa
          .connect(fixtures.entities.firstChairman)
          .cancelVotingSlot((await time.latestBlock()) + 10000, "")
      ).to.revertedWith("Voting slot does not exist!");
    });

    it("cancels voting slots without proposals", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      // create a new voting slot
      const secondVotingSlot = (await time.latestBlock()) + 2 * blocksInAMonth;
      await daa.connect(firstChairman).setVotingSlot(secondVotingSlot);
      expect(await daa.getSlotsLength()).to.eq(3);

      const reason = "no proposals there for this voting slot!";

      await expect(
        daa.connect(firstChairman).cancelVotingSlot(secondVotingSlot, reason)
      )
        .to.emit(daa, "VotingSlotCancelled")
        .withArgs(secondVotingSlot, reason);

      expect(await daa.getSlotsLength()).to.eq(2);
    });

    it("cancels voting slots and moves proposals", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      // create a new voting slot
      const secondVotingSlot = (await time.latestBlock()) + 2 * blocksInAMonth;
      await daa.connect(firstChairman).setVotingSlot(secondVotingSlot);
      expect(await daa.getSlotsLength()).to.eq(3);

      const reason = "I feel it's too early to vote on these matters.";

      await expect(
        daa
          .connect(firstChairman)
          .cancelVotingSlot(fixtures.firstVotingSlot, reason)
      )
        .to.emit(daa, "VotingSlotCancelled")
        .withArgs(fixtures.firstVotingSlot, reason)
        .and.to.emit(daa, "ProposalVotingTimeChanged")
        .withArgs(
          fixtures.proposal.id,
          fixtures.firstVotingSlot,
          secondVotingSlot
        );

      expect(await daa.getSlotsLength()).to.eq(2);
    });
  });

  describe("setNewBylawsHash", () => {
    it("can set new bylaws hash via proposal", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { firstChairman, secondChairman, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const oldHash = await daa.bylawsHash();
      const newHash =
        "0466442ae9a903c3028fbea8cb271e7e1ca0ac0ea51ab8823955d3c7e93809b4";
      const transferCalldatas = [
        daa.interface.encodeFunctionData("setNewBylawsHash", [newHash]),
      ];

      const targets = [daa.address];
      const values = [0];
      const description = "I would like to change the bylaws.";

      const proposalArgs = await createQueueAndVoteProposal(
        daa,
        firstChairman,
        firstVotingSlot,
        [firstChairman, secondChairman, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await expect(daa.connect(firstChairman).execute(...proposalArgs))
        .to.emit(daa, "BylawsChanged")
        .withArgs(oldHash, newHash);
    });

    it("bylaws hash can not be set directly", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      const newHash =
        "0466442ae9a903c3028fbea8cb271e7e1ca0ac0ea51ab8823955d3c7e93809b4";
      await expect(
        daa.connect(firstChairman).setNewBylawsHash(newHash)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("setSlotCloseTime", () => {
    it("can set new slot close time", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { firstChairman, secondChairman, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newSlotCloseTime = 100800;
      const transferCalldatas = [
        daa.interface.encodeFunctionData("setSlotCloseTime", [
          newSlotCloseTime,
        ]),
      ];

      const targets = [daa.address];
      const values = [0];
      const description = "I would like to expand the slot close time.";

      const proposalArgs = await createQueueAndVoteProposal(
        daa,
        firstChairman,
        firstVotingSlot,
        [firstChairman, secondChairman, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await daa.connect(firstChairman).execute(...proposalArgs);
      expect(await daa.connect(firstChairman).slotCloseTime()).to.eq(
        newSlotCloseTime
      );
    });

    it("slot close time can not be set directly", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      const newSlotCloseTime = 100800;
      await expect(
        daa.connect(firstChairman).setSlotCloseTime(newSlotCloseTime)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("setExtraOrdinaryAssemblyVotingPeriod", () => {
    it("can set new voting period for extra ordinary assemblies", async () => {
      const fixtures = await deployFixture();
      const { daa, timelock } = fixtures.contracts;
      const { firstChairman, secondChairman, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newExtraOrdinaryAssemblyVotingPeriod = 9001;
      const transferCalldatas = [
        daa.interface.encodeFunctionData(
          "setExtraOrdinaryAssemblyVotingPeriod",
          [newExtraOrdinaryAssemblyVotingPeriod]
        ),
      ];

      const targets = [daa.address];
      const values = [0];
      const description =
        "I think the time to vote on an extra ordinary assembly should be shorter.";

      const proposalArgs = await createQueueAndVoteProposal(
        daa,
        firstChairman,
        firstVotingSlot,
        [firstChairman, secondChairman, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await daa.connect(firstChairman).execute(...proposalArgs);
      expect(
        await daa.connect(firstChairman).extraOrdinaryAssemblyVotingPeriod()
      ).to.eq(newExtraOrdinaryAssemblyVotingPeriod);
    });

    it("cannot be changed directly", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { firstChairman } = fixtures.entities;

      await expect(
        daa.connect(firstChairman).setExtraOrdinaryAssemblyVotingPeriod(9876543)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });
});

async function createQueueAndVoteProposal(
  daa: Contract,
  proposingMember: SignerWithAddress,
  voteStart: number,
  forVoters: SignerWithAddress[],
  againstVoters: SignerWithAddress[],
  abstainVoters: SignerWithAddress[],
  transferCalldatas: string[],
  targets: string[],
  values: number[],
  description: string,
  votingPeriod: () => Promise<number> = daa.votingPeriod
) {
  const transaction = await daa
    .connect(proposingMember)
    .propose(targets, values, transferCalldatas, description);

  const proposalArgs = [
    targets,
    values,
    transferCalldatas,
    keccak256(toUtf8Bytes(description)),
  ];

  const receipt = await transaction.wait();
  const [proposalId] = receipt.events.find(
    (event: any) => event.event === "ProposalCreated"
  ).args;

  for (const againstVoter of againstVoters) {
    await castVote(voteStart, daa, againstVoter, proposalId, 1);
  }

  for (const forVoter of forVoters) {
    await castVote(voteStart, daa, forVoter, proposalId, 1);
  }

  for (const abstainVoter of abstainVoters) {
    await castVote(voteStart, daa, abstainVoter, proposalId, 1);
  }

  await mine(await votingPeriod());

  await queueProposal(daa, proposalArgs);

  return proposalArgs;
}

async function castVote(
  firstVotingSlot: number,
  daa: Contract,
  member: SignerWithAddress,
  proposalId: string,
  voteType: number // 0 = Against, 1 = For, 2 = Abstain
) {
  if ((await time.latestBlock()) < firstVotingSlot) {
    await mineUpTo(firstVotingSlot);
  }
  await daa.connect(member).castVote(proposalId, voteType);
}

async function queueProposal(daa: Contract, proposalArgs: any[]) {
  await daa.queue(...proposalArgs);
}
