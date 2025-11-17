import { encodeFunctionData } from "viem";
import { getUserOperationHash } from "viem/account-abstraction";
import { ENTRY_POINT_ADDRESS, PAYMASTER_ADDRESS, RPC_URL } from "../lib/config"
import {SMART_ACCOUNT_ABI} from "../lib/abi"

export async function executeSmartAccountTransaction(
    wallet,
    smartAccountAddress,
    eoaAddress,
    targetContract,
    targetAbi,
    functionName,
    args,
    usePaymaster
) {
    // Check balance
    const balance = await wallet.getBalance(smartAccountAddress);
    const balanceInWei = BigInt(balance);
    if (balanceInWei === 0n) {
        throw new Error(`Smart account ${smartAccountAddress} has no funds`);
    }

    // Get nonce
    const nonceData = encodeFunctionData({
        abi: SMART_ACCOUNT_ABI,
        functionName: "nonce",
    });
    const nonceResult = await wallet.callContract(smartAccountAddress, nonceData);
    const nonce = BigInt(nonceResult);

    // Build call data
    const targetCallData = encodeFunctionData({
        abi: targetAbi,
        functionName,
        args,
    });
    const callData = encodeFunctionData({
        abi: SMART_ACCOUNT_ABI,
        functionName: "execute",
        args: [targetContract, 0n, targetCallData],
    });

    // Get gas prices
    const feeHistory = await wallet.getFeeHistory("0x1", [50]);
    const baseFee = BigInt(feeHistory.baseFeePerGas[0]);
    const priorityFee = BigInt(feeHistory.reward[0][0]);

    const minPriorityFee = 100000000n;
    const minMaxFee = 100000025n;

    const actualPriorityFee = priorityFee > minPriorityFee ? priorityFee : minPriorityFee;
    const actualMaxFee =
        baseFee + actualPriorityFee > minMaxFee
            ? baseFee + actualPriorityFee
            : minMaxFee;

    const chainId = await wallet.getChainId();

    // Calculate UserOperation hash
    const userOpForHash = {
            sender: smartAccountAddress,
            nonce: "0x" + nonce.toString(16),
            callData: callData,
            callGasLimit: BigInt("0x70000"),
            verificationGasLimit: BigInt("0x40000"),
            preVerificationGas: BigInt("0x10000"),
            maxFeePerGas: BigInt(actualMaxFee),
            maxPriorityFeePerGas: BigInt(actualPriorityFee)
        };
    
    if (usePaymaster){
        userOpForHash.paymaster = PAYMASTER_ADDRESS;
        userOpForHash.paymasterData = "0x";
    }

    const userOpHash = getUserOperationHash({
        chainId: Number(chainId),
        entryPointAddress: ENTRY_POINT_ADDRESS,
        entryPointVersion: "0.7",
        userOperation: userOpForHash,
    });

    const signature = await wallet.personalSign(userOpHash, eoaAddress);

    // Build UserOperation
    const userOp = {
        sender: smartAccountAddress,
        nonce: "0x" + nonce.toString(16),
        callData: callData,
        callGasLimit: "0x70000",
        verificationGasLimit: "0x40000",
        preVerificationGas: "0x10000",
        maxFeePerGas: "0x" + actualMaxFee.toString(16),
        maxPriorityFeePerGas: "0x" + actualPriorityFee.toString(16),
        signature: signature
    };

    if (usePaymaster){
        userOp.paymaster = PAYMASTER_ADDRESS;
        userOp.paymasterData = "0x";
    }

    // Send to bundler
    const response = await fetch(RPC_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            jsonrpc: "2.0",
            id: 1,
            method: "eth_sendUserOperation",
            params: [userOp, ENTRY_POINT_ADDRESS],
        }),
    });

    const result = await response.json();

    if (result.error) {
        throw new Error(`Bundler error: ${result.error.message}`);
    }

    return result.result;
}

export async function pollUserOp(
    userOpHash,
    {
        pollInterval = 2000,
        maxRetries = 60,
        onUpdate = (status) => {}
    } = {}
    ) {
    let tries = 0;

    while (tries < maxRetries) {
        tries++;

        const res = await fetch(RPC_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            jsonrpc: "2.0",
            id: 1,
            method: "eth_getUserOperationReceipt",
            params: [userOpHash]
        }),
        }).then((r) => r.json());

        if (res.result) {
        const receipt = res.result;

        if (receipt.receipt?.status === "0x1") {
            onUpdate("UserOp succeeded!");
            return { success: true, ...receipt };
        }

        if (receipt.receipt?.status === "0x0") {
            onUpdate("UserOp reverted");
            return { success: false, ...receipt };
        }

        if (receipt.txHash) {
            onUpdate("UserOp included in a tx");
            return { success: null, ...receipt };
        }
        }

        onUpdate(`Waiting for UserOp… (${tries}/${maxRetries})`);

        await new Promise((resolve) => setTimeout(resolve, pollInterval));
    }

    onUpdate("UserOp polling timed out ⏳");
    return { success: null, timeout: true };
}