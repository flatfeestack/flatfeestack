import type { Abi, Address } from "viem";
import { formatEther } from "viem";
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
    amount: bigint,
    log?: any
) {
    const sa = ctx.smartAccount;
    const publicClient = ctx.public;
    const walletClient = ctx.wallet;

    // check current SA balance
    const balance = await publicClient.getBalance({ address: sa });

    if (balance >= amount) {
        return;
    }

    // compute missing amount
    const needed = amount - balance;
    log?.info?.(
        `Smart Account needs ${formatEther(needed)} ETH. Sending from EOA...`
    );

    // EOA funds the Smart Account
    const txHash = await walletClient.sendTransaction({
        to: sa,
        value: needed
    });

    log?.info?.(`Waiting for deposit tx`);
    await publicClient.waitForTransactionReceipt({ hash: txHash });

    log?.info?.(`✓ Smart Account funded with ${formatEther(needed)} ETH`);
}
