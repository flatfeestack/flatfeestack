import type { Abi, Address } from "viem";
import type { AAContext } from "./context";
import { readContract } from "../viem/read";

import FlatFeeStackDAO from "../../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackDAO.json";
import { DAO_CONTRACT_ADDRESS } from "../../config";

const DAO_ABI = FlatFeeStackDAO.abi as Abi;

export function delay(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function getDaoTokenAddress(ctx: AAContext): Promise<Address> {
    return readContract<Address>(
        ctx.public,
        DAO_CONTRACT_ADDRESS,
        DAO_ABI,
        "token"
    );
}

export async function ensureSmartAccountHasFunds(
    ctx: AAContext,
    requiredAmount: bigint,
    log?: any
) {
    const saAddr = ctx.smartAccount;
    const publicClient = ctx.public;

    const currentBalance = await publicClient.getBalance({ address: saAddr });

    if (currentBalance >= requiredAmount) {
        log?.info?.(
            `Smart account already has sufficient ETH (${publicClient.formatEther?.(currentBalance) ?? currentBalance} ETH)`
        );
        return;
    }

    const lacking = requiredAmount - currentBalance;

    log?.info?.(
        `Smart account is missing ${publicClient.formatEther?.(lacking) ?? lacking} ETH, sending from EOA...`
    );

    // EOA sends ETH directly to the smart account
    const txHash = await ctx.wallet.sendTransaction({
        to: saAddr,
        value: lacking
    });

    await publicClient.waitForTransactionReceipt({ hash: txHash });

    log?.info?.(
        `Sent ${publicClient.formatEther?.(lacking) ?? lacking} ETH → Smart Account`
    );
}
