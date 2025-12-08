import type { Address, Abi } from "viem";
import { encodeFunctionData, formatEther } from "viem";
import { readContract } from "../viem/read";
import { sendUserOp } from "../viem/write";
import FlatFeeStackNFT from "../../../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackNFT.json";
import { getDaoTokenAddress, ensureSmartAccountHasFunds } from "./utils";
import type { AAContext } from "./context";

const NFT_ABI = FlatFeeStackNFT.abi as Abi;

export async function listMembershipTokens(ctx: AAContext, owner: Address) {
    const nft = await getDaoTokenAddress(ctx);
    const balance = await readContract<bigint>(ctx.public, nft, NFT_ABI, "balanceOf", [owner]);

    const tokens = [];

    for (let i = 0n; i < balance; i++) {
        const tokenId = await readContract<bigint>(
            ctx.public,
            nft,
            NFT_ABI,
            "tokenOfOwnerByIndex",
            [owner, i]
        );

        const paid = await readContract<bigint>(
            ctx.public,
            nft,
            NFT_ABI,
            "membershipPayed",
            [tokenId]
        );

        tokens.push({ tokenId, membershipPaidUntil: paid });
    }

    return tokens;
}

export async function renewMembership(ctx: AAContext, tokenId: bigint, log: any) {
    const nft = await getDaoTokenAddress(ctx);
    const fee = await readContract<bigint>(ctx.public, nft, NFT_ABI, "membershipFee");
    
    log?.info(`Membership fee: ${formatEther(fee)} ETH`);

    await ensureSmartAccountHasFunds(ctx, fee, log);

    const data = encodeFunctionData({
        abi: NFT_ABI,
        functionName: "payMembership",
        args: [tokenId]
    });

    return await sendUserOp(ctx.smartClient, nft, data, fee, log);
}