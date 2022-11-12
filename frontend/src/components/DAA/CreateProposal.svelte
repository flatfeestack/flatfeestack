<script lang="ts">
  import { Editor } from "bytemd";
  import { onMount } from "svelte";
  import { navigate } from "svelte-routing";
  import {
    daaContract,
    membershipContract,
    membershipStatusValue,
    signer,
  } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import type { ProposalType } from "../../types/daa";
  import Navigation from "./Navigation.svelte";
  import ChangeRepresentative from "./proposals/ChangeRepresentative.svelte";
  import RequestFunds from "./proposals/RequestFunds.svelte";

  const proposalTypes: ProposalType[] = [
    {
      component: RequestFunds,
      text: "Request funds",
    },
    {
      component: ChangeRepresentative,
      text: "Change representative",
    },
  ];

  let selected = 0;

  let targets: string[] = [];
  let values: number[] = [];
  let description = "";
  let transferCallData: string[] = [];

  onMount(async () => {
    if ($signer === null || $membershipContract === null) {
      moveToVotesPage();
      return;
    }

    if ($membershipStatusValue != 3) {
      moveToVotesPage();
    }
  });

  function handleDescriptionChange(e) {
    description = e.detail.value;
  }

  function moveToVotesPage() {
    $error = "You are not allowed to review this page.";
    navigate("/daa/votes");
  }

  async function createProposal() {
    $isSubmitting = true;

    await $daaContract.propose(targets, values, transferCallData, description);

    $isSubmitting = false;
    navigate("/daa/votes");
  }
</script>

<style>
  h1 {
    color: var(--secondary-900) !important;
  }

  .wrapper {
    display: grid;
    grid-template-columns: 1fr 1fr;
    row-gap: 0.5em;
  }
</style>

<Navigation>
  <h1 class="text-secondary-900">Create a proposal</h1>

  <div class="wrapper">
    <label for="proposalType">Proposal type</label>
    <select bind:value={selected} name="proposalType" required>
      {#each proposalTypes as proposalType, index}
        <option value={index}>
          {proposalType.text}
        </option>
      {/each}
    </select>

    <svelte:component
      this={proposalTypes[selected].component}
      bind:targets
      bind:values
      bind:transferCallData
    />
  </div>

  <p>Description</p>
  <Editor value={description} on:change={handleDescriptionChange} />

  <button class="button1" on:click={() => createProposal()}
    >Create proposal</button
  >
</Navigation>
