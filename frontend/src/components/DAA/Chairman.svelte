<script lang="ts">
  import type { Signer } from "ethers";
  import { daaContract, membershipContract } from "../../ts/daaStore";
  import { isSubmitting } from "../../ts/mainStore";
  import Navigation from "./Navigation.svelte";

  let minimumWhitelister = 0;
  let toBeAdded = "";
  let toBeRemoved = "";

  let nonWhitelisters: Signer[] = [];
  let whitelisters: Signer[] = [];

  $: {
    if ($daaContract === null || $membershipContract === null) {
      $isSubmitting = true;
    } else if (whitelisters.length === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    await setWhitelisters();
    await setMembers();

    $isSubmitting = false;
  }

  async function setMembers() {
    const [membersLength, chairman] = await Promise.all([
      $membershipContract.getMembersLength(),
      $membershipContract.chairman(),
    ]);

    const allMembers = await Promise.all(
      [...Array(membersLength.toNumber()).keys()].map(async (index: Number) => {
        return await $membershipContract.members(index);
      })
    );

    nonWhitelisters = allMembers.filter(
      (address) =>
        !whitelisters.some((whitelister) => whitelister == address) &&
        chairman != address
    );
  }

  async function setWhitelisters() {
    const whitelisterLength = await $membershipContract.whitelisterListLength();
    minimumWhitelister = (
      await $membershipContract.minimumWhitelister()
    ).toNumber();

    whitelisters = await Promise.all(
      [...Array(whitelisterLength.toNumber()).keys()].map(
        async (index: Number) => {
          return await $membershipContract.whitelisterList(index);
        }
      )
    );
  }

  async function addWhitelister() {
    await $membershipContract.addWhitelister(toBeAdded);
  }

  async function removeWhitelister() {
    await $membershipContract.removeWhitelister(toBeRemoved);
  }
</script>

<style>
  .container {
    display: flex;
    justify-content: space-between;
  }
</style>

<Navigation>
  <h1 class="text-secondary-900">Chairman functions</h1>

  <h2 class="text-secondary-900">Add whitelister</h2>

  {#if nonWhitelisters.length === 0}
    <p>Anybody that could be a whitelister, is a whitelister.</p>
  {:else}
    <div class="flex justify-between">
      <div>
        <div class="block">
          <label for="toBeAdded">Affected member</label>
        </div>

        <div class="block">
          <select name="toBeAdded" bind:value={toBeAdded}>
            {#each nonWhitelisters as nonWhitelister}
              <option value={nonWhitelister}>
                {nonWhitelister}
              </option>
            {/each}
          </select>
        </div>
      </div>

      <button
        class="button1"
        on:click={() => addWhitelister()}
        disabled={toBeAdded === ""}>Add whitelister</button
      >
    </div>
  {/if}

  <h2 class="text-secondary-900">Remove whitelister</h2>

  {#if whitelisters.length - 1 < minimumWhitelister}
    <p>
      You cannot remove any whitelister as the minimum is {minimumWhitelister} (currently
      {whitelisters.length}).
    </p>
  {:else}
    <div class="flex justify-between">
      <div>
        <div class="block">
          <label for="toBeRemoved">Affected whitelister</label>
        </div>

        <div class="block">
          <select name="toBeRemoved" bind:value={toBeRemoved}>
            {#each whitelisters as whitelister}
              <option value={whitelister}>
                {whitelister}
              </option>
            {/each}
          </select>
        </div>
      </div>

      <button
        class="button1"
        on:click={() => removeWhitelister()}
        disabled={toBeRemoved === ""}>Remove whitelister</button
      >
    </div>
  {/if}
</Navigation>
