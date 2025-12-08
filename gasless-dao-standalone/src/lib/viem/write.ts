import type { SmartAccountClient } from "permissionless";
import type { Address, Hex } from "viem";
import type { Logger } from "../logger";

export async function sendUserOp(
    client: SmartAccountClient,
    to: Address,
    data: Hex,
    value: bigint = 0n,
    log: Logger
) {
    log?.info("Waiting for Signature...");
    const hash = await client.sendUserOperation({
        calls: [{ to, data, value }]
    });

    log?.info("Signature received");
    return hash;
}