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

  describe("request membership", () => {
    it("request membership emits event", async () => {
      const { newUser, membership } = await deployFixture();
      expect(await membership.connect(newUser).requestMembership())
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
});
