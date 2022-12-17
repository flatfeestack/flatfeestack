import { get } from "svelte/store";
import { currentBlockTimestamp } from "../ts/daaStore";
import formateDateTime from "./formatDateTime";

const secondsPerBlock = 12;

function futureBlockDate(
  futureBlockNumber: number,
  currentBlockNumber: number
): string {
  const blockDifference = futureBlockNumber - currentBlockNumber;
  const timeDifference = Math.abs(blockDifference * secondsPerBlock);

  const date = new Date(get(currentBlockTimestamp) * 1000);
  date.setSeconds(date.getSeconds() + timeDifference);

  return formateDateTime(date);
}

export { futureBlockDate, secondsPerBlock };
