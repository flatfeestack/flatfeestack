import { get } from "svelte/store";
import { currentBlockNumber, currentBlockTimestamp } from "../ts/daoStore";
import formatDateTime from "./formatDateTime";

const secondsPerBlock = 12;

function futureBlockDate(futureBlockNumber: number): string {
  const blockDifference = futureBlockNumber - get(currentBlockNumber);
  const timeDifference = Math.abs(blockDifference * secondsPerBlock);

  const date = new Date(get(currentBlockTimestamp) * 1000);
  date.setSeconds(date.getSeconds() + timeDifference);

  return formatDateTime(date);
}

export { futureBlockDate, secondsPerBlock };
