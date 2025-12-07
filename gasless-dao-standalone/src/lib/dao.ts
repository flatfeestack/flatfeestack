import { ensureSmartAccountClient, getPublicClient } from "./aaClient";
import { encodeFunctionData } from "viem";
import { DAO_CONTRACT_ADDRESS } from "../config";
import FlatFeeStackDAO from '../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackDAO.json' assert { type: "json" };
import { pushStatus } from "./statusfeed";
import type { ProposalView } from "./types";

const DAO_ABI = FlatFeeStackDAO.abi;

export async function loadAllProposals(): Promise<ProposalView[]> {
    try{
        const publicClient = getPublicClient();
        let proposals = await publicClient.readContract({
            address: DAO_CONTRACT_ADDRESS,
            abi: DAO_ABI,
            functionName: "getAllProposals",
        })as ProposalView[];

        console.log(proposals);
        return proposals;
    }
    catch(ex) {
        console.log(ex);
        return [];
    }
}

export async function getProposal(id: bigint): Promise<ProposalView | null> {
  try {
    const publicClient = getPublicClient();
    const p = await publicClient.readContract({
      address: DAO_CONTRACT_ADDRESS,
      abi: DAO_ABI,
      functionName: "getProposal",
      args: [id],
    }) as ProposalView;

    return p;
  } catch (err) {
    console.error("getProposal error:", err);
    return null;
  }
}

export async function createProposal(description: string, newBylawsHash: number) {
    const client = ensureSmartAccountClient();

    const calldata = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "setNewBylawsHash",
        args: [newBylawsHash]
    });

    const userOpData = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "propose",
        args: [
            [DAO_CONTRACT_ADDRESS],      // targets
            [0n],               // values
            [calldata],         // calldatas
            description
        ]
    });

    pushStatus("Preparing UserOperation ...");
    pushStatus("Waiting for signature ...");

    const userOpHash = await client.sendUserOperation({
        calls: [
            {
            to: DAO_CONTRACT_ADDRESS,
            data: userOpData,
            value: 0n,
            },
        ],
        });
    
    pushStatus('UserOperation is being processed by Bundler ...');
    const receipt = await client.waitForUserOperationReceipt({
        hash: userOpHash,
    });

    pushStatus("Proposal created successfully.");
    return receipt;
}

export async function voteOnProposal(id: bigint, support: number) {
    const client = ensureSmartAccountClient();
    
    const data = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "castVote",
        args: [id, support]
    });

    pushStatus("Preparing UserOperation ...");

    const hash = await client.sendUserOperation({
        calls: [{ to: DAO_ADDRESS, data }]
    });

    pushStatus("Bundler processing ...");

    const receipt = await client.waitForUserOperationReceipt({ hash });
    pushStatus("Vote submitted successfully");

    return receipt;
}
