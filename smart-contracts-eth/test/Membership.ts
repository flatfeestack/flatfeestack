import { mine, time } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  deployMembershipContract,
  deployWalletContract,
} from "./helpers/deployContracts";
import { addNewMember } from "./helpers/membershipHelpers";

describe("Membership", () => {
  async function deployFixture() {
    const [firstCouncilMember, secondCouncilMember, regularMember, newUser] =
      await ethers.getSigners();

    const { membership, wallet } = await deployMembershipContract(
      firstCouncilMember,
      secondCouncilMember,
      regularMember
    );

    return {
      firstCouncilMember,
      secondCouncilMember,
      regularMember,
      newUser,
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

  describe("isCouncilMember", () => {
    it("normal member is not a council member", async () => {
      const { regularMember, membership } = await deployFixture();
      expect(await membership.isCouncilMember(regularMember.address)).to.equal(
        false
      );
    });

    it("a council member is a council member", async () => {
      const { firstCouncilMember, secondCouncilMember, membership } =
        await deployFixture();

      expect(
        await membership.isCouncilMember(firstCouncilMember.address)
      ).to.equal(true);
      expect(
        await membership.isCouncilMember(secondCouncilMember.address)
      ).to.equal(true);
    });
  });

  // this function is only callable by the owner, which is the first council member in the tests, but the timelock controller in reality
  describe("addCouncilMember", () => {
    it("can not add council member who is already a council member", async () => {
      const { firstCouncilMember, secondCouncilMember, membership } =
        await deployFixture();
      await expect(
        membership
          .connect(firstCouncilMember)
          .addCouncilMember(secondCouncilMember.address)
      ).to.be.revertedWith("Is already council member!");
    });

    it("to become council member you must be a member", async () => {
      const { firstCouncilMember, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(firstCouncilMember).addCouncilMember(newUser.address)
      ).to.be.revertedWith("Must be a member");
    });

    it("member can be added as council member by owner", async () => {
      const { firstCouncilMember, regularMember, membership } =
        await deployFixture();
      await expect(
        membership
          .connect(firstCouncilMember)
          .addCouncilMember(regularMember.address)
      )
        .to.emit(membership, "ChangeInCouncilMember")
        .withArgs(regularMember.address, true);
    });
  });

  describe("approveMembership", () => {
    it("non member can not be approved", async () => {
      const { secondCouncilMember, newUser, membership } =
        await deployFixture();
      await expect(
        membership
          .connect(secondCouncilMember)
          .approveMembership(newUser.address)
      ).to.be.revertedWith("Invalid member status!");
    });

    it("requesting member gets approved by one council member", async () => {
      const { firstCouncilMember, newUser, membership } = await deployFixture();

      await membership.connect(newUser).requestMembership();

      await expect(
        membership
          .connect(firstCouncilMember)
          .approveMembership(newUser.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 2);
    });

    it("requesting member gets approved by second council member", async () => {
      const { firstCouncilMember, secondCouncilMember, newUser, membership } =
        await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);

      await expect(
        membership
          .connect(secondCouncilMember)
          .approveMembership(newUser.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 3);

      expect(await membership.nextMembershipFeePayment(newUser.address)).to.eq(
        0
      );
    });

    it("requesting member can not get approved by same council member", async () => {
      const { firstCouncilMember, newUser, membership } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);

      await expect(
        membership
          .connect(firstCouncilMember)
          .approveMembership(newUser.address)
      ).to.be.revertedWith("Invalid member status!");
    });
  });

  describe("removeCouncilMember", () => {
    it("the to be removed address must be a council member", async () => {
      const { firstCouncilMember, newUser, membership } = await deployFixture();
      await expect(
        membership
          .connect(firstCouncilMember)
          .removeCouncilMember(newUser.address)
      ).to.be.revertedWith("Is no council member!");
    });

    it("can not remove council member if number of council members becomes less than minimum amount", async () => {
      const { firstCouncilMember, secondCouncilMember, membership } =
        await deployFixture();

      await expect(
        membership
          .connect(firstCouncilMember)
          .removeCouncilMember(secondCouncilMember.address)
      ).to.be.revertedWith("Min council members not met!");
    });

    it("council members can be removed by contract owner", async () => {
      const {
        firstCouncilMember,
        secondCouncilMember,
        regularMember,
        membership,
      } = await deployFixture();

      await membership
        .connect(firstCouncilMember)
        .addCouncilMember(regularMember.address);

      await expect(
        membership
          .connect(firstCouncilMember)
          .removeCouncilMember(secondCouncilMember.address)
      )
        .to.emit(membership, "ChangeInCouncilMember")
        .withArgs(secondCouncilMember.address, false);

      expect(
        await membership.isCouncilMember(secondCouncilMember.address)
      ).to.equal(false);
      expect(await membership.isCouncilMember(regularMember.address)).to.equal(
        true
      );
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

    it("cannot be called by members with one membership approval", async () => {
      const { newUser, membership, firstCouncilMember } = await deployFixture();
      await membership.connect(newUser).requestMembership();

      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);

      await expect(
        membership.connect(newUser).payMembershipFee()
      ).to.be.revertedWith("only members");
    });

    it("reverts if payment amount doesn't cover membership fee", async () => {
      const { newUser, membership, firstCouncilMember, secondCouncilMember } =
        await deployFixture();
      await addNewMember(
        newUser,
        firstCouncilMember,
        secondCouncilMember,
        membership
      );

      await expect(
        membership.connect(newUser).payMembershipFee({
          value: ethers.utils.parseUnits("3", 3),
        })
      ).to.be.revertedWith("Membership fee not covered!");
    });

    it("allows to pay membership fees", async () => {
      const {
        newUser,
        membership,
        firstCouncilMember,
        secondCouncilMember,
        wallet,
      } = await deployFixture();
      await addNewMember(
        newUser,
        firstCouncilMember,
        secondCouncilMember,
        membership
      );

      const toBePaid = ethers.utils.parseUnits("3", 4); // exactly 30k wei

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

  // this function is only callable by the owner, which is the first council member in the tests, but the timelock controller in reality
  describe("setMembershipFee", () => {
    it("allows to set the membership fee", async () => {
      const { firstCouncilMember, membership } = await deployFixture();

      await membership
        .connect(firstCouncilMember)
        .setMembershipFee(ethers.utils.parseUnits("1", 1));

      expect(await membership.membershipFee()).to.eq(
        ethers.utils.parseUnits("1", 1)
      );
    });

    it("can only be set by owner", async () => {
      const { secondCouncilMember, membership } = await deployFixture();
      const newFee = ethers.utils.parseEther("2.0");
      await expect(
        membership.connect(secondCouncilMember).setMembershipFee(newFee)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  // this function is only callable by the owner, which is the first council member in the tests, but the timelock controller in reality
  describe("setMinimumCouncilMembers", () => {
    it("allows to set the minimum council members", async () => {
      const { firstCouncilMember, membership } = await deployFixture();

      await membership.connect(firstCouncilMember).setMinimumCouncilMembers(2);

      expect(await membership.minimumCouncilMembers()).to.eq(2);
    });

    it("can not increase if there are to few council members", async () => {
      const { firstCouncilMember, membership } = await deployFixture();
      await expect(
        membership.connect(firstCouncilMember).setMinimumCouncilMembers(3)
      ).to.be.revertedWith("To few council members!");
    });

    it("can only be set by owner", async () => {
      const { secondCouncilMember, membership } = await deployFixture();
      await expect(
        membership.connect(secondCouncilMember).setMinimumCouncilMembers(27)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  // this function is only callable by the owner, which is the first council member in the tests, but the timelock controller in reality
  describe("setNewWalletAddress", () => {
    it("allows to set a new wallet address", async () => {
      const { firstCouncilMember, membership } = await deployFixture();

      const newWallet = await deployWalletContract(firstCouncilMember);

      await expect(
        membership
          .connect(firstCouncilMember)
          .setNewWalletAddress(newWallet.address)
      ).to.emit(membership, "ChangeInWalletAddress");
    });

    it("can only be set by owner", async () => {
      const { firstCouncilMember, secondCouncilMember, membership } =
        await deployFixture();
      const newWallet = await deployWalletContract(firstCouncilMember);
      await expect(
        membership
          .connect(secondCouncilMember)
          .setNewWalletAddress(newWallet.address)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  // this function is only callable by the owner, which is the first council member in the tests, but the timelock controller in reality
  describe("lockMembership", () => {
    it("allows to lock the membership", async () => {
      const { firstCouncilMember, membership } = await deployFixture();

      await membership.connect(firstCouncilMember).lockMembership();

      expect(
        await membership.connect(firstCouncilMember).membershipActive()
      ).to.eq(false);
    });

    it("other than owner can not lock the membership", async () => {
      const { secondCouncilMember, membership } = await deployFixture();

      await expect(
        membership.connect(secondCouncilMember).lockMembership()
      ).to.be.revertedWith("Ownable: caller is not the owner");
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
      expect(await membership.getMembersLength()).to.equal(4);

      await expect(membership.connect(newUser).removeMember(newUser.address))
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 0);

      expect(await membership.getMembersLength()).to.equal(3);
    });

    it("can leave association if they were approved by one member", async () => {
      const { newUser, membership, firstCouncilMember } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);

      expect(await membership.getMembersLength()).to.equal(4);
      await expect(membership.connect(newUser).removeMember(newUser.address))
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 0);

      expect(await membership.getMembersLength()).to.equal(3);
    });

    it("can leave association if they are member", async () => {
      const { newUser, firstCouncilMember, secondCouncilMember, membership } =
        await deployFixture();
      await addNewMember(
        newUser,
        firstCouncilMember,
        secondCouncilMember,
        membership
      );

      expect(await membership.getMembersLength()).to.equal(4);

      await expect(membership.connect(newUser).removeMember(newUser.address))
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 0);

      expect(await membership.getMembersLength()).to.equal(3);
    });

    it("can leave association if they are council member", async () => {
      const {
        firstCouncilMember,
        membership,
        secondCouncilMember,
        regularMember,
      } = await deployFixture();

      // council member a replacement for our second council member
      await membership
        .connect(firstCouncilMember)
        .addCouncilMember(regularMember.address);

      await expect(
        membership
          .connect(secondCouncilMember)
          .removeMember(secondCouncilMember.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(secondCouncilMember.address, 0)
        .and.to.emit(membership, "ChangeInCouncilMember")
        .withArgs(secondCouncilMember.address, false);
    });

    it("cannot remove other members if they're a normal member", async () => {
      const { regularMember, membership, firstCouncilMember } =
        await deployFixture();

      await expect(
        membership
          .connect(regularMember)
          .removeMember(firstCouncilMember.address)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });

    it("can remove other members if they are the contract owner", async () => {
      const { regularMember, membership, firstCouncilMember } =
        await deployFixture();

      await expect(
        membership
          .connect(firstCouncilMember)
          .removeMember(regularMember.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(regularMember.address, 0);
    });

    it("council members cannot leave if minimum is not met", async () => {
      const { membership, firstCouncilMember } = await deployFixture();

      await expect(
        membership
          .connect(firstCouncilMember)
          .removeMember(firstCouncilMember.address)
      ).to.be.revertedWith("Min council members not met!");
    });
  });

  describe("removeMembersThatDidntPay", () => {
    it("don't remove anyone if not necessary", async () => {
      const { membership } = await deployFixture();

      expect(await membership.getMembersLength()).to.equal(3);

      await membership.removeMembersThatDidntPay();
      expect(await membership.getMembersLength()).to.equal(3);
    });

    it("don't remove new members who haven't paid membership fees", async () => {
      const { membership } = await deployFixture();

      await membership.removeMembersThatDidntPay();
      expect(await membership.getMembersLength()).to.equal(3);
    });

    it("remove members that haven't paid", async () => {
      const { membership, firstCouncilMember, secondCouncilMember, newUser } =
        await deployFixture();

      await addNewMember(
        newUser,
        firstCouncilMember,
        secondCouncilMember,
        membership
      );
      expect(await membership.getMembersLength()).to.equal(4);

      await time.increase(365 * 24 * 60 * 60);
      await membership.removeMembersThatDidntPay();

      expect(await membership.getMembersLength()).to.equal(3);
    });
  });

  describe("getVotes", () => {
    it("everyone has same voting right after initialize", async () => {
      const { firstCouncilMember, secondCouncilMember, membership } =
        await deployFixture();

      expect(await membership.getVotes(firstCouncilMember.address)).to.equal(1);
      expect(await membership.getVotes(secondCouncilMember.address)).to.equal(
        1
      );
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

    it("member approved by one does not have voting right", async () => {
      const { firstCouncilMember, newUser, membership } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);

      expect(await membership.getVotes(newUser.address)).to.equal(0);
    });

    it("new member gets voting power", async () => {
      const { firstCouncilMember, secondCouncilMember, newUser, membership } =
        await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);
      await membership
        .connect(secondCouncilMember)
        .approveMembership(newUser.address);

      await membership.connect(newUser).payMembershipFee({
        value: ethers.utils.parseUnits("3", 4),
      });

      expect(await membership.getVotes(newUser.address)).to.equal(1);
    });
  });

  describe("getPastVotes", () => {
    it("new member never had voting power in the past", async () => {
      const { firstCouncilMember, secondCouncilMember, newUser, membership } =
        await deployFixture();

      const firstBlock = await ethers.provider.getBlockNumber();
      await membership.connect(newUser).requestMembership();

      const secondBlock = await ethers.provider.getBlockNumber();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);

      const thirdBlock = await ethers.provider.getBlockNumber();
      await membership
        .connect(secondCouncilMember)
        .approveMembership(newUser.address);

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
      const { regularMember, membership } = await deployFixture();

      const firstBlock = await ethers.provider.getBlockNumber();
      await membership
        .connect(regularMember)
        .removeMember(regularMember.address);
      const secondBlock = await ethers.provider.getBlockNumber();

      await mine(2);
      expect(
        await membership.getPastVotes(regularMember.address, firstBlock)
      ).to.equal(1);
      expect(
        await membership.getPastVotes(regularMember.address, secondBlock)
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
      const { firstCouncilMember, secondCouncilMember, newUser, membership } =
        await deployFixture();

      const firstBlock = await ethers.provider.getBlockNumber();
      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);
      await membership
        .connect(secondCouncilMember)
        .approveMembership(newUser.address);
      await membership.connect(newUser).payMembershipFee({
        value: ethers.utils.parseUnits("3", 4),
      });

      const secondBlock = await ethers.provider.getBlockNumber();
      await mine(2);
      expect(await membership.getPastTotalSupply(firstBlock)).to.equal(3);
      expect(await membership.getPastTotalSupply(secondBlock)).to.equal(4);
    });
  });

  describe("getFirstApproval", () => {
    it("returns address zero if first approval is not known", async () => {
      const { membership, newUser } = await deployFixture();

      expect(await membership.getFirstApproval(newUser.address)).to.eq(
        ethers.constants.AddressZero
      );
    });

    it("returns address of the first council member who approved", async () => {
      const { membership, firstCouncilMember, newUser } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstCouncilMember)
        .approveMembership(newUser.address);

      expect(await membership.getFirstApproval(newUser.address)).to.eq(
        firstCouncilMember.address
      );
    });
  });
});
