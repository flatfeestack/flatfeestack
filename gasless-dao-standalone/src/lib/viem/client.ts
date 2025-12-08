import {
    createPublicClient,
    createWalletClient,
    custom,
    http,
    type PublicClient,
    type WalletClient,
    type Address
} from "viem";
import { sepolia } from "viem/chains";
import { entryPoint07Address } from "viem/account-abstraction";
import { createSmartAccountClient, type SmartAccountClient } from "permissionless";
import { toSimpleSmartAccount } from "permissionless/accounts";

import {
    SEPOLIA_RPC_URL,
    PIMLICO_URL,
    PAYMASTER_CONTRACT_ADDRESS
} from "../../config";

export interface ClientBundle {
    publicClient: PublicClient;
    walletClient: WalletClient;
    smartClient: SmartAccountClient;
    smartAccountAddress: Address;
}

export async function createClientBundle(eoa: Address, usePaymaster: boolean): Promise<ClientBundle> {
    const ethereum = (window as any).ethereum;
    if (!ethereum) throw new Error("MetaMask not found.");

    const publicClient = createPublicClient({
        chain: sepolia,
        transport: http(SEPOLIA_RPC_URL, { timeout: 30_000 })
    });

    const walletClient = createWalletClient({
        account: eoa,
        chain: sepolia,
        transport: custom(ethereum)
    });

    const simpleAccount = await toSimpleSmartAccount({
        client: publicClient as any,
        owner: walletClient,
        entryPoint: { address: entryPoint07Address, version: "0.7" }
    });

    const smartClient = createSmartAccountClient({
        account: simpleAccount,
        chain: sepolia,
        bundlerTransport: http(PIMLICO_URL),
        paymaster: {
        getPaymasterData: async () => {
            if (!usePaymaster) return {};
            return {
            paymaster: PAYMASTER_CONTRACT_ADDRESS,
            paymasterVerificationGasLimit: 100000n,
            paymasterPostOpGasLimit: 100000n,
            } as any;
        },
        },
        userOperation: {
        estimateFeesPerGas: async () => {
            const fee = await publicClient.estimateFeesPerGas();
            return {
            maxFeePerGas: fee.maxFeePerGas,
            maxPriorityFeePerGas: fee.maxPriorityFeePerGas,
            };
        },
        },
    });

    // Deploy smart account if needed
    const saCode = await publicClient.getCode({ address: simpleAccount.address });
    const undeployed = !saCode || saCode === "0x";

    if (undeployed) {
        const hash = await smartClient.sendUserOperation({ 
            calls: [
                {
                    to: eoa,
                    value: 0n,
                    data: "0x"
                }
            ]
        });

        await smartClient.waitForUserOperationReceipt({ hash });
    }

    return {
        publicClient,
        walletClient,
        smartClient,
        smartAccountAddress: simpleAccount.address
    };
}