<script lang="ts">
  import type { Contract, Signer } from "ethers";

  export let membershipContract: Contract;
  export let nonWhitelisters: Signer[];

  let toBeAdded: string;

  async function addWhitelister() {
    await membershipContract.addWhitelister(toBeAdded);
  }
</script>

<h2 class="text-secondary-900">Add whitelister</h2>

{#if nonWhitelisters.length === 0}
  <p>Anybody that could be a whitelister, is a whitelister.</p>
{:else}
  <div class="container-col2 my-2">
    <label for="toBeAdded">Affected member</label>
  </div>

  <div class="container-col2 my-2">
    <select name="toBeAdded" bind:value={toBeAdded}>
      {#each nonWhitelisters as nonWhitelister}
        <option value={nonWhitelister}>
          {nonWhitelister}
        </option>
      {/each}
    </select>
  </div>

  <div class="container-col2 my-2">
    <button
      class="button1"
      on:click={() => addWhitelister()}
      disabled={toBeAdded === ""}>Add whitelister</button
    >
  </div>
{/if}
