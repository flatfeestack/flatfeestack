import { keccak256, toUtf8Bytes } from "ethers/lib/utils";
import { get } from "svelte/store";
import { daoContract } from "$lib/ts/daoStore";

export async function queueProposal(
  targets: string[],
  values: number[],
  description: string,
  calldatas: string[]
) {
  await get(daoContract).queue(
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
  await get(daoContract).execute(
    targets,
    values,
    calldatas,
    keccak256(toUtf8Bytes(description))
  );
}
