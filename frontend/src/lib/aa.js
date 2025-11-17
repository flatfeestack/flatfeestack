import { encodeFunctionData, getContract, createPublicClient, http } from "viem";
import { getUserOperationHash } from "viem/account-abstraction";
import { sepolia } from "viem/chains";

const ENTRYPOINT_ADDRESS = "0x0000000071727De22E5E9d8BAf0edAc6f37da032";
const PAYMASTER = "0x92975fc534F1aC8a12826D45C00f501eE397e660";
const BUNDLER_URL = "https://api.pimlico.io/v2/11155111/rpc?apikey=pim_Y3GiAJnZqiZ7dF3cQJxamJ";

const SMART_ACCOUNT_ABI = [
    {
        name: "execute",
        type: "function",
        inputs: [
            { type: "address", name: "dest" },
            { type: "uint256", name: "value" },
            { type: "bytes", name: "func" },
        ],
    },
    {
        name: "nonce",
        type: "function",
        inputs: [],
        outputs: [{ type: "uint256" }],
        stateMutability: "view",
    },
];

const entryPointAbi = [
  {
    "inputs": [
      {
        "components": [
          { "internalType": "address", "name": "sender", "type": "address" },
          { "internalType": "bytes21", "name": "nonce", "type": "bytes21" },
          { "internalType": "bytes", "name": "initCode", "type": "bytes" },
          { "internalType": "bytes", "name": "callData", "type": "bytes" },
          { "internalType": "uint256", "name": "callGasLimit", "type": "uint256" },
          { "internalType": "uint256", "name": "verificationGasLimit", "type": "uint256" },
          { "internalType": "uint256", "name": "preVerificationGas", "type": "uint256" },
          { "internalType": "uint256", "name": "maxFeePerGas", "type": "uint256" },
          { "internalType": "uint256", "name": "maxPriorityFeePerGas", "type": "uint256" },
          { "internalType": "bytes", "name": "paymasterAndData", "type": "bytes" },
          { "internalType": "bytes", "name": "signature", "type": "bytes" }
        ],
        "internalType": "struct PackedUserOperation",
        "name": "userOp",
        "type": "tuple"
      }
    ],
    "name": "getUserOpHash",
    "outputs": [{ "internalType": "bytes32", "name": "", "type": "bytes32" }],
    "stateMutability": "view",
    "type": "function"
  },
    {
        "inputs": [{ "internalType": "address", "name": "account", "type": "address" }],
        "name": "balanceOf",
        "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
        "stateMutability": "view",
        "type": "function"
    },
    {
    "inputs": [
      { "internalType": "address", "name": "sender", "type": "address" },
      { "internalType": "uint192", "name": "key", "type": "uint192" }
    ],
    "name": "getNonce",
    "outputs": [
      { "internalType": "uint256", "name": "nonce", "type": "uint256" }
    ],
    "stateMutability": "view",
    "type": "function"
  }

    ];

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
        userOpForHash.paymaster = PAYMASTER;
        userOpForHash.paymasterData = "0x";
    }

    const userOpHash = getUserOperationHash({
        chainId: Number(chainId),
        entryPointAddress: ENTRYPOINT_ADDRESS,
        entryPointVersion: "0.7",
        userOperation: userOpForHash,
    });

    // Sign with EOA
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
        userOp.paymaster = PAYMASTER;
        userOp.paymasterData = "0x";
    }

    // Send to bundler
    const response = await fetch(BUNDLER_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            jsonrpc: "2.0",
            id: 1,
            method: "eth_sendUserOperation",
            params: [userOp, ENTRYPOINT_ADDRESS],
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

        // Ask the bundler / node for execution status
        const res = await fetch(BUNDLER_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            jsonrpc: "2.0",
            id: 1,
            method: "eth_getUserOperationReceipt",
            params: [userOpHash]
        }),
        }).then((r) => r.json());

        // Bundler returned something ?
        if (res.result) {
        const receipt = res.result;

        // SUCCESS
        if (receipt.receipt?.status === "0x1") {
            onUpdate("UserOp succeeded 🎉");
            return { success: true, ...receipt };
        }

        // EXECUTION FAILED (AA error or revert)
        if (receipt.receipt?.status === "0x0") {
            onUpdate("UserOp reverted ❌");
            return { success: false, ...receipt };
        }

        // Some bundlers return txHash but no status
        if (receipt.txHash) {
            onUpdate("UserOp included in a tx");
            return { success: null, ...receipt };
        }
        }

        // Not mined yet → keep polling
        onUpdate(`Waiting for UserOp… (${tries}/${maxRetries})`);

        await new Promise((resolve) => setTimeout(resolve, pollInterval));
    }

    // Timeout
    onUpdate("UserOp polling timed out ⏳");
    return { success: null, timeout: true };
}