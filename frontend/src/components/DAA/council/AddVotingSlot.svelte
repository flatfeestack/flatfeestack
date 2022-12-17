<script lang="ts">
  import { daaContract, provider } from "../../../ts/daaStore";
  import {
    futureBlockDate,
    secondsPerBlock,
  } from "../../../utils/futureBlockDate";
  import Spinner from "../../Spinner.svelte";

  let isLoading = true;
  let currentBlockNumber = 0;
  let plannedBlockNumber = 0;
  let minValue = 0;

  $: {
    if ($provider === null || $daaContract === null) {
      isLoading = true;
    } else if (currentBlockNumber === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    currentBlockNumber = await $provider.getBlockNumber();
    isLoading = false;
    minValue = currentBlockNumber + (60 * 60 * 24 * 7 * 4) / secondsPerBlock;
  }

  async function createVotingSlot() {
    await $daaContract.setVotingSlot(plannedBlockNumber);
  }
</script>

<h2 class="text-secondary-900">Add voting slot</h2>

{#if isLoading}
  <Spinner />
{:else}
  <p>
    The current block number is {currentBlockNumber}, voting slots need to be
    announced one month in advance, so the minimum value is {minValue} (approx. {futureBlockDate(
      minValue,
      currentBlockNumber
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
      class="button1"
      on:click={() => createVotingSlot()}
      disabled={plannedBlockNumber === 0}>Create voting slot</button
    >
  </div>
{/if}
