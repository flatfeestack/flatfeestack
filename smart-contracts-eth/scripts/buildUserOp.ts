import { ethers } from "hardhat";

export interface PackedUserOperation {
  sender: string;
  nonce: bigint;
  initCode: string;
  callData: string;
  accountGasLimits: string;
  preVerificationGas: bigint;
  gasFees: string;
  paymasterAndData: string;
  signature: string;
}

function packUint128Pair(high: bigint, low: bigint): string {
  const max128 = (1n << 128n) - 1n;
  if (high > max128 || low > max128) {
    throw new Error("value too big for uint128");
  }
  const packed = (high << 128n) | low;
  return ethers.hexlify(ethers.toBeHex(packed, 32));
}

export async function buildUserOp(
  sender: string,
  nonce: bigint,
  callData: string
): Promise<PackedUserOperation> {
  const callGasLimit = 3_000_000n;
  const verificationGasLimit = 3_000_000n;
  const preVerificationGas = 100_000n;

  const feeData = await ethers.provider.getFeeData();
  const maxFeePerGas = feeData.maxFeePerGas ?? feeData.gasPrice ?? 1n;
  const maxPriorityFeePerGas = feeData.maxPriorityFeePerGas ?? feeData.gasPrice ?? 1n;

  const accountGasLimits = packUint128Pair(verificationGasLimit, callGasLimit);
  const gasFees = packUint128Pair(maxPriorityFeePerGas, maxFeePerGas);

  return {
    sender,
    nonce,
    initCode: "0x",
    callData,
    accountGasLimits,
    preVerificationGas,
    gasFees,
    paymasterAndData: "0x",
    signature: "0x1234",         // dummy sig
  };
}
