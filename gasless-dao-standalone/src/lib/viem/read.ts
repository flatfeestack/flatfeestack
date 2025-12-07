import type { Abi, Address, PublicClient } from "viem";

export function readContract<T = any>(
    publicClient: PublicClient,
    address: Address,
    abi: Abi,
    fn: string,
    args: any[] = []
): Promise<T> {
    return publicClient.readContract({
        address,
        abi,
        functionName: fn,
        args
    }) as Promise<T>;
}
