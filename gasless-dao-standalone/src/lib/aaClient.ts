import {
  createPublicClient, createWalletClient,
  custom, http, encodeFunctionData, formatEther,
  type PublicClient,
  type WalletClient,
  type Hex,
  type Address,
} from 'viem';
import { sepolia } from 'viem/chains';
import { entryPoint07Abi, entryPoint07Address } from 'viem/account-abstraction';
import { PIMLICO_URL, SEPOLIA_RPC_URL, PAYMASTER_CONTRACT_ADDRESS, COUNTER_CONTRACT_ADDRESS } from "../config"
import { createSmartAccountClient, type SmartAccountClient } from 'permissionless';
import { toSimpleSmartAccount } from 'permissionless/accounts';
import FlatFeeStackPaymaster from '../../artifacts/contracts/FlatFeeStackDAOPaymaster.sol/FlatFeeStackDAOPaymaster.json' assert { type: "json" };
import FlatFeeStackDAO from '../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackDAO.json' assert { type: "json" };
import FlatFeeStackNFT from '../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackNFT.json' assert { type: "json" };
import { DAO_CONTRACT_ADDRESS } from "../config";
import { type MembershipTokenInfo } from './types'

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

export function ensurePublicClient(): PublicClient {
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

export function getSmartAccountAddress(): Address {
  if (!state.smartAccountAddress) {
    throw new Error('Smart account not initialized yet.');
  }
  return state.smartAccountAddress;
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

  await ensureSmartAccountDeployment(publicClient, simpleAccount);

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

async function ensureSmartAccountDeployment(publicClient: PublicClient, simpleAccount: any){
  const saAddress = simpleAccount.address;
  const saCode = await publicClient.getCode({ address: saAddress });
  const isDeployed = saCode && saCode !== "0x";

  if (!isDeployed) {
    console.log("Smart Account not deployed — deploying now...");

    const deployClient = createSmartAccountClient({
      account: simpleAccount,
      chain: sepolia,
      bundlerTransport: http(PIMLICO_URL),
    });

    const deployOpHash = await deployClient.sendUserOperation({
      calls: [],
    });

    await deployClient.waitForUserOperationReceipt({ hash: deployOpHash });

    console.log("Smart Account deployed:", saAddress);
  }
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

export async function incrementCounter(onStatus?: (s: string) => void) {
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

  onStatus?.("Waiting for signature ...");

  const userOpHash = await smartAccountClient.sendUserOperation({
    calls: [
      {
        to: COUNTER_CONTRACT_ADDRESS,
        data,
        value: 0n,
      },
    ],
  });

  onStatus?.('UserOperation is being processed by Bundler ...');
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
  onUpdate: (text: string) => void,
) {
  const publicClient = ensurePublicClient();
  let confirmations = 0;
  let lastUpdatedConfirmation = -1;

  while (confirmations < confirmationsRequired) {
    await new Promise((r) => setTimeout(r, 2000));

    const block = await publicClient.getBlockNumber();
    confirmations = Number(block - receipt.blockNumber);

    if (lastUpdatedConfirmation < 0 || lastUpdatedConfirmation < confirmations){
      onUpdate(`Confirmations: ${confirmations}/${confirmationsRequired}`);
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

/**
 * Resolve the NFT (membership) contract address from the DAO
 */
export async function getDaoTokenAddress(): Promise<Address> {
  const publicClient = ensurePublicClient();

  return publicClient.readContract({
    address: DAO_CONTRACT_ADDRESS,
    abi: FlatFeeStackDAO.abi as any,
    functionName: 'token',
  }) as Promise<Address>;
}

/**
 * List membership NFTs owned by a given address
 */
export async function listMembershipTokens(owner: Address): Promise<MembershipTokenInfo[]> {
  const publicClient = ensurePublicClient();
  const nftAddress = await getDaoTokenAddress();

  const balance: bigint = await publicClient.readContract({
    address: nftAddress,
    abi: FlatFeeStackNFT.abi as any,
    functionName: 'balanceOf',
    args: [owner],
  });

  const results: MembershipTokenInfo[] = [];

  for (let i = 0n; i < balance; i++) {
    const tokenId: bigint = await publicClient.readContract({
      address: nftAddress,
      abi: FlatFeeStackNFT.abi as any,
      functionName: 'tokenOfOwnerByIndex',
      args: [owner, i],
    });

    const paidUntil: bigint = await publicClient.readContract({
      address: nftAddress,
      abi: FlatFeeStackNFT.abi as any,
      functionName: 'membershipPayed',
      args: [tokenId],
    });

    results.push({ tokenId, membershipPaidUntil: paidUntil });
  }

  console.log(results);
  return results;
}

/**
 * Renew/pay membership for a token owned by the smart account or eoa
 * This makes the token a valid membership (membershipPayed >= now)
 * Requires the EOA to have ETH for the membership fee (0.1 ETH by default)
 */
export async function renewMembership(tokenId: bigint, onStatus?: (s: string) => void) {
  const smartAccountClient = ensureSmartAccountClient();
  const publicClient = ensurePublicClient();
  const nftAddress = await getDaoTokenAddress();

  // Get the membership fee
  const membershipFee: bigint = await publicClient.readContract({
    address: nftAddress,
    abi: FlatFeeStackNFT.abi as any,
    functionName: 'membershipFee',
  });

  onStatus?.(`Renewing membership for token #${tokenId}...`);
  onStatus?.(`Membership fee: ${formatEther(membershipFee)} ETH`);

  await ensureSmartAccountHasFunds(membershipFee, onStatus);

  const data = encodeFunctionData({
    abi: FlatFeeStackNFT.abi as any,
    functionName: 'payMembership',
    args: [tokenId],
  });

  onStatus?.('Submitting membership payment UserOperation...');

  // EOA sends the fee, Smart Account signs the operation
  const userOpHash = await smartAccountClient.sendUserOperation({
    calls: [
      {
        to: nftAddress,
        data,
        value: membershipFee,
      },
    ],
    onSigningRequest: async ({ request }) => {
      // This ensures MetaMask asks the EOA to approve sending funds
      return await state.walletClient!.signTransaction(request);
    },
  });

  onStatus?.('Waiting for membership renewal confirmation...');
  const receipt = await smartAccountClient.waitForUserOperationReceipt({
    hash: userOpHash,
  });

  onStatus?.(`✓ Membership renewed for token #${tokenId}`);
  return { userOpHash, receipt };
}

async function ensureSmartAccountHasFunds(amount: bigint, onStatus?: (msg: string) => void) {
  const publicClient = ensurePublicClient();
  const walletClient = state.walletClient!;
  const smartAccountAddress = state.smartAccountAddress!;

  // 1. Get smart account balance
  const saBalance = await publicClient.getBalance({ address: smartAccountAddress });

  if (saBalance >= amount) {
    onStatus?.(`Smart Account already has enough ETH (${formatEther(saBalance)}).`);
    return;
  }

  // 2. Calculate missing amount
  const needed = amount - saBalance;
  onStatus?.(`Smart Account needs ${formatEther(needed)} ETH. Sending from EOA...`);

  // 3. Send ETH from EOA -> Smart Account
  const txHash = await walletClient.sendTransaction({
    to: smartAccountAddress,
    value: needed,
  });

  onStatus?.(`Waiting for deposit tx...`);
  await publicClient.waitForTransactionReceipt({ hash: txHash });

  onStatus?.(`✓ Sent ${formatEther(needed)} ETH to Smart Account.`);
}

/**
 * Create a governance proposal via smart account (gasless if member)
 * @param proposal - Object containing targets, values, calldatas, and description
 * @param onStatus - Callback for status updates
 * @returns proposalId and receipt
 */
export async function createProposal(
  proposal: ProposalDetails,
  onStatus?: (s: string) => void
) {
  const smartAccountClient = ensureSmartAccountClient();
  const publicClient = ensurePublicClient();

  onStatus?.("Checking membership authorization...");
  
  // Verify membership before attempting proposal
  const isMember = await checkIsMember(state.smartAccountAddress, state.eoa);
  onStatus?.(`Membership check: ${isMember ? "✓ Member" : "✗ Not a member"}`);
  
  if (!isMember) {
    throw new Error(
      "Cannot create proposal: You must be a DAO member. " +
      "Ensure your EOA or smart account owns an active membership NFT."
    );
  }

  // Check smart account balance for gas
  const saBalance = await publicClient.getBalance({ address: state.smartAccountAddress! });
  onStatus?.(`Smart account balance: ${formatEther(saBalance)} ETH`);

  // Encode the propose function call
  const data = encodeFunctionData({
    abi: FlatFeeStackDAO.abi as any,
    functionName: 'propose',
    args: [
      proposal.targets,
      proposal.values,
      proposal.calldatas,
      proposal.description,
    ],
  });

  console.log('📝 Encoded proposal data:', {
    to: DAO_CONTRACT_ADDRESS,
    data,
    dataLength: data.length,
    targetCount: proposal.targets.length,
    targets: proposal.targets,
    values: proposal.values.map(v => v.toString()),
    calldatas: proposal.calldatas,
    description: proposal.description,
  });

  if (proposal.targets.length === 0) {
    console.warn('⚠️ WARNING: Creating proposal with NO actions (empty targets)');
  }

  onStatus?.("Preparing proposal UserOperation ...");

  try {
    const userOpHash = await smartAccountClient.sendUserOperation({
      calls: [
        {
          to: DAO_CONTRACT_ADDRESS,
          data,
          value: 0n,
        },
      ],
      // Explicitly set gas limits to avoid estimation failures
      // propose() requires significant gas for storage writes and Governor logic
      callGasLimit: 500000n,
      verificationGasLimit: 500000n,
      preVerificationGas: 100000n,
    });

    console.log('✉️ UserOp sent with hash:', userOpHash);

    onStatus?.("Proposal UserOperation is being processed by Bundler ...");
    const receipt = await smartAccountClient.waitForUserOperationReceipt({
      hash: userOpHash,
    });

    onStatus?.(`UserOp receipt status: ${receipt.receipt.status}`);
    console.log('📦 UserOp Receipt:', {
      userOpHash,
      txHash: receipt.receipt.transactionHash,
      status: receipt.receipt.status,
      gasUsed: receipt.receipt.gasUsed,
      logs: receipt.receipt.logs,
    });

    // Verify proposal was created by checking all proposals
    try {
      const proposals = await publicClient.readContract({
        address: DAO_CONTRACT_ADDRESS,
        abi: FlatFeeStackDAO.abi as any,
        functionName: 'getAllProposals',
      }) as any[];
      onStatus?.(`✅ Proposal created! DAO now has ${proposals.length} proposal(s)`);
      if (proposals.length > 0) {
        const latest = proposals[proposals.length - 1];
        console.log('Latest proposal:', {
          id: latest.id?.toString(),
          description: latest.description,
          proposer: latest.proposer,
        });
      }
    } catch (err: any) {
      console.warn('❌ Could not verify proposal creation:', err.message);
    }

    return { userOpHash, receipt };
  } catch (err: any) {
    // If error includes revert data, try to decode it
    if (err.message?.includes('0x447b05d0')) {
      throw new Error('Proposal creation failed: Check if you have sufficient gas or if the DAO allows proposals');
    }
    throw err;
  }
}