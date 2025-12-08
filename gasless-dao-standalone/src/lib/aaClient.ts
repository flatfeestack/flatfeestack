import {
  formatEther,
  type PublicClient,
  type Hex,
} from 'viem';
import { entryPoint07Abi, entryPoint07Address } from 'viem/account-abstraction';
import { PAYMASTER_CONTRACT_ADDRESS } from "../config"
import { type Logger } from "./logger";
import { type AAContext } from "./aa/context";

export interface FinalizerOptions {
    confirmations?: number; 
    logger?: Logger;
    setLastTxHash?: (hash: Hex) => void;
    onBalancesUpdated?: () => Promise<void>;
    onPaymasterFunds?: (eth: string) => void;
    onGasCost?: (eth: string) => void;
}

export async function finalizeUserOp(
    userOpHash: Hex,
    ctx: AAContext,
    opts: FinalizerOptions = {}
) {
    const {
        confirmations = 3,
        logger,
        setLastTxHash,
        onBalancesUpdated,
        onPaymasterFunds,
        onGasCost
    } = opts;

    if (!userOpHash) return;

    logger?.info("UserOperation sent to bundler.");
    logger?.info(`Waiting for transaction to be mined...`);

    const publicClient = ctx.public;
    const uoReceipt = await ctx.smartClient.waitForUserOperationReceipt({
        hash: userOpHash
    });

    const txHash = uoReceipt.receipt.transactionHash;
    setLastTxHash?.(txHash);

    const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });

    logger?.info?.(`Transaction mined in block ${receipt.blockNumber}.`);

    const gasCost = await getGasCost(txHash, publicClient);
    if (onGasCost){
      onGasCost(gasCost.totalCostEth);
    }

    const paymaster = await getPaymasterDeposit(publicClient);
    if (onPaymasterFunds) {
        onPaymasterFunds(paymaster.eth);
    }

    if (confirmations > 0) {
        logger?.info(`Waiting for ${confirmations} confirmations...`);
        await waitForConfirmations(
            { blockNumber: receipt.blockNumber },
            confirmations,
            publicClient,
            (msg) => logger?.info(msg)
        );
    }

    logger?.info("UserOp successfull!");

    if (onBalancesUpdated) {
        await onBalancesUpdated();
    }
}

async function getGasCost(txHash: Hex, publicClient: PublicClient) {
  const receipt = await publicClient.getTransactionReceipt({ hash: txHash });

  const gasUsed = receipt.gasUsed;
  const effectiveGasPrice = receipt.effectiveGasPrice;
  const totalCostWei = gasUsed * effectiveGasPrice;

  return {
    gasUsed,
    effectiveGasPrice,
    totalCostWei,
    totalCostEth: formatEther(totalCostWei),
  };
}

async function getPaymasterDeposit(publicClient: PublicClient) {
  const deposit = await publicClient.readContract({
    address: entryPoint07Address,
    abi: entryPoint07Abi,
    functionName: 'balanceOf',
    args: [PAYMASTER_CONTRACT_ADDRESS],
  });

  return {
    raw: deposit,
    eth: formatEther(deposit),
  };
}

export async function waitForConfirmations(
  receipt: { blockNumber: bigint },
  confirmationsRequired: number,
  publicClient: PublicClient,
  onUpdate: (text: string) => void,
) {
  let confirmations = 0;
  let lastUpdatedConfirmation = -1;

  while (confirmations < confirmationsRequired) {
    await new Promise((r) => setTimeout(r, 2000));

    const block = await publicClient.getBlockNumber();
    confirmations = Number(block - receipt.blockNumber);
    if (confirmations == 0) continue;

    if (lastUpdatedConfirmation < 0 || lastUpdatedConfirmation < confirmations){
      onUpdate(`Confirmations: ${confirmations}/${confirmationsRequired}`);
      lastUpdatedConfirmation = confirmations;
    }
  }
}