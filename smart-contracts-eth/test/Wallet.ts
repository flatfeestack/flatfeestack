import { expect } from "chai";
import { ethers } from "hardhat";
import { deployWalletContract } from "./helpers/deployContracts";

describe("Wallet", () => {
  async function deployFixture() {
    // Contracts are deployed using the first signer/account by default
    const [owner, otherAccount] = await ethers.getSigners();
    const wallet = await deployWalletContract(owner);

    await owner.sendTransaction({
      to: wallet.address,
      value: ethers.utils.parseEther("1.0"),
    });

    return { otherAccount, owner, wallet };
  }

  describe("increaseAllowance", () => {
    it("cannot increase more than total balance of wallet", async () => {
      const { owner, otherAccount, wallet } = await deployFixture();

      await expect(
        wallet
          .connect(owner)
          .increaseAllowance(
            otherAccount.address,
            ethers.utils.parseEther("200.0")
          )
      ).to.be.revertedWith("Keep allowance below balance!");
    });

    it("increases allowance for given address", async () => {
      const { owner, otherAccount, wallet } = await deployFixture();

      const halfAnEther = ethers.utils.parseEther("0.5");

      await expect(
        wallet
          .connect(owner)
          .increaseAllowance(otherAccount.address, halfAnEther)
      )
        .to.emit(wallet, "IncreaseAllowance")
        .withArgs(otherAccount.address, halfAnEther);
      expect(await wallet.allowance(otherAccount.address)).to.eq(halfAnEther);
    });

    it("other than owner can not increase allowance", async () => {
      const { otherAccount, wallet } = await deployFixture();

      await expect(
        wallet
          .connect(otherAccount)
          .increaseAllowance(
            otherAccount.address,
            ethers.utils.parseEther("0.5")
          )
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  describe("payContribution", () => {
    it("increases total balance and contribution - no previous contributions", async () => {
      const { otherAccount, wallet } = await deployFixture();
      const currentBalance = await wallet.totalBalance();
      const halfAnEther = ethers.utils.parseEther("0.5");

      await expect(
        wallet.payContribution(otherAccount.address, {
          value: halfAnEther,
        })
      )
        .to.emit(wallet, "AcceptPayment")
        .withArgs(otherAccount.address, halfAnEther);

      expect(await wallet.individualContribution(otherAccount.address)).to.eq(
        halfAnEther
      );
      expect(await wallet.totalBalance()).to.eq(
        halfAnEther.add(currentBalance)
      );
    });

    it("increases total balance and contribution - previous contributions", async () => {
      const { owner, wallet } = await deployFixture();
      const currentBalance = await wallet.totalBalance();
      const contribution = ethers.utils.parseEther("0.5");

      await expect(
        wallet.payContribution(owner.address, {
          value: contribution,
        })
      )
        .to.emit(wallet, "AcceptPayment")
        .withArgs(owner.address, contribution);

      const expectedContribution = contribution.add(currentBalance);
      expect(await wallet.totalBalance()).to.eq(expectedContribution);
      expect(await wallet.individualContribution(owner.address)).to.eq(
        expectedContribution
      );
    });
  });

  describe("withdrawMoney", () => {
    it("cannot withdraw without allowance", async () => {
      const { otherAccount, wallet } = await deployFixture();

      await expect(
        wallet.withdrawMoney(otherAccount.address)
      ).to.be.revertedWith("No allowance for this account!");
    });

    it("can withdraw allowance", async () => {
      const { owner, otherAccount, wallet } = await deployFixture();
      const withdrawAmount = ethers.utils.parseEther("0.5");

      await wallet
        .connect(owner)
        .increaseAllowance(otherAccount.address, withdrawAmount);

      await expect(wallet.withdrawMoney(otherAccount.address))
        .to.emit(wallet, "WithdrawFunds")
        .withArgs(owner.address, withdrawAmount)
        .and.to.changeEtherBalances(
          [otherAccount, wallet],
          [withdrawAmount, withdrawAmount.mul(BigInt(-1))]
        );
    });
  });

  describe("knownSender", () => {
    it("should add and remove known sender", async () => {
      const { owner, otherAccount, wallet } = await deployFixture();
      expect(await wallet.isKnownSender(owner.address)).to.be.true;
      expect(await wallet.isKnownSender(otherAccount.address)).to.be.false;

      await wallet.addKnownSender(otherAccount.address);
      expect(await wallet.isKnownSender(otherAccount.address)).to.be.true;

      await wallet.removeKnownSender(otherAccount.address);
      expect(await wallet.isKnownSender(otherAccount.address)).to.be.false;
    });

    it("should not allow to remove owner from known senders", async () => {
      const { owner, wallet } = await deployFixture();
      await expect(wallet.removeKnownSender(owner.address)).to.be.revertedWith(
        "Owner cannot be removed!"
      );
    });
  });

  describe("liquidate", () => {
    it("can liquidate wallet", async () => {
      const { owner, otherAccount, wallet } = await deployFixture();
      const totalBalance = await wallet.totalBalance();

      await expect(wallet.connect(owner).liquidate(otherAccount.address))
        .to.emit(wallet, "WithdrawFunds")
        .withArgs(otherAccount.address, totalBalance);
    });

    it("other than owner can not liquidate", async () => {
      const { otherAccount, wallet } = await deployFixture();

      await expect(
        wallet.connect(otherAccount).liquidate(otherAccount.address)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  describe("receive", () => {
    it("emits event when receiving funds", async () => {
      const { owner, wallet } = await deployFixture();

      const currentTotalBalance = await wallet.totalBalance();
      const oneEther = ethers.utils.parseEther("1.0");

      await expect(
        owner.sendTransaction({
          to: wallet.address,
          value: oneEther,
        })
      )
        .to.emit(wallet, "AcceptPayment")
        .withArgs(owner.address, oneEther);

      expect(await wallet.totalBalance()).to.eq(
        currentTotalBalance.add(oneEther)
      );
    });
  });
});
