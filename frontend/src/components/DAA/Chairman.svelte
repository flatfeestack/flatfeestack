<script lang="ts">
  import type { Signer } from "ethers";
  import { navigate } from "svelte-routing";
  import {
    chairmanAddress,
    daaContract,
    membershipContract,
    userEthereumAddress,
    whitelisters,
  } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import Navigation from "./Navigation.svelte";

  let minimumWhitelister = 0;
  let toBeAdded = "";
  let toBeRemoved = "";

  let nonWhitelisters: Signer[] = [];

  $: {
    if (
      $daaContract === null ||
      $membershipContract === null ||
      $whitelisters === null
    ) {
      $isSubmitting = true;
    } else if ($chairmanAddress !== $userEthereumAddress) {
      $error = "You are not allowed to review this page.";
      navigate("/daa/votes");
    } else if (nonWhitelisters.length === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    minimumWhitelister = minimumWhitelister = (
      await $membershipContract.minimumWhitelister()
    ).toNumber();

    await setMembers();

    $isSubmitting = false;
  }

  async function setMembers() {
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
  }

  async function addWhitelister() {
    await $membershipContract.addWhitelister(toBeAdded);
  }

  async function removeWhitelister() {
    await $membershipContract.removeWhitelister(toBeRemoved);
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Chairman functions</h1>

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

  <h2 class="text-secondary-900">Remove whitelister</h2>

  {#if $whitelisters.length - 1 < minimumWhitelister}
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
</Navigation>
