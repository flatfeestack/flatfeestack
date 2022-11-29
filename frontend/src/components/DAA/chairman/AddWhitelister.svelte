<script lang="ts">
  import type { Signer } from "ethers";

  import {
    chairmanAddress,
    membershipContract,
    whitelisters,
  } from "../../../ts/daaStore";
  import Spinner from "../../Spinner.svelte";

  let isLoading = true;
  let nonWhitelisters: Signer[] | null = null;
  let toBeAdded: string;

  $: {
    if (
      $chairmanAddress === null ||
      $membershipContract === null ||
      $whitelisters === null
    ) {
      isLoading = true;
    } else if (nonWhitelisters === null) {
      prepareView();
    }
  }

  async function prepareView() {
    const membersLength = await $membershipContract.getMembersLength();

    const allMembers = await Promise.all(
      [...Array(membersLength.toNumber()).keys()].map(async (index: Number) => {
        return await $membershipContract.members(index);
      })
    );

    nonWhitelisters = allMembers.filter(
      (address) =>
        !$whitelisters.some((whitelister) => whitelister == address) &&
        $chairmanAddress != address
    );

    isLoading = false;
  }

  async function addWhitelister() {
    await $membershipContract.addWhitelister(toBeAdded);
  }
</script>

<h2 class="text-secondary-900">Add whitelister</h2>

{#if isLoading}
  <Spinner />
{:else if nonWhitelisters.length === 0}
  <p>Anybody that could be a whitelister, is a whitelister.</p>
{:else}
  <div class="container-col2 my-2">
    <label for="toBeAddedWhiteLister">Affected member</label>
  </div>

  <div class="container-col2 my-2">
    <select
      name="toBeAddedWhiteLister"
      id="toBeAddedWhiteLister"
      bind:value={toBeAdded}
    >
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
