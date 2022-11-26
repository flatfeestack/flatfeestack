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
  chairman: SignerWithAddress,
  whitelisterOne: SignerWithAddress,
  whitelisterTwo: SignerWithAddress
) {
  const wallet = await deployWalletContract(chairman);
  const Membership = await ethers.getContractFactory("Membership");
  const membership = await upgrades.deployProxy(Membership, [
    chairman.address,
    whitelisterOne.address,
    whitelisterTwo.address,
    wallet.address,
  ]);

  await membership.deployed();
  await wallet.addKnownSender(membership.address);

  // pay membership fees
  const toBePaid = ethers.utils.parseUnits("3", 4); // exactly 30k wei

  await membership.connect(chairman).payMembershipFee({
    value: toBePaid,
  });
  await membership.connect(whitelisterOne).payMembershipFee({
    value: toBePaid,
  });
  await membership.connect(whitelisterTwo).payMembershipFee({
    value: toBePaid,
  });

  await mine(2);

  return { membership: membership, wallet: wallet };
}
