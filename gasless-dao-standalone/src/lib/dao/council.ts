import type { AAContext } from "../aa/context";
import { encodePacked, keccak256, encodeFunctionData, type Abi } from "viem";
import FlatFeeStackDAO from "../../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackDAO.json";
import FlatFeeStackNFT from "../../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackNFT.json";
import { DAO_CONTRACT_ADDRESS, NFT_CONTRACT_ADDRESS } from "../../config";

const DAO_ABI = FlatFeeStackDAO.abi as Abi;
const NFT_ABI = FlatFeeStackNFT.abi as Abi;

export async function prepareVotingDelaySignature(
    ctx: AAContext,
    delay: bigint
) {
    const packed = encodePacked(
        ["address", "string", "uint256"],
        [DAO_CONTRACT_ADDRESS, "setVotingDelay", delay]
    );

    const hash = keccak256(packed);

    const signature = await (window as any).ethereum.request({
        method: "personal_sign",
        params: [hash, ctx.eoa]
    });

    return { hash, signature };
}

export async function executeVotingDelay(
    ctx: AAContext,
    delay: bigint,
    signature2: `0x${string}`,
    log?: any
) {
    const data = encodeFunctionData({
        abi: DAO_ABI,
        functionName: "setCouncilVotingDelayOverride",
        args: [delay, 0n, 0n, signature2]
    });

    log?.info?.("Sending council override transaction...");

    const { hash, receipt } = await ctx.smartClient.sendUserOperation({
        calls: [{ to: DAO_CONTRACT_ADDRESS, data }]
    });

    return { hash, receipt };
}

export async function prepareMintSignature(
    ctx: AAContext,
    newMember: `0x${string}`,
    tokenId: bigint
) {
    const payloadHash = keccak256(
        encodePacked(
            ["address", "string", "address", "string", "uint256"],
            [NFT_CONTRACT_ADDRESS, "safeMint", newMember, "#", tokenId]
        )
    );

    const signature = await (window as any).ethereum.request({
        method: "personal_sign",
        params: [payloadHash, ctx.eoa]
    });

    return { payloadHash, signature };
}

export async function executeMint(
    ctx: AAContext,
    newMember: `0x${string}`,
    tokenId: bigint,
    sig1: `0x${string}`,
    sig2: `0x${string}`,
    log?: any
) {
    const data = encodeFunctionData({
        abi: NFT_ABI,
        functionName: "safeMint",
        args: [newMember, 1n, sig1, 2n, sig2]
    });

    log?.info?.("Executing council mint...");

    return ctx.smartClient.sendUserOperation({
        calls: [{ to: NFT_CONTRACT_ADDRESS, data, value: 0n }]
    });
}
