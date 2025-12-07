import { createLogger } from "../logger";
import { createClientBundle } from "../viem/client";
import { isMember } from "./paymaster";

import type { Address } from "viem";

export type AAContext = {
    eoa: Address;
    smartAccount: Address;
    public: any;
    wallet: any;
    smartClient: any;
    isMember: boolean;
    usePaymaster: boolean;
};

export async function createAAContext(eoa: Address, log = createLogger()) {
    const { publicClient, walletClient, smartClient, smartAccountAddress } =
        await createClientBundle(eoa, true);

    const membership = await isMember(publicClient, smartAccountAddress, eoa);

    const ctx: AAContext = {
        eoa,
        smartAccount: smartAccountAddress,
        public: publicClient,
        wallet: walletClient,
        smartClient,
        isMember: membership,
        usePaymaster: membership
    };

    log.info(`Membership: ${ctx.isMember}`);

    return ctx;
}
