<script lang="ts">
  import type { Contract } from "ethers";

  export let daaContract: Contract;
  export let votingSlots: number[];

  let reason = "";
  let toBeRemoved: 0;

  async function cancelVotingSlot() {
    await daaContract.cancelVotingSlot(toBeRemoved, reason);
  }
</script>

<h2 class="text-secondary-900">Cancel voting slot</h2>

<p>Voting slots can be cancelled max. 24 hours before the voting starts.</p>
<p>Assigned proposals will be moved to the next available voting slot</p>

<div class="container-col2 my-2">
  <label for="toBeRemoved">Affected voting slot</label>
</div>

<div class="container-col2 my-2">
  <select name="toBeRemoved" bind:value={toBeRemoved}>
    {#each votingSlots as votingSlot}
      <option value={votingSlot}>
        {votingSlot}
      </option>
    {/each}
  </select>
</div>

<div class="container-col2 my-2">
  <label for="reason">Reason</label>
</div>

<div class="container-col2 my-2">
  <textarea class="box-sizing-border" bind:value={reason} rows="10" cols="50" />
</div>

<div class="container-col2 my-2">
  <button
    class="button1"
    on:click={() => cancelVotingSlot()}
    disabled={toBeRemoved === 0 || reason === ""}>Cancel voting slot</button
  >
</div>
