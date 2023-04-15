import type { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import type { Signature } from "@ethersproject/bytes";
import { ethers } from "hardhat";

async function generateSignature(
  amount: Number,
  owner: SignerWithAddress,
  userId: string,
  symbol: string
) {
  const encodedUserId = ethers.utils.id(userId);
  const payload = ethers.utils.defaultAbiCoder.encode(
    ["bytes32", "string", "uint256", "string"],
    [encodedUserId, "#", amount, symbol]
  );
  const payloadHash = ethers.utils.keccak256(payload);

  const signature = await owner.signMessage(ethers.utils.arrayify(payloadHash));

  return {
    encodedUserId,
    signature: ethers.utils.splitSignature(signature),
  };
}

export default generateSignature;
