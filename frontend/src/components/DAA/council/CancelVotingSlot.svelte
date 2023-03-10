<script lang="ts">
  import { daaContract } from "../../../ts/daaStore";
  import { votingSlots } from "../../../ts/proposalStore";
  import Spinner from "../../Spinner.svelte";

  let isLoading = true;
  let reason = "";
  let toBeRemoved: 0;

  $: {
    if ($votingSlots === null) {
      isLoading = true;
    } else {
      isLoading = false;
    }
  }

  async function cancelVotingSlot() {
    await $daaContract.cancelVotingSlot(toBeRemoved, reason);
  }
</script>

<h2 class="text-secondary-900">Cancel voting slot</h2>

{#if isLoading}
  <Spinner />
{:else}
  <p>Voting slots can be cancelled max. 24 hours before the voting starts.</p>
  <p>Assigned proposals will be moved to the next available voting slot.</p>

  <div class="container-col2 my-2">
    <label for="votingSlotToBeRemoved">Affected voting slot</label>
  </div>

  <div class="container-col2 my-2">
    <select
      name="votingSlotToBeRemoved"
      id="votingSlotToBeRemoved"
      bind:value={toBeRemoved}
    >
      {#each $votingSlots as votingSlot}
        <option value={votingSlot}>
          {votingSlot}
        </option>
      {/each}
    </select>
  </div>

  <div class="container-col2 my-2">
    <label for="cancellationReason">Reason</label>
  </div>

  <div class="container-col2 my-2">
    <textarea
      class="box-sizing-border"
      name="cancellationReason"
      id="cancellationReason"
      bind:value={reason}
      rows="10"
      cols="50"
    />
  </div>

  <div class="container-col2 my-2">
    <button
      class="button4"
      on:click={() => cancelVotingSlot()}
      disabled={toBeRemoved === 0 || reason === ""}>Cancel voting slot</button
    >
  </div>
{/if}
