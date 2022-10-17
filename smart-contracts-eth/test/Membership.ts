import { ethers, upgrades } from "hardhat";
import { expect } from "chai";

describe("Membership", () => {
  async function deployFixture() {
    const [delegate, whitelisterOne, whitelisterTwo, newUser] =
      await ethers.getSigners();
    const Membership = await ethers.getContractFactory("Membership");
    const membership = await upgrades.deployProxy(Membership, [
      delegate.address,
      whitelisterOne.address,
      whitelisterTwo.address,
    ]);
    await membership.deployed();

    return { delegate, whitelisterOne, whitelisterTwo, newUser, membership };
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
        "This function can only be called by non-members"
      );
    });
  });

  describe("isWhitelister", () => {
    it("delegate is not whitelister", async () => {
      const { delegate, membership } = await deployFixture();
      expect(await membership.isWhitelister(delegate.address)).to.equal(false);
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
      const { delegate, whitelisterOne, membership } = await deployFixture();
      await expect(
        membership.connect(delegate).addWhitelister(whitelisterOne.address)
      ).to.be.revertedWith("This address is already a whitelister");
    });

    it("to become whitelister you must be a member", async () => {
      const { delegate, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(delegate).addWhitelister(newUser.address)
      ).to.be.revertedWith("A whitelister must be a member");
    });

    it("a delegate can't become a whitelister", async () => {
      const { delegate, membership } = await deployFixture();
      await expect(
        membership.connect(delegate).addWhitelister(delegate.address)
      ).to.be.revertedWith("The delegate can't become a whitelister");
    });

    it("member can be added as whitelister by delegate", async () => {
      const { delegate, whitelisterOne, whitelisterTwo, newUser, membership } =
        await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await membership.connect(whitelisterTwo).whitelistMember(newUser.address);
      await expect(membership.connect(delegate).addWhitelister(newUser.address))
        .to.emit(membership, "ChangeInWhiteLister")
        .withArgs(newUser.address, true);
    });
  });

  describe("whitelistMember", () => {
    it("non member can not be whitelisted", async () => {
      const { whitelisterOne, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(whitelisterOne).whitelistMember(newUser.address)
      ).to.be.revertedWithoutReason();
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
    });

    it("requesting member can not get whitelisted by same whitelister", async () => {
      const { whitelisterOne, newUser, membership } = await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await expect(
        membership.connect(whitelisterOne).whitelistMember(newUser.address)
      ).to.be.revertedWithoutReason();
    });
  });

  describe("removeWhitelister", () => {
    it("the to be removed address must be a whitelister", async () => {
      const { delegate, newUser, membership } = await deployFixture();
      await expect(
        membership.connect(delegate).removeWhitelister(newUser.address)
      ).to.be.revertedWith("This address is not a whitelister");
    });

    it("can not remove whitelister if number of whitelisters becomes less than minimum number of whitelisters", async () => {
      const { delegate, whitelisterOne, membership } = await deployFixture();
      await expect(
        membership.connect(delegate).removeWhitelister(whitelisterOne.address)
      ).to.be.revertedWith(
        "Can't remove because there is a minimum of 2 whitelisters"
      );
    });

    it("whitelister can be removed by delegate", async () => {
      const { delegate, whitelisterOne, whitelisterTwo, newUser, membership } =
        await deployFixture();
      await membership.connect(newUser).requestMembership();
      await membership.connect(whitelisterOne).whitelistMember(newUser.address);
      await membership.connect(whitelisterTwo).whitelistMember(newUser.address);
      await membership.connect(delegate).addWhitelister(newUser.address);
      await expect(
        membership.connect(delegate).removeWhitelister(whitelisterOne.address)
      )
        .to.emit(membership, "ChangeInWhiteLister")
        .withArgs(whitelisterOne.address, false);

      expect(await membership.isWhitelister(whitelisterOne.address)).to.equal(
        false
      );
      expect(await membership.isWhitelister(newUser.address)).to.equal(true);
    });
  });
});
