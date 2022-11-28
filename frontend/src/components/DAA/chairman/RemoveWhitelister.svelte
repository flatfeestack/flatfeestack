<script lang="ts">
  import type { Contract, Signer } from "ethers";

  export let membershipContract: Contract;
  export let whitelisters: Signer[];
  export let minimumWhitelister: number;

  let toBeRemoved: string;

  async function removeWhitelister() {
    await membershipContract.removeWhitelister(toBeRemoved);
  }
</script>

<h2 class="text-secondary-900">Remove whitelister</h2>

{#if whitelisters.length - 1 < minimumWhitelister}
  <p>
    You cannot remove any whitelister as the minimum is {minimumWhitelister} (currently
    {whitelisters.length}).
  </p>
{:else}
  <div class="container-col2 my-2">
    <label for="toBeRemoved">Affected whitelister</label>
  </div>

  <div class="container-col2 my-2">
    <select name="toBeRemoved" bind:value={toBeRemoved}>
      {#each whitelisters as whitelister}
        <option value={whitelister}>
          {whitelister}
        </option>
      {/each}
    </select>
  </div>

  <div class="container-col2 my-2">
    <button
      class="button1"
      on:click={() => removeWhitelister()}
      disabled={toBeRemoved === ""}>Remove whitelister</button
    >
  </div>
{/if}
