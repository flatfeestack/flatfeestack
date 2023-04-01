import { keccak256 } from "@ethersproject/keccak256";
import { toUtf8Bytes } from "@ethersproject/strings";
import { mine, mineUpTo, time } from "@nomicfoundation/hardhat-network-helpers";
import type { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import type { Contract } from "ethers";
import { ethers, upgrades } from "hardhat";
import { deployMembershipContract } from "./helpers/deployContracts";
import { addNewMember } from "./helpers/membershipHelpers";

describe("DAO", () => {
  const blocksInAMonth = 201600;
  const blocksInAWeek = 50400;

  async function deployFixture() {
    const [nonMember, firstCouncilMember, secondCouncilMember, regularMember] =
      await ethers.getSigners();

    const { membership, wallet } = await deployMembershipContract(
      firstCouncilMember,
      secondCouncilMember,
      regularMember
    );

    await firstCouncilMember.sendTransaction({
      to: wallet.address,
      value: ethers.utils.parseEther("1.0"),
    });

    // deploy timelock controller
    const Timelock = await ethers.getContractFactory("Timelock");
    const timelock = await upgrades.deployProxy(Timelock, [
      firstCouncilMember.address,
    ]);
    await timelock.deployed();

    // move wallet contract ownership to timelock
    await wallet
      .connect(firstCouncilMember)
      .transferOwnership(timelock.address);

    // move membership contract ownership to timelock
    await membership
      .connect(firstCouncilMember)
      .transferOwnership(timelock.address);

    // deploy DAO
    const DAO = await ethers.getContractFactory("DAO");
    const bylawsHash =
      "3d3cb723c544b48169a908737027aadfdc56540a7b9121e6bf90695e214e209c";
    const bylawsUrl = "https://flatfeestack.github.io/bylaws/";
    const dao = await upgrades.deployProxy(DAO, [
      membership.address,
      timelock.address,
      bylawsHash,
      bylawsUrl,
    ]);
    await dao.deployed();

    // set proper permissions on timelock controller
    const proposerRole = await timelock.PROPOSER_ROLE();
    await timelock
      .connect(firstCouncilMember)
      .grantRole(proposerRole, dao.address);

    const adminRole = await timelock.TIMELOCK_ADMIN_ROLE();
    await timelock
      .connect(firstCouncilMember)
      .revokeRole(adminRole, firstCouncilMember.address);

    // Approve founding proposal
    const initialVotingSlot = await dao.slots(0);
    const events = await dao.queryFilter(
      dao.filters.DAOProposalCreated(
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
    ] = events.find((event: any) => event.event === "DAOProposalCreated")?.args;
    await castVote(
      initialVotingSlot,
      dao,
      firstCouncilMember,
      initialProposalId,
      1
    );
    await castVote(
      initialVotingSlot,
      dao,
      secondCouncilMember,
      initialProposalId,
      1
    );
    await castVote(initialVotingSlot, dao, regularMember, initialProposalId, 1);
    await mine(await dao.votingPeriod());
    const initialProposalArgs = [
      initialTargets,
      initialValues,
      initialCalldatas,
      keccak256(toUtf8Bytes(initialDescription)),
    ];
    await queueProposal(dao, initialProposalArgs);

    await mine(await timelock.getMinDelay());
    await dao.connect(firstCouncilMember).execute(...initialProposalArgs);

    expect(await dao.bylawsHash()).to.eq(bylawsHash);
    expect(await dao.bylawsUrl()).to.eq(bylawsUrl);

    // create proposal slot
    const firstVotingSlot =
      (await time.latestBlock()) + blocksInAMonth + blocksInAWeek;
    await dao.connect(firstCouncilMember).setVotingSlot(firstVotingSlot);

    // create proposal
    const transferCalldata = [
      wallet.interface.encodeFunctionData("increaseAllowance", [
        firstCouncilMember.address,
        ethers.utils.parseEther("1.0"),
      ]),
    ];
    const targets = [wallet.address];
    const values = [0];
    const description = "Give me, the president, some money!";

    const transaction = await dao
      .connect(firstCouncilMember)
      ["propose(address[],uint256[],bytes[],string)"](
        targets,
        values,
        transferCalldata,
        description
      );
    const receipt = await transaction.wait();
    const [proposalId] = receipt.events.find(
      (event: any) => event.event === "DAOProposalCreated"
    ).args;

    return {
      contracts: {
        dao,
        membership,
        timelock,
        wallet,
      },
      entities: {
        nonMember,
        firstCouncilMember,
        secondCouncilMember,
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
      const { dao, wallet } = fixtures.contracts;
      const { nonMember } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [nonMember.address, ethers.utils.parseEther("1.0")]
      );

      await expect(
        dao
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
      const { dao, wallet } = fixtures.contracts;
      const { secondCouncilMember } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [secondCouncilMember.address, ethers.utils.parseEther("1.0")]
      );

      await expect(
        dao
          .connect(secondCouncilMember)
          .propose(
            [wallet.address],
            [0],
            [transferCalldata],
            "I would like to have some money to expand my island in Animal crossing."
          )
      )
        .to.emit(dao, "ProposalCreated")
        .and.to.emit(dao, "DAOProposalCreated");
    });

    it("can propose an extraordinary assembly", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { secondCouncilMember } = fixtures.entities;

      const transferCalldata = dao.interface.encodeFunctionData(
        "setVotingSlot",
        [987654321]
      );

      await expect(
        dao
          .connect(secondCouncilMember)
          .propose(
            [dao.address],
            [0],
            [transferCalldata],
            "I would like to propose an extraordinary assembly to vote stuff."
          )
      )
        .to.emit(dao, "ProposalCreated")
        .and.to.emit(dao, "ExtraOrdinaryAssemblyRequested");

      expect(await dao.getExtraOrdinaryProposalsLength()).to.eq(1);
    });

    it("proposal events emits correct data", async () => {
      const fixtures = await deployFixture();
      const { dao, wallet } = fixtures.contracts;
      const { secondCouncilMember } = fixtures.entities;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [secondCouncilMember.address, ethers.utils.parseEther("1.0")]
      );

      let description = "I would like to have an ExtraordinaryVote.";
      let calldata = [transferCalldata];
      let values = [0];
      let targets = [wallet.address];

      const proposal = await dao
        .connect(secondCouncilMember)
        .propose(targets, values, calldata, description);

      const receiptProposal = await proposal.wait();
      const event = receiptProposal.events.find(
        (event: any) => event.event === "DAOProposalCreated"
      ).args;

      expect(event[1]).to.eq(secondCouncilMember.address);
      expect(event[2]).to.deep.eq(targets, "Targets don't match");
      expect(event[3]).to.deep.eq(values, "Values don't match");
      expect(event[5]).to.deep.eq(calldata, "Calldata don't match");
      expect(event[8]).to.eq(description);
      expect(event[9]).to.eq(0);
    });

    it("proposals get assigned to the correct voting slots", async () => {
      const fixtures = await deployFixture();
      const { dao, wallet } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const { firstVotingSlot } = fixtures;
      const proposalId1 = fixtures.proposal.id;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [firstCouncilMember.address, ethers.utils.parseEther("1.0")]
      );

      await mineUpTo(firstVotingSlot);

      const latestBlock = await time.latestBlock();
      const secondVotingSlot = latestBlock + 2 * blocksInAMonth;
      await dao.connect(firstCouncilMember).setVotingSlot(secondVotingSlot);
      const thirdVotingSlot = latestBlock + 4 * blocksInAMonth;
      await dao.connect(firstCouncilMember).setVotingSlot(thirdVotingSlot);

      const proposal2 = await dao
        .connect(firstCouncilMember)
        ["propose(address[],uint256[],bytes[],string)"](
          [wallet.address],
          [0],
          [transferCalldata],
          "I would like to have some money to buy new plants for the office."
        );
      const receiptProposal2 = await proposal2.wait();
      const [proposalId2] = receiptProposal2.events.find(
        (event: any) => event.event === "DAOProposalCreated"
      ).args;

      const proposal3 = await dao
        .connect(firstCouncilMember)
        ["propose(address[],uint256[],bytes[],string)"](
          [wallet.address],
          [0],
          [transferCalldata],
          "I would like to have some money to buy me a better pc."
        );
      const receiptProposal3 = await proposal3.wait();
      const [proposalId3] = receiptProposal3.events.find(
        (event: any) => event.event === "DAOProposalCreated"
      ).args;

      await mineUpTo(secondVotingSlot);

      const proposal4 = await dao
        .connect(firstCouncilMember)
        ["propose(address[],uint256[],bytes[],string)"](
          [wallet.address],
          [0],
          [transferCalldata],
          "I would like to have some money to build a new feature."
        );
      const receiptProposal4 = await proposal4.wait();
      const [proposalId4] = receiptProposal4.events.find(
        (event: any) => event.event === "DAOProposalCreated"
      ).args;

      expect(await dao.slots(1)).to.eq(firstVotingSlot);
      expect(await dao.slots(2)).to.eq(secondVotingSlot);
      expect(await dao.slots(3)).to.eq(thirdVotingSlot);
      expect(await dao.getSlotsLength()).to.eq(4);
      expect(await dao.votingSlots(firstVotingSlot, 0)).to.eq(proposalId1);
      expect(await dao.getNumberOfProposalsInVotingSlot(firstVotingSlot)).to.eq(
        1
      );
      expect(await dao.votingSlots(secondVotingSlot, 0)).to.eq(proposalId2);
      expect(await dao.votingSlots(secondVotingSlot, 1)).to.eq(proposalId3);
      expect(
        await dao.getNumberOfProposalsInVotingSlot(secondVotingSlot)
      ).to.eq(2);
      expect(await dao.votingSlots(thirdVotingSlot, 0)).to.eq(proposalId4);
      expect(await dao.getNumberOfProposalsInVotingSlot(thirdVotingSlot)).to.eq(
        1
      );
    });

    it("can not propose if there is no voting slot announced", async () => {
      const fixtures = await deployFixture();
      const { dao, wallet } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const transferCalldata = wallet.interface.encodeFunctionData(
        "increaseAllowance",
        [firstCouncilMember.address, ethers.utils.parseEther("1.0")]
      );

      await mineUpTo(firstVotingSlot);

      await expect(
        dao
          .connect(firstCouncilMember)
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
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(dao.connect(firstCouncilMember).castVote(proposalId, 0))
        .to.emit(dao, "VoteCast")
        .withArgs(firstCouncilMember.address, proposalId, 0, 1, "");
    });
  });

  describe("getVotes", () => {
    it("account has votes", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const blockNumber = await time.latestBlock();

      await mineUpTo(firstVotingSlot);

      expect(
        await dao
          .connect(firstCouncilMember)
          .getVotes(firstCouncilMember.address, blockNumber)
      ).to.equal(1);
    });
  });

  describe("getVotesWithParams", () => {
    it("account has votes", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const blockNumber = await time.latestBlock();

      await mineUpTo(firstVotingSlot);

      expect(
        await dao
          .connect(firstCouncilMember)
          .getVotesWithParams(firstCouncilMember.address, blockNumber, 0xaa)
      ).to.equal(1);
    });
  });

  describe("castVoteBySig", () => {
    it("member can not cast vote by a signature", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(
        dao
          .connect(firstCouncilMember)
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
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await mineUpTo(firstVotingSlot);

      await expect(
        dao
          .connect(firstCouncilMember)
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
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      await expect(
        dao
          .connect(firstCouncilMember)
          .castVoteWithReason(proposalId, 0, "No power to the president!")
      ).to.revertedWith("Vote not currently active");
    });

    it("member can cast vote with reason", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const reason =
        "I think it's good that we pay the president a fair share.";

      // votingDelay
      await mineUpTo(firstVotingSlot);

      await expect(
        dao
          .connect(firstCouncilMember)
          .castVoteWithReason(proposalId, 0, reason)
      )
        .to.emit(dao, "VoteCast")
        .withArgs(firstCouncilMember.address, proposalId, 0, 1, reason);
    });

    it("member cannot cast vote after voting ends", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const delay = firstVotingSlot + (await dao.votingPeriod()) + 1;
      mineUpTo(delay);

      await expect(
        dao
          .connect(firstCouncilMember)
          .castVoteWithReason(proposalId, 0, "No power to the president!")
      ).to.revertedWith("Vote not currently active");
    });
  });

  describe("castVoteWithReasonAndParams", () => {
    it("member cannot cast vote with params before voting starts", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      await expect(
        dao
          .connect(firstCouncilMember)
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
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const reason =
        "I think it's good that we pay the president a fair share.";

      const param = 0xaa;

      // votingDelay
      await mineUpTo(firstVotingSlot);

      await expect(
        dao
          .connect(firstCouncilMember)
          .castVoteWithReasonAndParams(proposalId, 0, reason, param)
      ).to.emit(dao, "VoteCastWithParams");
    });

    it("member cannot cast vote with params after voting ends", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      const delay = firstVotingSlot + (await dao.votingPeriod()) + 1;
      mineUpTo(delay);

      await expect(
        dao
          .connect(firstCouncilMember)
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
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);

      // voting period
      await mine(await dao.votingPeriod());
      await expect(dao.queue(...fixtures.proposal.proposalArgs)).to.emit(
        dao,
        "ProposalQueued"
      );
    });

    it("cannot re-queue proposal", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);
      await mine(await dao.votingPeriod());
      await queueProposal(dao, fixtures.proposal.proposalArgs);

      expect(await dao.connect(firstCouncilMember).state(proposalId)).to.equal(
        5
      );

      await expect(
        queueProposal(dao, fixtures.proposal.proposalArgs)
      ).to.revertedWith("Proposal not successful");
    });
  });

  describe("state", () => {
    it("unknown proposal id should revert ", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = 123;

      await expect(
        dao.connect(firstCouncilMember).state(proposalId)
      ).to.revertedWith("Governor: unknown proposal id");
    });

    it("extra ordinary assembly: proposal is defeated if quorum is not reached", async () => {
      const fixtures = await deployFixture();
      const { dao, membership } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember } = fixtures.entities;
      const membershipFee = ethers.utils.parseUnits("3", 4);

      // our fixtures ships with three confirmed members, therefore one vote is already 33%
      // add 3 additional members, so one vote is equal to a quorum of 17%
      let previousBlock = await ethers.provider.getBlockNumber();
      await mine(1);
      expect(await membership.getPastTotalSupply(previousBlock)).to.eq(3);

      const [
        _nonMember,
        _firstCouncilMember,
        _secondCouncilMember,
        _regularMember,
        newMember1,
        newMember2,
        newMember3,
      ] = await ethers.getSigners();

      await addNewMember(
        newMember1,
        firstCouncilMember,
        secondCouncilMember,
        membership
      );
      await membership.connect(newMember1).payMembershipFee({
        value: membershipFee,
      });

      await addNewMember(
        newMember2,
        firstCouncilMember,
        secondCouncilMember,
        membership
      );
      await membership.connect(newMember2).payMembershipFee({
        value: membershipFee,
      });

      await addNewMember(
        newMember3,
        firstCouncilMember,
        secondCouncilMember,
        membership
      );
      await membership.connect(newMember3).payMembershipFee({
        value: membershipFee,
      });

      previousBlock = await ethers.provider.getBlockNumber();
      await mine(1);
      expect(await membership.getPastTotalSupply(previousBlock)).to.eq(6);

      const transferCalldata = dao.interface.encodeFunctionData(
        "setVotingSlot",
        [987654321]
      );

      const transaction = await dao
        .connect(secondCouncilMember)
        .propose(
          [dao.address],
          [0],
          [transferCalldata],
          "I would like to propose an extraordinary assembly to vote stuff."
        );
      const receipt = await transaction.wait();
      const [proposalId] = receipt.events.find(
        (event: any) => event.event === "ExtraOrdinaryAssemblyRequested"
      ).args;

      expect(await dao.getExtraOrdinaryProposalsLength()).to.eq(1);
      expect(await dao.state(proposalId)).to.eq(0);

      const currentTime = await time.latestBlock();

      castVote(currentTime, dao, firstCouncilMember, proposalId, 0);
      castVote(currentTime, dao, secondCouncilMember, proposalId, 1);

      await mine(await dao.extraOrdinaryAssemblyVotingPeriod());

      expect(await dao.connect(firstCouncilMember).state(proposalId)).to.eq(3);
    });

    it("executed proposal has state executed", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);
      await mine(await dao.votingPeriod());
      await queueProposal(dao, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await dao
        .connect(firstCouncilMember)
        .execute(...fixtures.proposal.proposalArgs);

      expect(await dao.connect(firstCouncilMember).state(proposalId)).to.equal(
        7
      );
    });
  });

  describe("hasVoted", () => {
    it("vote should be persistent", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);

      expect(
        await dao
          .connect(firstCouncilMember)
          .hasVoted(proposalId, firstCouncilMember.address)
      ).to.eq(true);
    });

    it("if user hasn't voted, he hasn't voted", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);

      expect(
        await dao
          .connect(firstCouncilMember)
          .hasVoted(proposalId, secondCouncilMember.address)
      ).to.eq(false);
    });
  });

  describe("proposalVotes", () => {
    it("should load correct proposal votes with one voting", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);

      const result = await dao
        .connect(firstCouncilMember)
        .proposalVotes(proposalId);

      expect(result.againstVotes).to.eq(0);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(0);
    });

    it("should load correct proposal votes with two votes", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await mineUpTo(firstVotingSlot);
      await dao.connect(firstCouncilMember).castVote(proposalId, 0);
      await dao.connect(secondCouncilMember).castVote(proposalId, 1);
      await dao.connect(regularMember).castVote(proposalId, 2);

      const result = await dao.proposalVotes(proposalId);
      expect(result.againstVotes).to.eq(1);
      expect(result.forVotes).to.eq(1);
      expect(result.abstainVotes).to.eq(1);
    });
  });

  describe("relay", () => {
    it("council member can not call the function because he is not the governance", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      await expect(
        dao
          .connect(firstCouncilMember)
          .relay(firstCouncilMember.address, 5, 0xaa)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("countingMode", () => {
    it("counting mode is correct", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;

      expect(await dao.COUNTING_MODE()).to.eq(
        "support=bravo&quorum=for,abstain"
      );
    });
  });

  describe("name", () => {
    it("name is correct", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;

      expect(await dao.name()).to.eq("FlatFeeStack");
    });
  });

  describe("setVotingSlot", () => {
    it("can not be set by regular member", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { regularMember } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAMonth + 1;

      await expect(
        dao.connect(regularMember).setVotingSlot(slot)
      ).to.revertedWith("only council member or governor");
    });

    it("can not set same slot twice", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      const slot = (await time.latestBlock()) + 3 * blocksInAMonth;

      await dao.connect(firstCouncilMember).setVotingSlot(slot);

      await expect(
        dao.connect(firstCouncilMember).setVotingSlot(slot)
      ).to.revertedWith("Vote slot already exists");
    });

    it("can not set slot too late", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAWeek + 1;

      await expect(
        dao.connect(firstCouncilMember).setVotingSlot(slot)
      ).to.revertedWith("Announcement too late!");
    });

    it("emits event after setting new slot", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      const slot = (await time.latestBlock()) + blocksInAMonth + 1;

      await expect(dao.connect(firstCouncilMember).setVotingSlot(slot))
        .to.emit(dao, "NewTimeslotSet")
        .withArgs(slot);
    });

    it("can be requested using a proposal", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { regularMember, firstCouncilMember, secondCouncilMember } =
        fixtures.entities;

      const proposedVotingSlotTime = 987654321;
      const transferCalldatas = [
        dao.interface.encodeFunctionData("setVotingSlot", [
          proposedVotingSlotTime,
        ]),
      ];

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        regularMember,
        await time.latestBlock(),
        [firstCouncilMember, secondCouncilMember],
        [],
        [],
        transferCalldatas,
        [dao.address],
        [0],
        "I would like to have an extraordinary voting slot",
        dao.extraOrdinaryAssemblyVotingPeriod
      );

      await mine(await timelock.getMinDelay());

      await expect(dao.execute(...proposalArgs))
        .to.emit(dao, "NewTimeslotSet")
        .withArgs(proposedVotingSlotTime);
    });

    it("keeps the slots in sorted order", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      const slotInThreeMonths =
        (await time.latestBlock()) + 3 * blocksInAMonth + 1;
      await dao.connect(firstCouncilMember).setVotingSlot(slotInThreeMonths);

      const slotInAMonth = (await time.latestBlock()) + 1 * blocksInAMonth + 1;
      await dao.connect(firstCouncilMember).setVotingSlot(slotInAMonth);

      const slotInTwoMonths =
        (await time.latestBlock()) + 2 * blocksInAMonth + 1;
      await dao.connect(firstCouncilMember).setVotingSlot(slotInTwoMonths);

      expect(await dao.slots(1)).to.eq(slotInAMonth);
      expect(await dao.slots(2)).to.eq(fixtures.firstVotingSlot);
      expect(await dao.slots(3)).to.eq(slotInTwoMonths);
      expect(await dao.slots(4)).to.eq(slotInThreeMonths);
    });
  });

  describe("execute", () => {
    it("proposal can be executed by other than the proposer", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, regularMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, regularMember, proposalId, 1);
      await mine(await dao.votingPeriod());
      await queueProposal(dao, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await expect(
        dao
          .connect(firstCouncilMember)
          .execute(...fixtures.proposal.proposalArgs)
      )
        .to.emit(dao, "ProposalExecuted")
        .withArgs(proposalId);

      expect(await dao.state(proposalId)).to.equal(7);
    });

    it("cannot execute without queueing", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);

      await mine(await dao.votingPeriod());
      await expect(
        dao.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("TimelockController: operation is not ready");

      expect(await dao.state(proposalId)).to.equal(4);
    });

    it("cannot execute proposal too early", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);
      await mine(await dao.votingPeriod());
      await queueProposal(dao, fixtures.proposal.proposalArgs);

      await expect(
        dao.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("TimelockController: operation is not ready");

      expect(await dao.state(proposalId)).to.equal(5);
    });

    it("cannot re-execute proposal", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      // votingDelay
      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);
      await mine(await dao.votingPeriod());
      await queueProposal(dao, fixtures.proposal.proposalArgs);

      await mine(await timelock.getMinDelay());
      await expect(
        dao
          .connect(secondCouncilMember)
          .execute(...fixtures.proposal.proposalArgs)
      )
        .to.emit(dao, "ProposalExecuted")
        .withArgs(proposalId);

      expect(await dao.state(proposalId)).to.equal(7);

      await expect(
        dao.execute(...fixtures.proposal.proposalArgs)
      ).to.revertedWith("Proposal not successful");
    });
  });

  describe("updateTimelock", () => {
    it("is protected", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;

      await expect(dao.updateTimelock(timelock.address)).to.revertedWith(
        "Governor: onlyGovernance"
      );
    });
  });

  describe("timelock", () => {
    it("returns address of timelock controller", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;

      expect(await dao.timelock()).to.eq(timelock.address);
    });
  });

  describe("proposalEta", () => {
    it("returns timestamp when proposal can be executed", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;
      const proposalId = fixtures.proposal.id;
      const { firstVotingSlot } = fixtures;

      await castVote(firstVotingSlot, dao, firstCouncilMember, proposalId, 1);
      await mine(await dao.votingPeriod());
      await queueProposal(dao, fixtures.proposal.proposalArgs);

      expect((await dao.proposalEta(proposalId)).toNumber()).to.eq(
        (await time.latest()) + 86400
      );
    });

    it("returns 0 if proposal is unknown", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const proposalId = fixtures.proposal.id;

      expect((await dao.proposalEta(proposalId)).toNumber()).to.eq(0);
    });
  });

  describe("cancelVotingSlot", () => {
    it("reverts if sender is not a council member", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.dao
          .connect(fixtures.entities.regularMember)
          .cancelVotingSlot(1234, "")
      ).to.revertedWith("only council member");
    });

    it("cannot cancel too late", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.dao
          .connect(fixtures.entities.firstCouncilMember)
          .cancelVotingSlot(await time.latestBlock(), "")
      ).to.revertedWith("Must be a day before slot!");
    });

    it("cannot cancel non-existent voting slot", async () => {
      const fixtures = await deployFixture();

      await expect(
        fixtures.contracts.dao
          .connect(fixtures.entities.firstCouncilMember)
          .cancelVotingSlot((await time.latestBlock()) + 10000, "")
      ).to.revertedWith("Voting slot does not exist!");
    });

    it("cancels voting slots without proposals", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      // create a new voting slot
      const secondVotingSlot = (await time.latestBlock()) + 2 * blocksInAMonth;
      await dao.connect(firstCouncilMember).setVotingSlot(secondVotingSlot);
      expect(await dao.getSlotsLength()).to.eq(3);

      const reason = "no proposals there for this voting slot!";

      await expect(
        dao
          .connect(firstCouncilMember)
          .cancelVotingSlot(secondVotingSlot, reason)
      )
        .to.emit(dao, "VotingSlotCancelled")
        .withArgs(secondVotingSlot, reason);

      expect(await dao.getSlotsLength()).to.eq(2);
    });

    it("cancels voting slots and moves proposals", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      // create a new voting slot
      const secondVotingSlot = (await time.latestBlock()) + 2 * blocksInAMonth;
      await dao.connect(firstCouncilMember).setVotingSlot(secondVotingSlot);
      expect(await dao.getSlotsLength()).to.eq(3);

      const reason = "I feel it's too early to vote on these matters.";

      await expect(
        dao
          .connect(firstCouncilMember)
          .cancelVotingSlot(fixtures.firstVotingSlot, reason)
      )
        .to.emit(dao, "VotingSlotCancelled")
        .withArgs(fixtures.firstVotingSlot, reason)
        .and.to.emit(dao, "ProposalVotingTimeChanged")
        .withArgs(
          fixtures.proposal.id,
          fixtures.firstVotingSlot,
          secondVotingSlot
        );

      expect(await dao.getSlotsLength()).to.eq(2);
    });
  });

  describe("setNewBylaws", () => {
    it("can set new bylaws via proposal", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newHash =
        "0466442ae9a903c3028fbea8cb271e7e1ca0ac0ea51ab8823955d3c7e93809b4";
      const newUrl = "https://www.nyan.cat/";

      const transferCalldatas = [
        dao.interface.encodeFunctionData("setNewBylaws", [newHash, newUrl]),
      ];

      const targets = [dao.address];
      const values = [0];
      const description = "I would like to change the bylaws.";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await expect(dao.connect(firstCouncilMember).execute(...proposalArgs))
        .to.emit(dao, "BylawsChanged")
        .withArgs(newUrl, newHash);
    });

    it("bylaws hash can not be set directly", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      const newHash =
        "0466442ae9a903c3028fbea8cb271e7e1ca0ac0ea51ab8823955d3c7e93809b4";
      const newUrl = "https://www.nyan.cat/";

      await expect(
        dao.connect(firstCouncilMember).setNewBylaws(newHash, newUrl)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("setSlotCloseTime", () => {
    it("can set new slot close time", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newSlotCloseTime = 100800;
      const transferCalldatas = [
        dao.interface.encodeFunctionData("setSlotCloseTime", [
          newSlotCloseTime,
        ]),
      ];

      const targets = [dao.address];
      const values = [0];
      const description = "I would like to expand the slot close time.";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await dao.connect(firstCouncilMember).execute(...proposalArgs);
      expect(await dao.connect(firstCouncilMember).slotCloseTime()).to.eq(
        newSlotCloseTime
      );
    });

    it("slot close time can not be set directly", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      const newSlotCloseTime = 100800;
      await expect(
        dao.connect(firstCouncilMember).setSlotCloseTime(newSlotCloseTime)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("setExtraOrdinaryAssemblyVotingPeriod", () => {
    it("can set new voting period for extra ordinary assemblies", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newExtraOrdinaryAssemblyVotingPeriod = 9001;
      const transferCalldatas = [
        dao.interface.encodeFunctionData(
          "setExtraOrdinaryAssemblyVotingPeriod",
          [newExtraOrdinaryAssemblyVotingPeriod]
        ),
      ];

      const targets = [dao.address];
      const values = [0];
      const description =
        "I think the time to vote on an extra ordinary assembly should be shorter.";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await dao.connect(firstCouncilMember).execute(...proposalArgs);
      expect(
        await dao
          .connect(firstCouncilMember)
          .extraOrdinaryAssemblyVotingPeriod()
      ).to.eq(newExtraOrdinaryAssemblyVotingPeriod);
    });

    it("cannot be changed directly", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      await expect(
        dao
          .connect(firstCouncilMember)
          .setExtraOrdinaryAssemblyVotingPeriod(9876543)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("getMinDelay", () => {
    it("should return minimum delay of timelock controller", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;

      expect(await dao.getMinDelay()).to.eq(await timelock.getMinDelay());
    });
  });

  describe("setVotingSlotAnnouncementPeriod", () => {
    it("can set new announcement period for voting slots", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newVotingSlotAnnouncementPeriod = 123456;
      const transferCalldatas = [
        dao.interface.encodeFunctionData("setVotingSlotAnnouncementPeriod", [
          newVotingSlotAnnouncementPeriod,
        ]),
      ];

      const targets = [dao.address];
      const values = [0];
      const description = "Reducing this time would be nicer";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await dao.connect(firstCouncilMember).execute(...proposalArgs);
      expect(
        await dao.connect(firstCouncilMember).votingSlotAnnouncementPeriod()
      ).to.eq(newVotingSlotAnnouncementPeriod);
    });

    it("cannot be changed directly", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      await expect(
        dao.connect(firstCouncilMember).setVotingSlotAnnouncementPeriod(9876543)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("dissolveDAO", () => {
    it("can dissolve the dao", async () => {
      const fixtures = await deployFixture();
      const { dao, membership, wallet, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      expect(await dao.connect(firstCouncilMember).daoActive()).to.eq(true);

      const transferCalldatas = [
        wallet.interface.encodeFunctionData("liquidate", [
          firstCouncilMember.address,
        ]),
        membership.interface.encodeFunctionData("lockMembership"),
        dao.interface.encodeFunctionData("dissolveDAO"),
      ];

      const targets = [wallet.address, membership.address, dao.address];
      const values = [0, 0, 0];
      const description =
        "I want to dissolve the DAO and the firstCouncilMember is the liquidator.";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await dao.connect(firstCouncilMember).execute(...proposalArgs);
      expect(await dao.connect(firstCouncilMember).daoActive()).to.eq(false);
    });

    it("can not only liquidate", async () => {
      const fixtures = await deployFixture();
      const { dao, wallet } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      expect(await dao.connect(firstCouncilMember).daoActive()).to.eq(true);

      const transferCalldatas = [
        wallet.interface.encodeFunctionData("liquidate", [
          firstCouncilMember.address,
        ]),
      ];

      const targets = [wallet.address];
      const values = [0];
      const description = "I want to only liquidate to steal the money";

      await expect(
        dao
          .connect(firstCouncilMember)
          .propose(targets, values, transferCalldatas, description)
      ).to.be.revertedWith("Wrong functions");
    });
  });

  describe("setExtraordinaryVoteQuorumNominator", () => {
    it("can set new value", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newValue = 15;
      const transferCalldatas = [
        dao.interface.encodeFunctionData(
          "setExtraordinaryVoteQuorumNominator",
          [newValue]
        ),
      ];

      const targets = [dao.address];
      const values = [0];
      const description = "I want to increase the value";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await dao.connect(firstCouncilMember).execute(...proposalArgs);
      expect(
        await dao.connect(firstCouncilMember).extraordinaryVoteQuorumNominator()
      ).to.eq(newValue);
    });

    it("can not be more than 20", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newValue = 21;
      const transferCalldatas = [
        dao.interface.encodeFunctionData(
          "setExtraordinaryVoteQuorumNominator",
          [newValue]
        ),
      ];

      const targets = [dao.address];
      const values = [0];
      const description = "I want to increase the value";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await expect(
        dao.connect(firstCouncilMember).execute(...proposalArgs)
      ).to.be.revertedWith(
        "TimelockController: underlying transaction reverted"
      );
    });

    it("cannot be changed directly", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      await expect(
        dao.connect(firstCouncilMember).setExtraordinaryVoteQuorumNominator(15)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });

  describe("setAssociationDissolutionQuorumNominator", () => {
    it("can set new value", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newValue = 50;
      const transferCalldatas = [
        dao.interface.encodeFunctionData(
          "setAssociationDissolutionQuorumNominator",
          [newValue]
        ),
      ];

      const targets = [dao.address];
      const values = [0];
      const description = "I want to increase the value";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await dao.connect(firstCouncilMember).execute(...proposalArgs);
      expect(
        await dao
          .connect(firstCouncilMember)
          .associationDissolutionQuorumNominator()
      ).to.eq(newValue);
    });

    it("can not be more than 100", async () => {
      const fixtures = await deployFixture();
      const { dao, timelock } = fixtures.contracts;
      const { firstCouncilMember, secondCouncilMember, regularMember } =
        fixtures.entities;
      const { firstVotingSlot } = fixtures;

      const newValue = 101;
      const transferCalldatas = [
        dao.interface.encodeFunctionData(
          "setAssociationDissolutionQuorumNominator",
          [newValue]
        ),
      ];

      const targets = [dao.address];
      const values = [0];
      const description = "I want to increase the value";

      const proposalArgs = await createQueueAndVoteProposal(
        dao,
        firstCouncilMember,
        firstVotingSlot,
        [firstCouncilMember, secondCouncilMember, regularMember],
        [],
        [],
        transferCalldatas,
        targets,
        values,
        description
      );

      await mine(await timelock.getMinDelay());
      await expect(
        dao.connect(firstCouncilMember).execute(...proposalArgs)
      ).to.be.revertedWith(
        "TimelockController: underlying transaction reverted"
      );
    });

    it("cannot be changed directly", async () => {
      const fixtures = await deployFixture();
      const { dao } = fixtures.contracts;
      const { firstCouncilMember } = fixtures.entities;

      await expect(
        dao
          .connect(firstCouncilMember)
          .setAssociationDissolutionQuorumNominator(15)
      ).to.revertedWith("Governor: onlyGovernance");
    });
  });
});

async function createQueueAndVoteProposal(
  dao: Contract,
  proposingMember: SignerWithAddress,
  voteStart: number,
  forVoters: SignerWithAddress[],
  againstVoters: SignerWithAddress[],
  abstainVoters: SignerWithAddress[],
  transferCalldatas: string[],
  targets: string[],
  values: number[],
  description: string,
  votingPeriod: () => Promise<number> = dao.votingPeriod
) {
  const transaction = await dao
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
    await castVote(voteStart, dao, againstVoter, proposalId, 1);
  }

  for (const forVoter of forVoters) {
    await castVote(voteStart, dao, forVoter, proposalId, 1);
  }

  for (const abstainVoter of abstainVoters) {
    await castVote(voteStart, dao, abstainVoter, proposalId, 1);
  }

  await mine(await votingPeriod());

  await queueProposal(dao, proposalArgs);

  return proposalArgs;
}

async function castVote(
  firstVotingSlot: number,
  dao: Contract,
  member: SignerWithAddress,
  proposalId: string,
  voteType: number // 0 = Against, 1 = For, 2 = Abstain
) {
  if ((await time.latestBlock()) < firstVotingSlot) {
    await mineUpTo(firstVotingSlot);
  }
  await dao.connect(member).castVote(proposalId, voteType);
}

async function queueProposal(dao: Contract, proposalArgs: any[]) {
  await dao.queue(...proposalArgs);
}
