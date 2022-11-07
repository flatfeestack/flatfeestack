<script lang="ts">
  import type { ProposalType } from "../../types/daa";
  import Navigation from "./Navigation.svelte";
  import RequestFunds from "./proposals/RequestFunds.svelte";
  import { Editor } from "bytemd";
  import ChangeRepresentative from "./proposals/ChangeRepresentative.svelte";

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
  let transferCallData: string = "";

  function handleDescriptionChange(e) {
    description = e.detail.value;
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
</Navigation>
