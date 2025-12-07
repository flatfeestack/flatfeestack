import { ensureSmartAccountClient, ensurePublicClient } from "./aaClient";
import { encodeFunctionData } from "viem";
import { DAO_CONTRACT_ADDRESS } from "../config";
import FlatFeeStackDAO from '../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackDAO.json' assert { type: "json" };
import type { ProposalView } from "./types";

const DAO_ABI = FlatFeeStackDAO.abi;

export async function loadAllProposals(): Promise<ProposalView[]> {
    try{
        const publicClient = ensurePublicClient();
        
        // Get basic proposal list
        let basicProposals = await publicClient.readContract({
            address: DAO_CONTRACT_ADDRESS,
            abi: DAO_ABI,
            functionName: "getAllProposals",
        }) as any[];

        // Fetch full details for each proposal using Governor's base functions
        const detailedProposals = await Promise.all(
            basicProposals.map(async (p: any) => {
                try {
                    // Call state() from Governor base contract
                    const proposalState = await publicClient.readContract({
                        address: DAO_CONTRACT_ADDRESS,
                        abi: DAO_ABI,
                        functionName: "state",
                        args: [p.id],
                    }) as number;

                    // Call proposalVotes() from Governor base contract
                    const votes = await publicClient.readContract({
                        address: DAO_CONTRACT_ADDRESS,
                        abi: DAO_ABI,
                        functionName: "proposalVotes",
                        args: [p.id],
                    }) as [bigint, bigint, bigint];

                    return {
                        id: p.id,
                        description: p.description,
                        startTime: p.startTime,
                        endTime: p.endTime,
                        proposalState: proposalState,
                        againstVotes: votes[0],
                        forVotes: votes[1],
                        abstainVotes: votes[2],
                        proposer: p.proposer,
                    } as ProposalView;
                } catch (err) {
                    console.error(`Error fetching details for proposal ${p.id}:`, err);
                    // Return basic info if detailed fetch fails
                    return {
                        id: p.id,
                        description: p.description,
                        startTime: p.startTime,
                        endTime: p.endTime,
                        proposalState: 0,
                        againstVotes: 0n,
                        forVotes: 0n,
                        abstainVotes: 0n,
                        proposer: p.proposer,
                    } as ProposalView;
                }
            })
        );

        console.log(detailedProposals);
        return detailedProposals;
    }
    catch(ex) {
        console.log(ex);
        return [];
    }
}

export async function getProposal(id: bigint): Promise<ProposalView | null> {
  try {
    const publicClient = ensurePublicClient();
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

/*export async function createProposal(description: string, newBylawsHash: number, onStatus?: (s: string) => void) {
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
            [0n],                        // values
            [calldata],                  // calldatas
            description
        ]
    });

    onStatus?.("Preparing UserOperation ...");
    onStatus?.("Waiting for signature ...");

    const userOpHash = await client.sendUserOperation({
        calls: [
            {
            to: DAO_CONTRACT_ADDRESS,
            data: userOpData,
            value: 0n,
            },
        ],
        });
    
    onStatus?.('UserOperation is being processed by Bundler ...');
    const receipt = await client.waitForUserOperationReceipt({
        hash: userOpHash,
    });

    onStatus?.("Proposal created successfully.");
    return receipt;
}*/

export async function voteOnProposal(id: bigint, support: number, onStatus?: (s: string) => void) {
    const client = ensureSmartAccountClient();
    
    const data = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "castVote",
        args: [id, support]
    });

    onStatus?.("Preparing UserOperation ...");

    const hash = await client.sendUserOperation({
        calls: [{ to: DAO_CONTRACT_ADDRESS, data }]
    });

    onStatus?.("Bundler processing ...");

    const receipt = await client.waitForUserOperationReceipt({ hash });
    onStatus?.("Vote submitted successfully");

    return receipt;
}

export async function setVotingDelayViaCouncil(delayInSeconds: number, onStatus?: (s: string) => void) {
    // Encode the setVotingDelay call
    const setVotingDelayCalldata = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "setVotingDelay",
        args: [delayInSeconds]
    });

    onStatus?.(`Preparing to set voting delay to ${delayInSeconds} seconds...`);
    onStatus?.("⚠️ This requires TWO council signatures!");
    onStatus?.("You'll need to provide a signature from a second council member.");
    
    // For now, return the data needed for manual execution
    // In a full implementation, this would handle the signature collection
    return {
        targets: [DAO_CONTRACT_ADDRESS],
        values: [0n],
        calldatas: [setVotingDelayCalldata],
        description: `Set voting delay to ${delayInSeconds} seconds`,
        encodedCalldata: setVotingDelayCalldata,
    };
}
