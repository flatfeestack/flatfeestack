import type { Abi } from "viem";
import { encodeFunctionData } from "viem";
import { sendUserOp } from "../viem/write";
import type { AAContext } from "./context";
import FlatFeeStackDAO from "../../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackDAO.json";
import { DAO_CONTRACT_ADDRESS } from "../../config";

const DAO_ABI = FlatFeeStackDAO.abi as Abi;

export async function createProposal(
    ctx: AAContext,
    proposal: {
        targets: string[];
        values: bigint[];
        calldatas: string[];
        description: string;
    },
    log: any
) {
    if (!ctx.isMember) {
        throw new Error("Not a DAO member");
    }

    const data = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "propose",
        args: [
            proposal.targets,
            proposal.values,
            proposal.calldatas,
            proposal.description
        ]
    });

    log.info("Submitting proposal via smart account…");

    return sendUserOp(ctx.smartClient, DAO_CONTRACT_ADDRESS, data, 0n, log);
}
