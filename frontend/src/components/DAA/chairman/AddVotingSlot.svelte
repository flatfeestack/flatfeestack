<script lang="ts">
  import type { Contract } from "ethers";
  import {
    futureBlockDate,
    secondsPerBlock,
  } from "../../../utils/futureBlockDate";

  let blockNumber: number = 0;
  export let currentBlockNumber: number;
  export let daaContract: Contract;

  const minValue =
    currentBlockNumber + (60 * 60 * 24 * 7 * 4) / secondsPerBlock;

  async function createVotingSlot() {
    await daaContract.setVotingSlot(blockNumber);
  }
</script>

<h2 class="text-secondary-900">Add voting slot</h2>

<p>
  For context: The current block number is {currentBlockNumber}, voting slots
  need to be announced one month in advance, so the minimum value is {minValue}
  {#await futureBlockDate(minValue, currentBlockNumber) then futureDate}
    (approx .{futureDate})
  {/await}
</p>

<div class="container-col2 my-2">
  <label for="blockNumber">Voting should start at block number</label>
</div>

<div class="container-col2 my-2">
  <input
    type="number"
    name="blockNumber"
    min={minValue}
    bind:value={blockNumber}
  />
</div>

{#if blockNumber !== 0}
  {#await futureBlockDate(blockNumber, currentBlockNumber) then futureDate}
    <p>The voting for this block would start approx. at {futureDate}</p>
  {/await}
{/if}

<div class="container-col2 my-2">
  <button
    class="button1"
    on:click={() => createVotingSlot()}
    disabled={blockNumber === 0}>Create voting slot</button
  >
</div>
