import { get } from "svelte/store";
import { daoContract } from "../ts/daoStore";
import { Result, keccak256, toUtf8Bytes } from "ethers";

export async function queueProposal(
  targets: Result,
  values: Result,
  description: string,
  calldatas: Result
) {
  await get(daoContract).queue(
    targets.toArray(),
    values.toArray(),
    calldatas.toArray(),
    keccak256(toUtf8Bytes(description))
  );
}

export async function executeProposal(
  targets: Result,
  values: Result,
  description: string,
  calldatas: Result
) {
  await get(daoContract).execute(
    targets.toArray(),
    values.toArray(),
    calldatas.toArray(),
    keccak256(toUtf8Bytes(description))
  );
}
