import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import type { Contract } from "ethers";
import type { Wallet } from "@ethersproject/wallet";

export async function addNewMember(
  futureMember: SignerWithAddress | Wallet,
  firstChairman: SignerWithAddress,
  secondChairman: SignerWithAddress,
  membershipContract: Contract
) {
  await membershipContract.connect(futureMember).requestMembership();
  await membershipContract
    .connect(firstChairman)
    .approveMembership(futureMember.address);
  await membershipContract
    .connect(secondChairman)
    .approveMembership(futureMember.address);
}
