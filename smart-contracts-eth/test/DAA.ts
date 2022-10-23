import { keccak256 } from "@ethersproject/keccak256";
import { toUtf8Bytes } from "@ethersproject/strings";
import { mine } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers, upgrades } from "hardhat";
import { deployMembershipContract } from "./helpers/deployContracts";

describe("DAA", () => {
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
      },
      proposal: {
        callData: transferCalldata,
        description: description,
        id: proposalId,
        targets: targets,
        values: values,
      },
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
      ).to.revertedWith("Governor: proposer votes below proposal threshold");
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
  });

  describe("castVote", () => {
    it("member can cast vote without reason", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      await mine(await daa.votingDelay());

      await expect(daa.connect(whitelisterOne).castVote(proposalId, 0))
        .to.emit(daa, "VoteCast")
        .withArgs(whitelisterOne.address, proposalId, 0, 1, "");
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
      ).to.revertedWith("Governor: vote not currently active");
    });

    it("member can cast vote with reason", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { whitelisterOne } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      const reason =
        "I think it's good that we pay the president a fair share.";

      // votingDelay
      await mine(await daa.votingDelay());

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

      const delay = (await daa.votingDelay()) + (await daa.votingPeriod()) + 1;
      mine(delay);

      await expect(
        daa
          .connect(whitelisterOne)
          .castVoteWithReason(proposalId, 0, "No power to the president!")
      ).to.revertedWith("Governor: vote not currently active");
    });
  });

  describe("execute", () => {
    it("successful proposal can be executed", async () => {
      const fixtures = await deployFixture();
      const { daa } = fixtures.contracts;
      const { representative } = fixtures.entities;
      const proposalId = fixtures.proposal.id;

      // votingDelay
      await mine(await daa.votingDelay());
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
});
