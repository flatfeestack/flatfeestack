<script lang="ts">
  import { currentBlockNumber, daoContract } from "../../../ts/daoStore";
  import { provider } from "../../../ts/ethStore";
  import {
    futureBlockDate,
    secondsPerBlock,
  } from "../../../utils/futureBlockDate";
  import Spinner from "../../Spinner.svelte";

  let isLoading = true;
  let plannedBlockNumber = 0;
  let minValue = 0;

  $: {
    if ($provider === null || $daoContract === null) {
      isLoading = true;
    } else if (minValue === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    isLoading = false;
    minValue = $currentBlockNumber + (60 * 60 * 24 * 7 * 4) / secondsPerBlock;
  }

  async function createVotingSlot() {
    await $daoContract.setVotingSlot(plannedBlockNumber);
  }
</script>

<h2 class="text-secondary-900">Add voting slot</h2>

{#if isLoading}
  <Spinner />
{:else}
  <p>
    The current block number is {currentBlockNumber}, voting slots need to be
    announced one month in advance, so the minimum value is {minValue} (approx. {futureBlockDate(
      minValue
    )}).
  </p>

  <div class="container-col2 my-2">
    <label for="votingSlotBlockNumber"
      >Voting should start at block number</label
    >
  </div>

  <div class="container-col2 my-2">
    <input
      type="number"
      name="votingSlotBlockNumber"
      id="votingSlotBlockNumber"
      min={minValue}
      bind:value={plannedBlockNumber}
    />
  </div>

  {#if plannedBlockNumber !== 0}
    <p>
      The voting for this block would start approx. at {futureBlockDate(
        plannedBlockNumber,
        currentBlockNumber
      )}
    </p>
  {/if}

  <div class="container-col2 my-2">
    <button
      class="button4"
      on:click={() => createVotingSlot()}
      disabled={plannedBlockNumber === 0}>Create voting slot</button
    >
  </div>
{/if}
