import { network } from "hardhat";

export async function timeTravel(seconds:number) {
    await network.provider.send("evm_increaseTime", [seconds]);
    await network.provider.send("evm_mine");
}