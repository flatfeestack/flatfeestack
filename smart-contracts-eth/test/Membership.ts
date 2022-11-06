import { mine } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { deployMembershipContract } from "./helpers/deployContracts";

describe("Membership", () => {
  async function deployFixture() {
    const [representative, whitelisterOne, whitelisterTwo, newUser] =
      await ethers.getSigners();

    const { membership, wallet } = await deployMembershipContract(
      representative,
      whitelisterOne,
      whitelisterTwo
    );

    return {
      representative,
      whitelisterOne,
      whitelisterTwo,
      newUser,
      membership,
      wallet,
    };
  }

  async function deployFixtureWhitelisted() {
    const [representative, whitelisterOne, whitelisterTwo, newUserWhitelisted] =
      await ethers.getSigners();
    const { membership, wallet } = await deployMembershipContract(
      representative,
      whitelisterOne,
      whitelisterTwo
    );

    await membership.connect(newUserWhitelisted).requestMembership();
    await membership
      .connect(whitelisterOne)
      .whitelistMember(newUserWhitelisted.address);
    await membership
      .connect(whitelisterTwo)
      .whitelistMember(newUserWhitelisted.address);

    return {
      representative,
      whitelisterOne,
      whitelisterTwo,
      newUserWhitelisted,
      membership,
      wallet,
    };
  }

  describe("requestMembership", () => {
    it("request membership emits event", async () => {
      const { newUser, membership } = await deployFixture();
      await expect(membership.connect(newUser).requestMembership())
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 1);
    });

    it("member can't request membership again", async () => {
      const { membership } = await deployFixture();
      await expect(membership.requestMembership()).to.be.revertedWith(
        "only non-members"
      );
    });
  });

  describe("isWhitelister", () => {
    it("representative is not whitelister", async () => {
      const { representative, membership } = await deployFixture();
      expect(await membership.isWhitelister(representative.address)).to.equal(
        false
      );
    });

    it("whitelister is whitelister", async () => {
      const { whitelisterOne, whitelisterTwo, membership } =
        await deployFixture();
      expect(await membership.isWhitelister(whitelisterOne.address)).to.equal(
        true
      );
      expect(await membership.isWhitelister(whitelisterTwo.address)).to.equal(
        true
      );
    });
  });

  describe("addWhitelister", () => {
    it("can not add whitelister who is already a whitelister", async () => {
      const { representative, whitelisterOne, membership } =
        await deployFixture();
      await expect(
        membership
          .connect(representative)
          .addWhitelister(whitelisterOne.address)
      ).to.be.revertedWith("Is already whitelister!");
    });

    it("to become whitelister you must be a member", async () => {
      const { representative, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(representative).addWhitelister(newUser.address)
      ).to.be.revertedWith("A whitelister must be a member");
    });

    it("a representative can't become a whitelister", async () => {
      const { representative, membership } = await deployFixture();
      await expect(
        membership
          .connect(representative)
          .addWhitelister(representative.address)
      ).to.be.revertedWith("Can't become whitelister!");
    });

    it("member can be added as whitelister by representative", async () => {
      const { representative, newUserWhitelisted, membership } =
        await deployFixtureWhitelisted();
      await expect(
        membership
          .connect(representative)
          .addWhitelister(newUserWhitelisted.address)
      )
        .to.emit(membership, "ChangeInWhiteLister")
        .withArgs(newUserWhitelisted.address, true);
    });
  });

  describe("whitelistMember", () => {
    it("non member can not be whitelisted", async () => {
      const { whitelisterOne, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(whitelisterOne).whitelistMember(newUser.address)
      ).to.be.revertedWith("Invalid member status!");
    });

    it("requesting member gets whitelisted by one whitelister", async () => {
      const { whitelisterOne, newUser, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await expect(
        membership.connect(whitelisterOne).whitelistMember(newUser.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 2);
    });

    it("requesting member gets whitelisted by second whitelister", async () => {
      const { whitelisterOne, whitelisterTwo, newUser, membership } =
        await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await expect(
        membership.connect(whitelisterTwo).whitelistMember(newUser.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 3);

      expect(await membership.nextMembershipFeePayment(newUser.address)).to.eq(
        0
      );
    });

    it("requesting member can not get whitelisted by same whitelister", async () => {
      const { whitelisterOne, newUser, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await expect(
        membership.connect(whitelisterOne).whitelistMember(newUser.address)
      ).to.be.revertedWith("Invalid member status!");
    });
  });

  describe("removeWhitelister", () => {
    it("the to be removed address must be a whitelister", async () => {
      const { representative, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(representative).removeWhitelister(newUser.address)
      ).to.be.revertedWith("Is no whitelister!");
    });

    it("can not remove whitelister if number of whitelisters becomes less than minimum number of whitelisters", async () => {
      const { representative, whitelisterOne, membership } =
        await deployFixture();
      await expect(
        membership
          .connect(representative)
          .removeWhitelister(whitelisterOne.address)
      ).to.be.revertedWith("Minimum whitelister not met!");
    });

    it("whitelister can be removed by representative", async () => {
      const { representative, whitelisterOne, newUserWhitelisted, membership } =
        await deployFixtureWhitelisted();
      await membership
        .connect(representative)
        .addWhitelister(newUserWhitelisted.address);
      await expect(
        membership
          .connect(representative)
          .removeWhitelister(whitelisterOne.address)
      )
        .to.emit(membership, "ChangeInWhiteLister")
        .withArgs(whitelisterOne.address, false);

      expect(await membership.isWhitelister(whitelisterOne.address)).to.equal(
        false
      );
      expect(
        await membership.isWhitelister(newUserWhitelisted.address)
      ).to.equal(true);
    });
  });

  describe("payMembershipFee", () => {
    it("cannot be called by non-members", async () => {
      const { membership, newUser } = await deployFixture();
      await expect(
        membership.connect(newUser).payMembershipFee()
      ).to.be.revertedWith("only members");
    });

    it("cannot be called by requesting members", async () => {
      const { newUser, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await expect(
        membership.connect(newUser).payMembershipFee()
      ).to.be.revertedWith("only members");
    });

    it("cannot be called by members with one whitelister approval", async () => {
      const { newUser, membership, whitelisterOne } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await expect(
        membership.connect(newUser).payMembershipFee()
      ).to.be.revertedWith("only members");
    });

    it("reverts if payment amount doesn't cover membership fee", async () => {
      const { newUser, membership, whitelisterOne, whitelisterTwo } =
        await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await membership.connect(whitelisterTwo).whitelistMember(newUser.address);

      await expect(
        membership.connect(newUser).payMembershipFee({
          value: ethers.utils.parseUnits("3", 3),
        })
      ).to.be.revertedWith("Membership fee not covered!");
    });

    it("allows to pay membership fees", async () => {
      const { newUser, membership, wallet, whitelisterOne, whitelisterTwo } =
        await deployFixture();
      const toBePaid = ethers.utils.parseUnits("3", 4); // exactly 30k wei

      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await membership.connect(whitelisterTwo).whitelistMember(newUser.address);

      await expect(
        membership.connect(newUser).payMembershipFee({
          value: toBePaid,
        })
      )
        .to.emit(wallet, "AcceptPayment")
        .withArgs(newUser.address, toBePaid);

      const blockNumBefore = await ethers.provider.getBlockNumber();
      const blockBefore = await ethers.provider.getBlock(blockNumBefore);

      expect(
        await membership.nextMembershipFeePayment(newUser.address)
      ).to.greaterThan(blockBefore.timestamp);

      expect(await wallet.individualContribution(newUser.address)).to.eq(
        toBePaid
      );
    });
  });

  describe("setMembershipFee", () => {
    it("allows to set the membership fee", async () => {
      const { representative, membership } = await deployFixture();

      await membership
        .connect(representative)
        .setMembershipFee(ethers.utils.parseUnits("1", 1));

      expect(await membership.membershipFee()).to.eq(
        ethers.utils.parseUnits("1", 1)
      );
    });
  });

  describe("setRepresentative", () => {
    it("non member can't become representative", async () => {
      const { newUser, membership } = await deployFixture();
      await expect(
        membership.setRepresentative(newUser.address)
      ).to.be.revertedWith("Address is not a member!");
    });

    it("requesting member can't become representative", async () => {
      const { newUser, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await expect(
        membership.setRepresentative(newUser.address)
      ).to.be.revertedWith("Address is not a member!");
    });

    it("requesting member whitelisted by one can't become representative", async () => {
      const { newUser, whitelisterOne, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await expect(
        membership.setRepresentative(newUser.address)
      ).to.be.revertedWith("Address is not a member!");
    });

    it("representative can't become representative", async () => {
      const { representative, membership } = await deployFixture();
      await expect(
        membership.setRepresentative(representative.address)
      ).to.be.revertedWith("Address is the representative!");
    });

    it("set new representative emits event", async () => {
      const { representative, newUserWhitelisted, membership } =
        await deployFixtureWhitelisted();
      await expect(membership.setRepresentative(newUserWhitelisted.address))
        .to.emit(membership, "ChangeInRepresentative")
        .withArgs(newUserWhitelisted.address, true)
        .and.to.emit(membership, "ChangeInRepresentative")
        .withArgs(representative.address, false);
    });
  });

  describe("removeMember", () => {
    it("cannot leave association if they're no member", async () => {
      const { newUser, membership } = await deployFixture();

      await expect(
        membership.connect(newUser).removeMember(newUser.address)
      ).to.be.revertedWith("Address is not a member!");
    });

    it("can leave association if they request membership", async () => {
      const { newUser, membership } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await expect(membership.connect(newUser).removeMember(newUser.address))
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 0);
    });

    it("can leave association if they were whitelisted by one member", async () => {
      const { newUser, membership, whitelisterOne } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);

      await expect(membership.connect(newUser).removeMember(newUser.address))
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 0);
    });

    it("can leave association if they are member", async () => {
      const { newUserWhitelisted, membership } =
        await deployFixtureWhitelisted();

      await expect(
        membership
          .connect(newUserWhitelisted)
          .removeMember(newUserWhitelisted.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUserWhitelisted.address, 0);
    });

    it("can leave association if they are whitelister", async () => {
      const { whitelisterOne, membership, representative, newUserWhitelisted } =
        await deployFixtureWhitelisted();

      // representative a replacement for our whitelisterOne
      await membership
        .connect(representative)
        .addWhitelister(newUserWhitelisted.address);

      await expect(
        membership.connect(whitelisterOne).removeMember(whitelisterOne.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(whitelisterOne.address, 0)
        .and.to.emit(membership, "ChangeInWhiteLister")
        .withArgs(whitelisterOne.address, false);
    });

    it("cannot remove other members if they're a normal member", async () => {
      const { newUserWhitelisted, membership, whitelisterOne } =
        await deployFixtureWhitelisted();

      await expect(
        membership
          .connect(newUserWhitelisted)
          .removeMember(whitelisterOne.address)
      ).to.be.revertedWith("Restricted to representative!");
    });

    it("can remove other members if they're a representative", async () => {
      const { newUserWhitelisted, membership, representative } =
        await deployFixtureWhitelisted();

      await expect(
        membership
          .connect(representative)
          .removeMember(newUserWhitelisted.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUserWhitelisted.address, 0);
    });

    it("representative cannot leave association", async () => {
      const { membership, representative } = await deployFixtureWhitelisted();

      await expect(
        membership.connect(representative).removeMember(representative.address)
      ).to.be.revertedWith("Representative cannot leave!");
    });

    it("whitelister cannot leave if minimum is not met", async () => {
      const { membership, whitelisterOne } = await deployFixtureWhitelisted();

      await expect(
        membership.connect(whitelisterOne).removeMember(whitelisterOne.address)
      ).to.be.revertedWith("Minimum whitelister not met!");
    });
  });

  describe("getVotes", () => {
    it("everyone has same voting right after initialize", async () => {
      const { representative, whitelisterOne, whitelisterTwo, membership } =
        await deployFixture();
      expect(await membership.getVotes(representative.address)).to.equal(1);
      expect(await membership.getVotes(whitelisterOne.address)).to.equal(1);
      expect(await membership.getVotes(whitelisterTwo.address)).to.equal(1);
    });

    it("non-member does not have voting right", async () => {
      const { newUser, membership } = await deployFixture();
      expect(await membership.getVotes(newUser.address)).to.equal(0);
    });

    it("requesting member does not have voting right", async () => {
      const { newUser, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      expect(await membership.getVotes(newUser.address)).to.equal(0);
    });

    it("member whitelisted by one does not have voting right", async () => {
      const { whitelisterOne, newUser, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      expect(await membership.getVotes(newUser.address)).to.equal(0);
    });

    it("new member gets voting power", async () => {
      const { whitelisterOne, whitelisterTwo, newUser, membership } =
        await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await membership.connect(whitelisterTwo).whitelistMember(newUser.address);

      await membership.connect(newUser).payMembershipFee({
        value: ethers.utils.parseUnits("3", 4),
      });
      expect(await membership.getVotes(newUser.address)).to.equal(1);
    });
  });

  describe("getPastVotes", () => {
    it("new member never had voting power in the past", async () => {
      const { whitelisterOne, whitelisterTwo, newUser, membership } =
        await deployFixture();
      const firstBlock = await ethers.provider.getBlockNumber();
      await membership.connect(newUser).requestMembership();
      const secondBlock = await ethers.provider.getBlockNumber();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      const thirdBlock = await ethers.provider.getBlockNumber();
      await membership.connect(whitelisterTwo).whitelistMember(newUser.address);
      const fourthBlock = await ethers.provider.getBlockNumber();
      await membership.connect(newUser).payMembershipFee({
        value: ethers.utils.parseUnits("3", 4),
      });
      const fifthBlock = await ethers.provider.getBlockNumber();
      await mine(5);
      expect(
        await membership.getPastVotes(newUser.address, firstBlock)
      ).to.equal(0);
      expect(
        await membership.getPastVotes(newUser.address, secondBlock)
      ).to.equal(0);
      expect(
        await membership.getPastVotes(newUser.address, thirdBlock)
      ).to.equal(0);
      expect(
        await membership.getPastVotes(newUser.address, fourthBlock)
      ).to.equal(0);
      expect(
        await membership.getPastVotes(newUser.address, fifthBlock)
      ).to.equal(1);
    });

    it("removed member had voting power in the past but not now", async () => {
      const { newUserWhitelisted, membership } =
        await deployFixtureWhitelisted();
      await membership.connect(newUserWhitelisted).payMembershipFee({
        value: ethers.utils.parseUnits("3", 4),
      });
      const firstBlock = await ethers.provider.getBlockNumber();
      await membership
        .connect(newUserWhitelisted)
        .removeMember(newUserWhitelisted.address);
      const secondBlock = await ethers.provider.getBlockNumber();

      await mine(2);
      expect(
        await membership.getPastVotes(newUserWhitelisted.address, firstBlock)
      ).to.equal(1);
      expect(
        await membership.getPastVotes(newUserWhitelisted.address, secondBlock)
      ).to.equal(0);
    });
  });

  describe("getPastTotalSupply", () => {
    it("supply is right after initialization", async () => {
      const { membership } = await deployFixture();
      const firstBlock = await ethers.provider.getBlockNumber();
      await mine(1);
      expect(await membership.getPastTotalSupply(firstBlock)).to.equal(3);
    });

    it("after adding new member past supply doesn't change", async () => {
      const { whitelisterOne, whitelisterTwo, newUser, membership } =
        await deployFixture();
      const firstBlock = await ethers.provider.getBlockNumber();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await membership.connect(whitelisterTwo).whitelistMember(newUser.address);
      await membership.connect(newUser).payMembershipFee({
        value: ethers.utils.parseUnits("3", 4),
      });
      const secondBlock = await ethers.provider.getBlockNumber();
      await mine(2);
      expect(await membership.getPastTotalSupply(firstBlock)).to.equal(3);
      expect(await membership.getPastTotalSupply(secondBlock)).to.equal(4);
    });
  });
});
