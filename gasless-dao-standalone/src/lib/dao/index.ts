import { encodeFunctionData, type Abi } from "viem";
import { readContract } from "../viem/read";
import FlatFeeStackDAO from "../../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackDAO.json";
import { DAO_CONTRACT_ADDRESS } from "../../config";
import { sendUserOp } from "../viem/write";
import type { AAContext } from "../aa/context";
import { on } from "svelte/events";

const DAO_ABI = FlatFeeStackDAO.abi as Abi;

export async function loadAllProposals(ctx: AAContext) {
    const list = await readContract<any[]>(
        ctx.public,
        DAO_CONTRACT_ADDRESS,
        DAO_ABI,
        "getAllProposals"
    );

    const enriched = await Promise.all(
        list.map(async (p) => {
            const state = await readContract<number>(
                ctx.public,
                DAO_CONTRACT_ADDRESS,
                DAO_ABI,
                "state",
                [p.id]
            );

            const votes = await readContract<[bigint, bigint, bigint]>(
                ctx.public,
                DAO_CONTRACT_ADDRESS,
                DAO_ABI,
                "proposalVotes",
                [p.id]
            );

            return {
                id: p.id,
                description: p.description,
                startTime: p.startTime,
                endTime: p.endTime,
                proposalState: state,
                againstVotes: votes[0],
                forVotes: votes[1],
                abstainVotes: votes[2],
                proposer: p.proposer
            };
        })
    );

    return enriched;
}

export async function voteOnProposal(
    ctx: AAContext,
    id: bigint,
    support: number,
    log: any
) {
    const data = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "castVote",
        args: [id, support]
    });

    log.info("Submitting vote…");
    return sendUserOp(ctx.smartClient, DAO_CONTRACT_ADDRESS, data, 0n, log);
}
