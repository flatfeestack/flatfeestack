import type { SmartAccountClient } from "permissionless";
import type { Address, Hex } from "viem";

export async function sendUserOp(
    client: SmartAccountClient,
    to: Address,
    data: Hex,
    value: bigint = 0n
) {
    const hash = await client.sendUserOperation({
        calls: [{ to, data, value }]
    });

    const receipt = await client.waitForUserOperationReceipt({ hash });
    return { hash, receipt };
}
