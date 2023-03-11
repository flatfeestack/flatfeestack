import { keccak256, toUtf8Bytes } from "ethers/lib/utils";
import { get } from "svelte/store";
import { daaContract } from "../ts/daaStore";

export async function queueProposal(
  targets: string[],
  values: number[],
  description: string,
  calldatas: string[]
) {
  await get(daaContract).queue(
    targets,
    values,
    calldatas,
    keccak256(toUtf8Bytes(description))
  );
}

export async function executeProposal(
  targets: string[],
  values: number[],
  description: string,
  calldatas: string[]
) {
  await get(daaContract).execute(
    targets,
    values,
    calldatas,
    keccak256(toUtf8Bytes(description))
  );
}
