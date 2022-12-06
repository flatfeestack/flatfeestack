import { mine, time } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import {
  deployMembershipContract,
  deployWalletContract,
} from "./helpers/deployContracts";
import type { Contract } from "ethers";

describe("Membership", () => {
  async function deployFixture() {
    const [firstChairman, secondChairman, regularMember, newUser] =
      await ethers.getSigners();

    const { membership, wallet } = await deployMembershipContract(
      firstChairman,
      secondChairman,
      regularMember
    );

    return {
      firstChairman,
      secondChairman,
      regularMember,
      newUser,
      membership,
      wallet,
    };
  }

  async function addNewMember(
    futureMember: SignerWithAddress,
    firstChairman: SignerWithAddress,
    secondChairman: SignerWithAddress,
    membershipContract: Contract
  ) {
    await membershipContract.connect(futureMember).requestMembership();
    await membershipContract
      .connect(firstChairman)
      .approveMembership(futureMember.address);
    await membershipContract
      .connect(secondChairman)
      .approveMembership(futureMember.address);
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

  describe("isChairman", () => {
    it("normal member is not a chairman", async () => {
      const { regularMember, membership } = await deployFixture();
      expect(await membership.isChairman(regularMember.address)).to.equal(
        false
      );
    });

    it("a chairman is a chairman", async () => {
      const { firstChairman, secondChairman, membership } =
        await deployFixture();

      expect(await membership.isChairman(firstChairman.address)).to.equal(true);
      expect(await membership.isChairman(secondChairman.address)).to.equal(
        true
      );
    });
  });

  // this function is only callable by the owner, which is the first chairman in the tests, but the timelock controller in reality
  describe("addChairman", () => {
    it("can not add chairman who is already a chairman", async () => {
      const { firstChairman, secondChairman, membership } =
        await deployFixture();
      await expect(
        membership.connect(firstChairman).addChairman(secondChairman.address)
      ).to.be.revertedWith("Is already chairman!");
    });

    it("to become chairman you must be a member", async () => {
      const { firstChairman, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(firstChairman).addChairman(newUser.address)
      ).to.be.revertedWith("A chairman must be a member");
    });

    it("member can be added as chairman by owner", async () => {
      const { firstChairman, regularMember, membership } =
        await deployFixture();
      await expect(
        membership.connect(firstChairman).addChairman(regularMember.address)
      )
        .to.emit(membership, "ChangeInChairman")
        .withArgs(regularMember.address, true);
    });
  });

  describe("approveMembership", () => {
    it("non member can not be approved", async () => {
      const { secondChairman, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(secondChairman).approveMembership(newUser.address)
      ).to.be.revertedWith("Invalid member status!");
    });

    it("requesting member gets approved by one chairman", async () => {
      const { firstChairman, newUser, membership } = await deployFixture();

      await membership.connect(newUser).requestMembership();

      await expect(
        membership.connect(firstChairman).approveMembership(newUser.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 2);
    });

    it("requesting member gets approved by second chairman", async () => {
      const { firstChairman, secondChairman, newUser, membership } =
        await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);

      await expect(
        membership.connect(secondChairman).approveMembership(newUser.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 3);

      expect(await membership.nextMembershipFeePayment(newUser.address)).to.eq(
        0
      );
    });

    it("requesting member can not get approved by same chairman", async () => {
      const { firstChairman, newUser, membership } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);

      await expect(
        membership.connect(firstChairman).approveMembership(newUser.address)
      ).to.be.revertedWith("Invalid member status!");
    });
  });

  describe("removeChairman", () => {
    it("the to be removed address must be a chairman", async () => {
      const { firstChairman, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(firstChairman).removeChairman(newUser.address)
      ).to.be.revertedWith("Is no chairman!");
    });

    it("can not remove chairman if number of chairmen becomes less than minimum amount", async () => {
      const { firstChairman, secondChairman, membership } =
        await deployFixture();

      await expect(
        membership.connect(firstChairman).removeChairman(secondChairman.address)
      ).to.be.revertedWith("Minimum chairmen not met!");
    });

    it("chairmen can be removed by contract owner", async () => {
      const { firstChairman, secondChairman, regularMember, membership } =
        await deployFixture();

      await membership
        .connect(firstChairman)
        .addChairman(regularMember.address);

      await expect(
        membership.connect(firstChairman).removeChairman(secondChairman.address)
      )
        .to.emit(membership, "ChangeInChairman")
        .withArgs(secondChairman.address, false);

      expect(await membership.isChairman(secondChairman.address)).to.equal(
        false
      );
      expect(await membership.isChairman(regularMember.address)).to.equal(true);
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
      const { newUser, membership, firstChairman } = await deployFixture();
      await membership.connect(newUser).requestMembership();

      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);

      await expect(
        membership.connect(newUser).payMembershipFee()
      ).to.be.revertedWith("only members");
    });

    it("reverts if payment amount doesn't cover membership fee", async () => {
      const { newUser, membership, firstChairman, secondChairman } =
        await deployFixture();
      await addNewMember(newUser, firstChairman, secondChairman, membership);

      await expect(
        membership.connect(newUser).payMembershipFee({
          value: ethers.utils.parseUnits("3", 3),
        })
      ).to.be.revertedWith("Membership fee not covered!");
    });

    it("allows to pay membership fees", async () => {
      const { newUser, membership, firstChairman, secondChairman, wallet } =
        await deployFixture();
      await addNewMember(newUser, firstChairman, secondChairman, membership);

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

  // this function is only callable by the owner, which is the first chairman in the tests, but the timelock controller in reality
  describe("setMembershipFee", () => {
    it("allows to set the membership fee", async () => {
      const { firstChairman, membership } = await deployFixture();

      await membership
        .connect(firstChairman)
        .setMembershipFee(ethers.utils.parseUnits("1", 1));

      expect(await membership.membershipFee()).to.eq(
        ethers.utils.parseUnits("1", 1)
      );
    });

    it("can only be set by owner", async () => {
      const { secondChairman, membership } = await deployFixture();
      const newFee = ethers.utils.parseEther("2.0");
      await expect(
        membership.connect(secondChairman).setMembershipFee(newFee)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  // this function is only callable by the owner, which is the first chairman in the tests, but the timelock controller in reality
  describe("setMinimumChairmen", () => {
    it("allows to set the minimum chairmen", async () => {
      const { firstChairman, membership } = await deployFixture();

      await membership.connect(firstChairman).setMinimumChairmen(2);

      expect(await membership.minimumChairmen()).to.eq(2);
    });

    it("can not increase if there are to few chairmen", async () => {
      const { firstChairman, membership } = await deployFixture();
      await expect(
        membership.connect(firstChairman).setMinimumChairmen(3)
      ).to.be.revertedWith("To few chairmen!");
    });

    it("can only be set by owner", async () => {
      const { secondChairman, membership } = await deployFixture();
      await expect(
        membership.connect(secondChairman).setMinimumChairmen(27)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  // this function is only callable by the owner, which is the first chairman in the tests, but the timelock controller in reality
  describe("setNewWalletAddress", () => {
    it("allows to set a new wallet address", async () => {
      const { firstChairman, membership } = await deployFixture();

      const newWallet = await deployWalletContract(firstChairman);

      await expect(
        membership.connect(firstChairman).setNewWalletAddress(newWallet.address)
      ).to.emit(membership, "ChangeInWalletAddress");
    });

    it("can only be set by owner", async () => {
      const { firstChairman, secondChairman, membership } =
        await deployFixture();
      const newWallet = await deployWalletContract(firstChairman);
      await expect(
        membership
          .connect(secondChairman)
          .setNewWalletAddress(newWallet.address)
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
      const { newUser, membership, firstChairman } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);

      expect(await membership.getMembersLength()).to.equal(4);
      await expect(membership.connect(newUser).removeMember(newUser.address))
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 0);

      expect(await membership.getMembersLength()).to.equal(3);
    });

    it("can leave association if they are member", async () => {
      const { newUser, firstChairman, secondChairman, membership } =
        await deployFixture();
      await addNewMember(newUser, firstChairman, secondChairman, membership);

      expect(await membership.getMembersLength()).to.equal(4);

      await expect(membership.connect(newUser).removeMember(newUser.address))
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(newUser.address, 0);

      expect(await membership.getMembersLength()).to.equal(3);
    });

    it("can leave association if they are chairman", async () => {
      const { firstChairman, membership, secondChairman, regularMember } =
        await deployFixture();

      // chairman a replacement for our second chairman
      await membership
        .connect(firstChairman)
        .addChairman(regularMember.address);

      await expect(
        membership.connect(secondChairman).removeMember(secondChairman.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(secondChairman.address, 0)
        .and.to.emit(membership, "ChangeInChairman")
        .withArgs(secondChairman.address, false);
    });

    it("cannot remove other members if they're a normal member", async () => {
      const { regularMember, membership, firstChairman } =
        await deployFixture();

      await expect(
        membership.connect(regularMember).removeMember(firstChairman.address)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });

    it("can remove other members if they are the contract owner", async () => {
      const { regularMember, membership, firstChairman } =
        await deployFixture();

      await expect(
        membership.connect(firstChairman).removeMember(regularMember.address)
      )
        .to.emit(membership, "ChangeInMembershipStatus")
        .withArgs(regularMember.address, 0);
    });

    it("chairmen cannot leave if minimum is not met", async () => {
      const { membership, firstChairman } = await deployFixture();

      await expect(
        membership.connect(firstChairman).removeMember(firstChairman.address)
      ).to.be.revertedWith("Minimum chairmen not met!");
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
      const { membership, firstChairman, secondChairman, newUser } =
        await deployFixture();

      await addNewMember(newUser, firstChairman, secondChairman, membership);
      expect(await membership.getMembersLength()).to.equal(4);

      await time.increase(365 * 24 * 60 * 60);
      await membership.removeMembersThatDidntPay();

      expect(await membership.getMembersLength()).to.equal(3);
    });
  });

  describe("getVotes", () => {
    it("everyone has same voting right after initialize", async () => {
      const { firstChairman, secondChairman, membership } =
        await deployFixture();

      expect(await membership.getVotes(firstChairman.address)).to.equal(1);
      expect(await membership.getVotes(secondChairman.address)).to.equal(1);
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
      const { firstChairman, newUser, membership } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);

      expect(await membership.getVotes(newUser.address)).to.equal(0);
    });

    it("new member gets voting power", async () => {
      const { firstChairman, secondChairman, newUser, membership } =
        await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);
      await membership
        .connect(secondChairman)
        .approveMembership(newUser.address);

      await membership.connect(newUser).payMembershipFee({
        value: ethers.utils.parseUnits("3", 4),
      });

      expect(await membership.getVotes(newUser.address)).to.equal(1);
    });
  });

  describe("getPastVotes", () => {
    it("new member never had voting power in the past", async () => {
      const { firstChairman, secondChairman, newUser, membership } =
        await deployFixture();

      const firstBlock = await ethers.provider.getBlockNumber();
      await membership.connect(newUser).requestMembership();

      const secondBlock = await ethers.provider.getBlockNumber();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);

      const thirdBlock = await ethers.provider.getBlockNumber();
      await membership
        .connect(secondChairman)
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
      const { firstChairman, secondChairman, newUser, membership } =
        await deployFixture();

      const firstBlock = await ethers.provider.getBlockNumber();
      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);
      await membership
        .connect(secondChairman)
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

    it("returns address of the first chairman who approved", async () => {
      const { membership, firstChairman, newUser } = await deployFixture();

      await membership.connect(newUser).requestMembership();
      await membership
        .connect(firstChairman)
        .approveMembership(newUser.address);

      expect(await membership.getFirstApproval(newUser.address)).to.eq(
        firstChairman.address
      );
    });
  });
});
