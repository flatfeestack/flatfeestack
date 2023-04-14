import { ethers, upgrades } from "hardhat";
import type { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import type { Signature } from "@ethersproject/bytes";
import { expect } from "chai";

describe("Payout", () => {
  async function deployFixture() {
    const [owner, developer] = await ethers.getSigners();
    const Payout = await ethers.getContractFactory("Payout", { signer: owner });
    const payout = await upgrades.deployProxy(Payout, []);
    await payout.deployed();

    await owner.sendTransaction({
      to: payout.address,
      value: ethers.utils.parseEther("1.0"),
    });

    return { owner, payout, developer };
  }

  async function generateSignature(
    amount: Number,
    owner: SignerWithAddress,
    userId: string
  ): Promise<Signature> {
    const payload = ethers.utils.defaultAbiCoder.encode(
      ["bytes32", "string", "uint256"],
      [userId, "#", amount]
    );
    const payloadHash = ethers.utils.keccak256(payload);

    const signature = await owner.signMessage(
      ethers.utils.arrayify(payloadHash)
    );
    return ethers.utils.splitSignature(signature);
  }

  describe("withdraw", () => {
    it("throws error when requesting 0 amount", async () => {
      const { payout, developer } = await deployFixture();

      await expect(
        payout
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
      const { payout, developer } = await deployFixture();

      await expect(
        payout
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
      const { owner, payout, developer } = await deployFixture();

      const userId = ethers.utils.formatBytes32String("someUserId");
      const amount = 10000;
      const signature = await generateSignature(amount, owner, userId);

      await expect(
        payout
          .connect(developer)
          .withdraw(
            developer.address,
            userId,
            amount,
            signature.v,
            signature.r,
            signature.s
          )
      ).to.changeEtherBalances([developer, payout], [amount, amount * -1]);

      // also check that using the signature a second time does not work
      await expect(
        payout
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

    it("should only payout difference", async () => {
      const { owner, payout, developer } = await deployFixture();

      const userId = ethers.utils.formatBytes32String("someUserId");
      const firstWithdraw = 10000;
      const firstSignature = await generateSignature(
        firstWithdraw,
        owner,
        userId
      );

      await expect(
        payout
          .connect(developer)
          .withdraw(
            developer.address,
            userId,
            firstWithdraw,
            firstSignature.v,
            firstSignature.r,
            firstSignature.s
          )
      ).to.changeEtherBalances(
        [developer, payout],
        [firstWithdraw, firstWithdraw * -1]
      );

      const secondWithdraw = 20000;
      const tea = secondWithdraw + firstWithdraw;
      const secondSignature = await generateSignature(tea, owner, userId);

      await expect(
        payout
          .connect(developer)
          .withdraw(
            developer.address,
            userId,
            tea,
            secondSignature.v,
            secondSignature.r,
            secondSignature.s
          )
      ).to.changeEtherBalances(
        [developer, payout],
        [secondWithdraw, secondWithdraw * -1]
      );
    });
  });
});
