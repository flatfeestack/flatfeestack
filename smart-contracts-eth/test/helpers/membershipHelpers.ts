import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import type { Contract } from "ethers";
import type { Wallet } from "@ethersproject/wallet";

export async function addNewMember(
  futureMember: SignerWithAddress | Wallet,
  firstCouncilMember: SignerWithAddress,
  secondCouncilMember: SignerWithAddress,
  membershipContract: Contract
) {
  await membershipContract.connect(futureMember).requestMembership();
  await membershipContract
    .connect(firstCouncilMember)
    .approveMembership(futureMember.address);
  await membershipContract
    .connect(secondCouncilMember)
    .approveMembership(futureMember.address);
}
