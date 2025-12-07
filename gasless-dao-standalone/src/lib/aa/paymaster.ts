import type { Address, Abi } from "viem";
import { readContract } from "../viem/read";
import FlatFeeStackPaymaster from "../../../artifacts/contracts/FlatFeeStackDAOPaymaster.sol/FlatFeeStackDAOPaymaster.json";

import {
    PAYMASTER_CONTRACT_ADDRESS
} from "../../config";

const PAYMASTER_ABI = FlatFeeStackPaymaster.abi as Abi;

export async function isMember(
    publicClient: any,
    smart: Address,
    eoa: Address
): Promise<boolean> {
    try {
        const sa = await readContract<boolean>(
            publicClient,
            PAYMASTER_CONTRACT_ADDRESS,
            PAYMASTER_ABI,
            "isAuthorizedMember",
            [smart]
        );

        const eo = await readContract<boolean>(
            publicClient,
            PAYMASTER_CONTRACT_ADDRESS,
            PAYMASTER_ABI,
            "isAuthorizedMember",
            [eoa]
        );

        return sa || eo;
    } catch {
        return false;
    }
}

export async function getPaymasterDeposit(publicClient: any) {
    const raw = await publicClient.getBalance({
        address: PAYMASTER_CONTRACT_ADDRESS
    });
    return raw;
}
