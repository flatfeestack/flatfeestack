import { provider } from "../ts/daaStore";
import { get } from "svelte/store";
import formateDateTime from "./formatDateTime";

const secondsPerBlock = 12;

async function futureBlockDate(
  futureBlockNumber: number,
  currentBlockNumber: number
): Promise<string> {
  const blockDifference = futureBlockNumber - currentBlockNumber;
  const timeDifference = Math.abs(blockDifference * secondsPerBlock);

  const currentBlockTimestamp = (
    await get(provider).getBlock(currentBlockNumber)
  ).timestamp;
  const date = new Date(currentBlockTimestamp * 1000);
  date.setSeconds(date.getSeconds() + timeDifference);

  return formateDateTime(date);
}

export { futureBlockDate, secondsPerBlock };
