import { ethers, upgrades } from "hardhat";
import { expect } from "chai";
import generateSignature from "./helpers/generateSignature";

describe("PayoutERC20", () => {
  async function deployFixture() {
    const [usdcTokenOwner, payoutOwner, developer] = await ethers.getSigners();

    const USDCToken = await ethers.getContractFactory("USDC", {
      signer: usdcTokenOwner,
    });
    const usdcToken = await upgrades.deployProxy(USDCToken, []);
    await usdcToken.deployed();

    const PayoutERC20 = await ethers.getContractFactory("PayoutERC20", {
      signer: payoutOwner,
    });
    const payoutERC20 = await upgrades.deployProxy(PayoutERC20, [
      usdcToken.address,
      "USDC",
    ]);
    await payoutERC20.deployed();

    await usdcToken.connect(usdcTokenOwner).transfer(payoutERC20.address, 100);

    return { payoutOwner, usdcToken, payoutERC20, developer };
  }

  describe("withdraw", () => {
    it("throws error when requesting 0 amount", async () => {
      const { payoutERC20, developer } = await deployFixture();

      await expect(
        payoutERC20
          .connect(developer)
          .withdraw(
            developer.address,
            "someUserId",
            0,
            0,
            ethers.utils.formatBytes32String("test"),
            ethers.utils.formatBytes32String("test")
          )
      ).to.be.revertedWith("No new funds to be withdrawn");
    });

    it("throws error when requesting with invalid signature", async () => {
      const { payoutERC20, developer } = await deployFixture();

      await expect(
        payoutERC20
          .connect(developer)
          .withdraw(
            developer.address,
            "someUserId",
            100,
            0,
            ethers.utils.formatBytes32String("test"),
            ethers.utils.formatBytes32String("test")
          )
      ).to.be.revertedWith("Signature no match");
    });

    it("should retrieve funds with correct signature", async () => {
      const { payoutOwner, payoutERC20, usdcToken, developer } =
        await deployFixture();

      const userId = "4fed2b83-f968-45cc-8869-a36f844cefdb";
      const amount = 10;
      const signature = await generateSignature(
        amount,
        payoutOwner,
        userId,
        "USDC"
      );

      const previousBalanceContract = await usdcToken.balanceOf(
        payoutERC20.address
      );
      const previousBalanceDeveloper = await usdcToken.balanceOf(
        developer.address
      );

      await payoutERC20
        .connect(developer)
        .withdraw(
          developer.address,
          userId,
          amount,
          signature.v,
          signature.r,
          signature.s
        );

      expect(await usdcToken.balanceOf(payoutERC20.address)).to.eq(
        previousBalanceContract.sub(amount)
      );
      expect(await usdcToken.balanceOf(developer.address)).to.eq(
        previousBalanceDeveloper.add(amount)
      );
      expect(await payoutERC20.getPayedOut(userId)).to.eq(amount);

      // also check that using the signature a second time does not work
      await expect(
        payoutERC20
          .connect(developer)
          .withdraw(
            developer.address,
            userId,
            amount,
            signature.v,
            signature.r,
            signature.s
          )
      ).to.be.revertedWith("No new funds to be withdrawn");
      expect(await payoutERC20.getPayedOut(userId)).to.eq(amount);
    });

    it("should only payoutEth difference", async () => {
      const { payoutOwner, payoutERC20, usdcToken, developer } =
        await deployFixture();

      const userId = "4fed2b83-f968-45cc-8869-a36f844cefdb";
      const firstWithdraw = 10;
      const firstSignature = await generateSignature(
        firstWithdraw,
        payoutOwner,
        userId,
        "USDC"
      );

      const previousBalanceContract = await usdcToken.balanceOf(
        payoutERC20.address
      );
      const previousBalanceDeveloper = await usdcToken.balanceOf(
        developer.address
      );

      await payoutERC20
        .connect(developer)
        .withdraw(
          developer.address,
          userId,
          firstWithdraw,
          firstSignature.v,
          firstSignature.r,
          firstSignature.s
        );

      expect(await usdcToken.balanceOf(payoutERC20.address)).to.eq(
        previousBalanceContract.sub(firstWithdraw)
      );
      expect(await usdcToken.balanceOf(developer.address)).to.eq(
        previousBalanceDeveloper.add(firstWithdraw)
      );
      expect(await payoutERC20.getPayedOut(userId)).to.eq(firstWithdraw);

      const secondWithdraw = 20;
      const tea = secondWithdraw + firstWithdraw;
      const secondSignature = await generateSignature(
        tea,
        payoutOwner,
        userId,
        "USDC"
      );

      await payoutERC20
        .connect(developer)
        .withdraw(
          developer.address,
          userId,
          tea,
          secondSignature.v,
          secondSignature.r,
          secondSignature.s
        );

      expect(await usdcToken.balanceOf(payoutERC20.address)).to.eq(
        previousBalanceContract.sub(tea)
      );
      expect(await usdcToken.balanceOf(developer.address)).to.eq(
        previousBalanceDeveloper.add(tea)
      );
      expect(await payoutERC20.getPayedOut(userId)).to.eq(tea);
    });
  });
});
