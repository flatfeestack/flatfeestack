<script lang="ts">
  import {
    membershipContract,
    provider,
    whitelisters,
  } from "../../../ts/daaStore";
  import Spinner from "../../Spinner.svelte";

  let currentBlockNumber = 0;
  let isLoading = true;
  let minimumWhitelister = 0;
  let toBeRemoved: string;

  $: {
    if ($membershipContract === null || $whitelisters === null) {
      isLoading = true;
    } else if (currentBlockNumber === 0 || minimumWhitelister === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    minimumWhitelister = minimumWhitelister = (
      await $membershipContract.minimumWhitelister()
    ).toNumber();
    currentBlockNumber = await $provider.getBlockNumber();

    isLoading = false;
  }

  async function removeWhitelister() {
    await $membershipContract.removeWhitelister(toBeRemoved);
  }
</script>

<h2 class="text-secondary-900">Remove whitelister</h2>

{#if isLoading}
  <Spinner />
{:else if $whitelisters.length - 1 < minimumWhitelister}
  <p>
    You cannot remove any whitelister as the minimum is {minimumWhitelister} (currently
    {$whitelisters.length}).
  </p>
{:else}
  <div class="container-col2 my-2">
    <label for="toBeRemoved">Affected whitelister</label>
  </div>

  <div class="container-col2 my-2">
    <select name="toBeRemoved" bind:value={toBeRemoved}>
      {#each $whitelisters as whitelister}
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
