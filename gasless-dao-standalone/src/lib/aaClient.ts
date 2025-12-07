import {
  createPublicClient, createWalletClient,
  custom, http, encodeFunctionData, formatEther, parseEther,
  type PublicClient,
  type WalletClient,
  type Hex,
  type Address,
} from 'viem';
import { sepolia } from 'viem/chains';
import { entryPoint07Abi, entryPoint07Address } from 'viem/account-abstraction';
import { PIMLICO_URL, SEPOLIA_RPC_URL, PAYMASTER_CONTRACT_ADDRESS, COUNTER_CONTRACT_ADDRESS, NFT_CONTRACT_ADDRESS } from "../config"
import { createSmartAccountClient, type SmartAccountClient } from 'permissionless';
import { toSimpleSmartAccount } from 'permissionless/accounts';
import FlatFeeStackPaymaster from '../../artifacts/contracts/FlatFeeStackDAOPaymaster.sol/FlatFeeStackDAOPaymaster.json' assert { type: "json" };
import { pushStatus } from "./statusfeed";

type AAClient = SmartAccountClient<any>;

type HelperState = {
  eoa?: Address;
  smartAccountAddress?: Address;
  publicClient?: PublicClient;
  walletClient?: WalletClient;
  smartAccountClient?: AAClient;
  usePaymaster: boolean;
};

const state: HelperState = {
  usePaymaster: false,
};

function getEthereum(): any {
  if (typeof window === 'undefined') {
    throw new Error('Ethereum provider is only available in the browser.');
  }

  const anyWindow = window as any;
  const ethereum = anyWindow.ethereum;
  if (!ethereum) {
    throw new Error('MetaMask not found in this browser.');
  }

  return ethereum;
}

function ensurePublicClient(): PublicClient {
  if (!state.publicClient) {
    throw new Error('Public client not initialized');
  }

  return state.publicClient;
}

export function ensureSmartAccountClient(): AAClient {
  if (!state.smartAccountClient || !state.smartAccountAddress) {
    throw new Error('Smart account not initialized');
  }

  return state.smartAccountClient;
}

export function getPublicClient(): PublicClient {
  if (!state.publicClient)
    throw new Error("PublicClient not initialized");
  
  return state.publicClient;
}

async function setupSmartAccountFromEOA(eoa: Address) {
  const ethereum = getEthereum();

  const chainIdHex = (await ethereum.request({ method: 'eth_chainId' })) as string;
  const sepoliaHex = `0x${sepolia.id.toString(16)}`;
  if (chainIdHex !== sepoliaHex) {
    throw new Error('Please switch MetaMask to the Sepolia network.');
  }

  if (!state.publicClient) {
    state.publicClient = createPublicClient({
      chain: sepolia,
      transport: http(SEPOLIA_RPC_URL, { timeout: 30_000 }),
    });
  }

  const publicClient = state.publicClient;

  const walletClient = createWalletClient({
    account: eoa,
    chain: sepolia,
    transport: custom(ethereum),
  });

  const simpleAccount = await toSimpleSmartAccount({
    client: publicClient as any,
    owner: walletClient,
    entryPoint: {
      address: entryPoint07Address,
      version: '0.7',
    },
  });

  state.eoa = eoa;
  state.publicClient = publicClient;
  state.walletClient = walletClient;
  state.smartAccountAddress = simpleAccount.address;

  const usePaymaster = await checkIsMember(simpleAccount.address, eoa);
  state.usePaymaster = usePaymaster;

  const smartAccountClient = createSmartAccountClient({
    account: simpleAccount,
    chain: sepolia,
    bundlerTransport: http(PIMLICO_URL),
    paymaster: {
      getPaymasterData: async () => {
        if (!state.usePaymaster) return {};
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

  state.smartAccountClient = smartAccountClient;

  return {
    eoa,
    smartAccountAddress: state.smartAccountAddress,
    paymasterUsed: state.usePaymaster,
  };
}

export async function getGasCost(txHash: Hex) {
  const publicClient = ensurePublicClient();
  const receipt = await publicClient.getTransactionReceipt({ hash: txHash });

  const gasUsed = receipt.gasUsed;
  const effectiveGasPrice = receipt.effectiveGasPrice;
  const totalCostWei = gasUsed * effectiveGasPrice;

  return {
    gasUsed,
    effectiveGasPrice,
    totalCostWei,
    totalCostEth: formatEther(totalCostWei),
  };
}

export async function autoConnectIfAuthorized() {
  try {
    const ethereum = getEthereum();

    const accounts = (await ethereum.request({
      method: 'eth_accounts',
    })) as string[];

    if (!accounts || accounts.length === 0) {
      return { eoa: null, smartAccountAddress: null, paymasterUsed: null };
    }

    const eoa = accounts[0] as Address;
    return await setupSmartAccountFromEOA(eoa);
  } catch {
    return { eoa: null, smartAccountAddress: null, paymasterUsed: null };
  }
}

export async function handleWalletConnect() {
  const ethereum = getEthereum();

  await ethereum.request({
    method: 'wallet_requestPermissions',
    params: [{ eth_accounts: {} }],
  });

  const accounts = (await ethereum.request({
    method: 'eth_requestAccounts',
  })) as string[];

  const eoa = accounts[0] as Address;
  return setupSmartAccountFromEOA(eoa);
}

export function disconnectWallet() {
  state.eoa = undefined;
  state.smartAccountAddress = undefined;
  state.smartAccountClient = undefined;
  state.walletClient = undefined;
  state.usePaymaster = false;

  console.log('Wallet disconnected');
}

export async function getPaymasterDeposit() {
  const publicClient = ensurePublicClient();

  const deposit = await publicClient.readContract({
    address: entryPoint07Address,
    abi: entryPoint07Abi,
    functionName: 'balanceOf',
    args: [PAYMASTER_CONTRACT_ADDRESS],
  });

  return {
    raw: deposit,
    eth: formatEther(deposit),
  };
}

export function getSmartAccountAddress(): Address {
  if (!state.smartAccountAddress) {
    throw new Error('Smart account not initialized yet.');
  }
  return state.smartAccountAddress;
}

export async function incrementCounter() {
  const smartAccountClient = ensureSmartAccountClient();

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

  pushStatus("Waiting for signature ...");

  const userOpHash = await smartAccountClient.sendUserOperation({
    calls: [
      {
        to: COUNTER_CONTRACT_ADDRESS,
        data,
        value: 0n,
      },
    ],
  });

  pushStatus('UserOperation is being processed by Bundler ...');
  const receipt = await smartAccountClient.waitForUserOperationReceipt({
    hash: userOpHash,
  });

  return { userOpHash, receipt };
}

export async function checkIsMember(smartAccountAddress?: Address, eoa?: Address): Promise<boolean> {
  const publicClient = ensurePublicClient();
  const paymasterAbi = FlatFeeStackPaymaster.abi;

  if (!smartAccountAddress && !eoa) return false;

  try {
    const [resultSA, resultEOA] = await Promise.all([
      smartAccountAddress
        ? publicClient.readContract({
            address: PAYMASTER_CONTRACT_ADDRESS,
            abi: paymasterAbi,
            functionName: 'isAuthorizedMember',
            args: [smartAccountAddress],
          })
        : Promise.resolve(false),
      eoa
        ? publicClient.readContract({
            address: PAYMASTER_CONTRACT_ADDRESS,
            abi: paymasterAbi,
            functionName: 'isAuthorizedMember',
            args: [eoa],
          })
        : Promise.resolve(false),
    ]);

    return Boolean(resultSA || resultEOA);
  } catch (err) {
    console.warn('checkIsMember failed', err);
    return false;
  }
}

export async function waitForTransactionReceipt(txHash: Hex) {
  const publicClient = ensurePublicClient();
  return publicClient.waitForTransactionReceipt({
    hash: txHash,
    pollingInterval: 2000,
  });
}

export async function waitForTxStatus(txHash: Hex) {
  const receipt = await waitForTransactionReceipt(txHash);
  return {
    status: receipt.status,
    blockNumber: receipt.blockNumber,
    gasUsed: receipt.gasUsed,
  };
}

export async function waitForConfirmations(
  receipt: { blockNumber: bigint },
  confirmationsRequired: number,
) {
  const publicClient = ensurePublicClient();
  let confirmations = 0;
  let lastUpdatedConfirmation = -1;

  while (confirmations < confirmationsRequired) {
    await new Promise((r) => setTimeout(r, 2000));

    const block = await publicClient.getBlockNumber();
    confirmations = Number(block - receipt.blockNumber);

    if (lastUpdatedConfirmation < 0 || lastUpdatedConfirmation < confirmations){
      pushStatus(`Confirmations: ${confirmations}/${confirmationsRequired}`);
      lastUpdatedConfirmation = confirmations;
    }
  }

  return confirmations;
}

export async function getBalance(address: Address) {
  const publicClient = ensurePublicClient();
  let balance = await publicClient.getBalance({ address });
  return formatEther(balance);
}

export async function fundPaymaster(amountEth: string) {
  const smartAccountClient = ensureSmartAccountClient();
  const amountWei = parseEther(String(amountEth));

  const data = encodeFunctionData({
    abi: entryPoint07Abi,
    functionName: 'depositTo',
    args: [PAYMASTER_CONTRACT_ADDRESS],
  });

  pushStatus("Waiting for signature ...");

  const userOpHash = await smartAccountClient.sendUserOperation({
    calls: [
      {
        to: entryPoint07Address,
        data,
        value: amountWei,
      },
    ],
  });
}

export async function delegateVotesToSmartAccount(smartAccountAddress: Address) {
  const client = ensureSmartAccountClient();

  const data = encodeFunctionData({
    abi: [
      {
        name: "delegate",
        type: "function",
        stateMutability: "nonpayable",
        inputs: [{ name: "delegatee", type: "address" }]
      }
    ],
    functionName: "delegate",
    args: [smartAccountAddress]
  });

  const userOpHash = await client.sendUserOperation({
    calls: [{ to: NFT_CONTRACT_ADDRESS, data }]
  });

  return client.waitForUserOperationReceipt({ hash: userOpHash });
}