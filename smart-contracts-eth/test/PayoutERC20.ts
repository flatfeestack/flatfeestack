import { ethers, upgrades } from "hardhat";
import { expect } from "chai";
import generateSignature from "./helpers/generateSignature";

describe("PayoutERC20", () => {
  async function deployFixture() {
    const [ffsTokenOwner, payoutOwner, developer] = await ethers.getSigners();

    const FFSToken = await ethers.getContractFactory("FlatFeeStackToken", {
      signer: ffsTokenOwner,
    });
    const ffsToken = await upgrades.deployProxy(FFSToken, []);
    await ffsToken.deployed();

    const PayoutERC20 = await ethers.getContractFactory("PayoutERC20", {
      signer: payoutOwner,
    });
    const payoutERC20 = await upgrades.deployProxy(PayoutERC20, [
      ffsToken.address,
      "FFST",
    ]);
    await payoutERC20.deployed();

    await ffsToken.connect(ffsTokenOwner).transfer(payoutERC20.address, 100);

    return { payoutOwner, ffsToken, payoutERC20, developer };
  }

  describe("withdraw", () => {
    it("throws error when requesting 0 amount", async () => {
      const { payoutERC20, developer } = await deployFixture();

      await expect(
        payoutERC20
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
      const { payoutERC20, developer } = await deployFixture();

      await expect(
        payoutERC20
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
      const { payoutOwner, payoutERC20, ffsToken, developer } =
        await deployFixture();

      const userId = ethers.utils.formatBytes32String("someUserId");
      const amount = 10;
      const signature = await generateSignature(
        amount,
        payoutOwner,
        userId,
        "FFST"
      );

      const previousBalanceContract = await ffsToken.balanceOf(
        payoutERC20.address
      );
      const previousBalanceDeveloper = await ffsToken.balanceOf(
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

      expect(await ffsToken.balanceOf(payoutERC20.address)).to.eq(
        previousBalanceContract.sub(amount)
      );
      expect(await ffsToken.balanceOf(developer.address)).to.eq(
        previousBalanceDeveloper.add(amount)
      );

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
    });

    it("should only payoutEth difference", async () => {
      const { payoutOwner, payoutERC20, ffsToken, developer } =
        await deployFixture();

      const userId = ethers.utils.formatBytes32String("someUserId");
      const firstWithdraw = 10;
      const firstSignature = await generateSignature(
        firstWithdraw,
        payoutOwner,
        userId,
        "FFST"
      );

      const previousBalanceContract = await ffsToken.balanceOf(
        payoutERC20.address
      );
      const previousBalanceDeveloper = await ffsToken.balanceOf(
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

      expect(await ffsToken.balanceOf(payoutERC20.address)).to.eq(
        previousBalanceContract.sub(firstWithdraw)
      );
      expect(await ffsToken.balanceOf(developer.address)).to.eq(
        previousBalanceDeveloper.add(firstWithdraw)
      );

      const secondWithdraw = 20;
      const tea = secondWithdraw + firstWithdraw;
      const secondSignature = await generateSignature(
        tea,
        payoutOwner,
        userId,
        "FFST"
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

      expect(await ffsToken.balanceOf(payoutERC20.address)).to.eq(
        previousBalanceContract.sub(tea)
      );
      expect(await ffsToken.balanceOf(developer.address)).to.eq(
        previousBalanceDeveloper.add(tea)
      );
    });
  });
});
