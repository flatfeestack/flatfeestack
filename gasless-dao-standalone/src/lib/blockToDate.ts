import { getPublicClient } from "./aaClient";

export async function blockToDate(targetBlock: bigint) {
    const publicClient = getPublicClient();
    const currentBlock = await publicClient.getBlock();
    const currentNumber = currentBlock.number;
    const currentTs = Number(currentBlock.timestamp);

    const AVG_BLOCK_TIME = 12; // seconds (Sepolia)

    const diff = Number(targetBlock - currentNumber);
    const secondsUntil = diff * AVG_BLOCK_TIME;

    return new Date((currentTs + secondsUntil) * 1000);
}