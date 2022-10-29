import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers, upgrades } from "hardhat";

export async function deployWalletContract(owner: SignerWithAddress) {
  const Wallet = await ethers.getContractFactory("Wallet", { signer: owner });
  const wallet = await upgrades.deployProxy(Wallet, []);
  await wallet.deployed();

  return wallet;
}

export async function deployMembershipContract(
  representative: SignerWithAddress,
  whitelisterOne: SignerWithAddress,
  whitelisterTwo: SignerWithAddress
) {
  const wallet = await deployWalletContract(representative);
  const Membership = await ethers.getContractFactory("Membership");
  const membership = await upgrades.deployProxy(Membership, [
    representative.address,
    whitelisterOne.address,
    whitelisterTwo.address,
    wallet.address,
  ]);

  await membership.deployed();
  await wallet.addKnownSender(membership.address);

  return { membership: membership, wallet: wallet };
}
