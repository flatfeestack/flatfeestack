import {
  createPublicClient, createWalletClient,
  custom, http, encodeFunctionData, formatEther
} from 'viem';
import { sepolia } from 'viem/chains';
import { entryPoint07Abi, entryPoint07Address } from 'viem/account-abstraction';

import { createSmartAccountClient } from 'permissionless';
import { toSimpleSmartAccount } from 'permissionless/accounts';
import FlatFeeStackPaymaster from '../../artifacts/contracts/FlatFeeStackDAOPaymaster.sol/FlatFeeStackDAOPaymaster.json' assert { type: "json" };

import dotenv from "dotenv";
dotenv.config();

// Loose typing here to avoid fighting generics for a tutorial
let smartAccountClient: any = null;
let eoa: `0x${string}` | null = null;
let smartAccountAddress: `0x${string}` | null = null;
let publicClient: any;
let usePaymaster: Boolean = false;

const pimlicoUrl = process.env.PIMLICO_URL;
const sepoliaRpc = process.env.SEPOLIA_RPC_URL;
const paymasterAddress = process.env.PAYMASTER_CONTRACT_ADDRESS as `0x${string}`;

export async function waitForTxStatus(txHash: `0x${string}`) {
  const receipt = await publicClient.waitForTransactionReceipt({
    hash: txHash,
    pollingInterval: 2000, // 2s
  });

  return {
    status: receipt.status,
    blockNumber: receipt.blockNumber,
    gasUsed: receipt.gasUsed,
  };
}

export async function getGasCost(txHash: `0x${string}`) {
  const receipt = await publicClient.getTransactionReceipt({ hash: txHash });

  const gasUsed = receipt.gasUsed;
  const effectiveGasPrice = receipt.effectiveGasPrice;

  const totalCostWei = gasUsed * effectiveGasPrice;

  return {
    gasUsed,
    effectiveGasPrice,
    totalCostWei,
    totalCostEth: formatEther(BigInt(totalCostWei)),
  };
}

export async function autoConnectIfAuthorized() {
    const anyWindow = window as any;
    const ethereum = anyWindow.ethereum;
    if (!ethereum) return { eoa: null, smartAccountAddress: null, paymasterUsed: null };

    const accounts = await ethereum.request({
      method: "eth_accounts",
    });

    if (accounts && accounts.length > 0) {
      eoa = accounts[0] as `0x${string}`;
      return await initSmartAccount();
    }

    return { eoa: null, smartAccountAddress: null, paymasterUsed: null };
  }

  export async function handleWalletConnect(){
    const anyWindow = window as any;
    const ethereum = anyWindow.ethereum;
    if (!ethereum) {
      throw new Error('MetaMask not found in this browser.');
    }

    await ethereum.request({
      method: "wallet_requestPermissions",
      params: [{ eth_accounts: {} }],
    });

    const accounts = await ethereum.request({
      method: "eth_requestAccounts",
    });

    eoa = accounts[0] as `0x${string}`;
    return await initSmartAccount();
  }

export async function initSmartAccount() {
  const anyWindow = window as any;
  const ethereum = anyWindow.ethereum;
  if (!ethereum) {
    throw new Error('MetaMask not found in this browser.');
  }

  const accounts = (await ethereum.request({
    method: 'eth_requestAccounts',
  })) as string[];
  eoa = accounts[0] as `0x${string}`;

  const chainIdHex = (await ethereum.request({
    method: 'eth_chainId',
  })) as string;
  const sepoliaHex = `0x${sepolia.id.toString(16)}`;
  if (chainIdHex !== sepoliaHex) {
    throw new Error('Please switch MetaMask to the Sepolia network.');
  }

  publicClient = createPublicClient({
    chain: sepolia,
    transport: http(sepoliaRpc, {
      timeout: 30_000,
    }),
  });

  const walletClient = createWalletClient({
    account: eoa,
    chain: sepolia,
    transport: custom(ethereum),
  });

  const simpleAccount = await toSimpleSmartAccount({
    client: publicClient,
    owner: walletClient, // MetaMask signer
    entryPoint: {
      address: entryPoint07Address,
      version: '0.7',
    },
  });

  smartAccountAddress = simpleAccount.address;
  usePaymaster = await checkIsMember();

  smartAccountClient = createSmartAccountClient({
    account: simpleAccount,
    chain: sepolia,
    bundlerTransport: http(pimlicoUrl),
    paymaster: {
      getPaymasterData: async () => {
        if (!usePaymaster) {
          return {}; // no paymaster
        }

        return {
          paymaster: paymasterAddress,
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

  return {
    eoa,
    smartAccountAddress,
    paymasterUsed: usePaymaster
  };
}

export function disconnectWallet() {
  eoa = null;
  smartAccountAddress = null;
  smartAccountClient = null;

  console.log("Wallet disconnected");
}

export async function getPaymasterDeposit() {
  const deposit = await publicClient.readContract({
    address: entryPoint07Address,
    abi: entryPoint07Abi,
    functionName: "balanceOf",
    args: [paymasterAddress],
  });

  return {
    raw: deposit,
    eth: formatEther(deposit),
  };
}

export function getSmartAccountAddress() {
  if (!smartAccountAddress) {
    throw new Error('Smart account not initialized yet.');
  }
  return smartAccountAddress;
}

export async function incrementCounter(onStatus?: (s: string) => void,) {
  let counterAddress = process.env.COUNTER_CONTRACT_ADDRESS;
  const data = encodeFunctionData({
    abi: [
      {
        name: 'increment',
        type: 'function',
        stateMutability: 'nonpayable',
        inputs: [],
        outputs: [],
      },
    ],
    functionName: 'increment',
    args: [],
  });

  onStatus?.("sending userOp...");
  const userOpHash = await smartAccountClient.sendUserOperation({
    calls: [
      {
        to: counterAddress,
        data,
        value: 0n,
      },
    ],
  });

  onStatus?.("waiting for receipt...");
  const txHash = await smartAccountClient.waitForUserOperationReceipt({
    hash: userOpHash,
  });

  return { userOpHash, txHash };
}

export async function checkIsMember() {
  const paymasterAbi = FlatFeeStackPaymaster.abi;
  const resultSA = await publicClient.readContract({
    address: paymasterAddress,
    abi: paymasterAbi,
    functionName: "isAuthorizedMember",
    args: [smartAccountAddress],
  });
  console.log(resultSA);

  const resultEOA = await publicClient.readContract({
    address: paymasterAddress,
    abi: paymasterAbi,
    functionName: "isAuthorizedMember",
    args: [eoa],
  });
  console.log(resultEOA);
  console.log(eoa);
  
  return resultSA || resultEOA;
}

export async function waitForTransactionReceipt(txHash: any){
  return await publicClient.waitForTransactionReceipt({
        hash: txHash,
        pollingInterval: 2000,
      });
}

export async function waitForConfirmations(
  receipt: { blockNumber: bigint },
  confirmationsRequired: number,
  onUpdate: (text: string) => void
) {
  let confirmations = 0;

  while (confirmations < confirmationsRequired) {
    await new Promise(r => setTimeout(r, 2000));

    const block = await publicClient.getBlockNumber();
    confirmations = Number(block - receipt.blockNumber);

    onUpdate(`confirmations: ${confirmations}/${confirmationsRequired}`);
  }

  return confirmations;
}

export async function getBalance(address: `0x${string}`) {
  console.log(address);
  return publicClient.getBalance({ address });
}