import { ethers, upgrades } from "hardhat";
import { expect } from "chai";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";

export async function deployWalletContract(owner: SignerWithAddress) {
  const Wallet = await ethers.getContractFactory("Wallet", { signer: owner });
  const wallet = await upgrades.deployProxy(Wallet, []);
  await wallet.deployed();

  return wallet;
}

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
      const { otherAccount, wallet } = await deployFixture();

      await expect(
        wallet.increaseAllowance(
          otherAccount.address,
          ethers.utils.parseEther("200.0")
        )
      ).to.be.revertedWith("Keep allowance below balance!");
    });

    it("increases allowance for given address", async () => {
      const { otherAccount, wallet } = await deployFixture();

      await wallet.increaseAllowance(
        otherAccount.address,
        ethers.utils.parseEther("0.5")
      );
      expect(await wallet.allowance(otherAccount.address)).to.eq(
        ethers.utils.parseEther("0.5")
      );
    });
  });

  describe("payContribution", () => {
    it("increases total balance and contribution - no previous contributions", async () => {
      const { otherAccount, wallet } = await deployFixture();
      const currentBalance = await wallet.totalBalance();

      await wallet.payContribution(otherAccount.address, {
        value: ethers.utils.parseEther("0.5"),
      });

      expect(await wallet.totalBalance()).to.eq(
        ethers.utils.parseEther("0.5").add(currentBalance)
      );
      expect(await wallet.individualContribution(otherAccount.address)).to.eq(
        ethers.utils.parseEther("0.5")
      );
    });

    it("increases total balance and contribution - previous contributions", async () => {
      const { owner, wallet } = await deployFixture();
      const currentBalance = await wallet.totalBalance();
      const contribution = ethers.utils.parseEther("0.5");

      await wallet.payContribution(owner.address, {
        value: contribution,
      });

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
      ).to.be.revertedWith("no allowance for this account!");
    });

    it("can withdraw allowance", async () => {
      const { otherAccount, wallet } = await deployFixture();
      const withdrawAmount = ethers.utils.parseEther("0.5");

      await wallet.increaseAllowance(otherAccount.address, withdrawAmount);

      await expect(
        wallet.withdrawMoney(otherAccount.address)
      ).to.changeEtherBalances(
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
});
