import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers, upgrades } from "hardhat";
import { mine } from "@nomicfoundation/hardhat-network-helpers";

export async function deployWalletContract(owner: SignerWithAddress) {
  const Wallet = await ethers.getContractFactory("Wallet", { signer: owner });
  const wallet = await upgrades.deployProxy(Wallet, []);
  await wallet.deployed();

  return wallet;
}

export async function deployMembershipContract(
  firstChairman: SignerWithAddress,
  secondChairman: SignerWithAddress,
  regularMember: SignerWithAddress
) {
  const wallet = await deployWalletContract(firstChairman);
  const Membership = await ethers.getContractFactory("Membership", {
    signer: firstChairman,
  });
  const membership = await upgrades.deployProxy(Membership, [
    firstChairman.address,
    secondChairman.address,
    wallet.address,
  ]);

  await membership.deployed();
  await wallet.addKnownSender(membership.address);

  // approve new member
  await membership.connect(regularMember).requestMembership();
  await membership
    .connect(firstChairman)
    .approveMembership(regularMember.address);
  await membership
    .connect(secondChairman)
    .approveMembership(regularMember.address);

  // pay membership fees
  const toBePaid = ethers.utils.parseUnits("3", 4); // exactly 30k wei

  await membership.connect(firstChairman).payMembershipFee({
    value: toBePaid,
  });
  await membership.connect(secondChairman).payMembershipFee({
    value: toBePaid,
  });
  await membership.connect(regularMember).payMembershipFee({
    value: toBePaid,
  });

  await mine(2);

  return { membership: membership, wallet: wallet };
}
