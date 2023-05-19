import { ethers, upgrades } from "hardhat";
import { expect } from "chai";
import generateSignature from "./helpers/generateSignature";

describe("PayoutEth", () => {
  async function deployFixture() {
    const [owner, developer] = await ethers.getSigners();
    const PayoutEth = await ethers.getContractFactory("PayoutEth", {
      signer: owner,
    });

    const payoutEth = await upgrades.deployProxy(PayoutEth, []);
    await payoutEth.deployed();

    await owner.sendTransaction({
      to: payoutEth.address,
      value: ethers.utils.parseEther("1.0"),
    });

    return { owner, payoutEth: payoutEth, developer };
  }

  describe("getContractBalance", () => {
    it("returns current contract balance", async () => {
      const { payoutEth, owner } = await deployFixture();
      expect(await payoutEth.connect(owner).getContractBalance()).to.eq(
        ethers.utils.parseEther("1.0")
      );
    });
  });

  describe("withdraw", () => {
    it("throws error when requesting 0 amount", async () => {
      const { payoutEth, developer } = await deployFixture();

      await expect(
        payoutEth
          .connect(developer)
          .withdraw(
            developer.address,
            ethers.utils.hashMessage("someUserId"),
            0,
            0,
            ethers.utils.formatBytes32String("test"),
            ethers.utils.formatBytes32String("test")
          )
      ).to.be.revertedWith("No new funds to be withdrawn");
    });

    it("throws error when requesting with invalid signature", async () => {
      const { payoutEth, developer } = await deployFixture();

      await expect(
        payoutEth
          .connect(developer)
          .withdraw(
            developer.address,
            ethers.utils.hashMessage("someUserId"),
            100,
            0,
            ethers.utils.formatBytes32String("test"),
            ethers.utils.formatBytes32String("test")
          )
      ).to.be.revertedWith("Signature no match");
    });

    it("should retrieve funds with correct signature", async () => {
      const { owner, payoutEth, developer } = await deployFixture();

      const userId = "4fed2b83-f968-45cc-8869-a36f844cefdb";
      const amount = 10000;
      const { encodedUserId, signature } = await generateSignature(
        amount,
        owner,
        userId,
        "ETH"
      );

      await expect(
        payoutEth
          .connect(developer)
          .withdraw(
            developer.address,
            encodedUserId,
            amount,
            signature.v,
            signature.r,
            signature.s
          )
      ).to.changeEtherBalances([developer, payoutEth], [amount, amount * -1]);
      expect(await payoutEth.getPayedOut(encodedUserId)).to.eq(amount);

      // also check that using the signature a second time does not work
      await expect(
        payoutEth
          .connect(developer)
          .withdraw(
            developer.address,
            encodedUserId,
            amount,
            signature.v,
            signature.r,
            signature.s
          )
      ).to.be.revertedWith("No new funds to be withdrawn");
      expect(await payoutEth.getPayedOut(encodedUserId)).to.eq(amount);
    });

    it("should only payoutEth difference", async () => {
      const { owner, payoutEth, developer } = await deployFixture();

      const userId = "4fed2b83-f968-45cc-8869-a36f844cefdb";
      const firstWithdraw = 10000;
      const { encodedUserId, signature: firstSignature } =
        await generateSignature(firstWithdraw, owner, userId, "ETH");

      await expect(
        payoutEth
          .connect(developer)
          .withdraw(
            developer.address,
            encodedUserId,
            firstWithdraw,
            firstSignature.v,
            firstSignature.r,
            firstSignature.s
          )
      ).to.changeEtherBalances(
        [developer, payoutEth],
        [firstWithdraw, firstWithdraw * -1]
      );
      expect(await payoutEth.getPayedOut(encodedUserId)).to.eq(firstWithdraw);

      const secondWithdraw = 20000;
      const tea = secondWithdraw + firstWithdraw;
      const { signature: secondSignature } = await generateSignature(
        tea,
        owner,
        userId,
        "ETH"
      );

      await expect(
        payoutEth
          .connect(developer)
          .withdraw(
            developer.address,
            encodedUserId,
            tea,
            secondSignature.v,
            secondSignature.r,
            secondSignature.s
          )
      ).to.changeEtherBalances(
        [developer, payoutEth],
        [secondWithdraw, secondWithdraw * -1]
      );
      expect(await payoutEth.getPayedOut(encodedUserId)).to.eq(tea);
    });
  });
});
