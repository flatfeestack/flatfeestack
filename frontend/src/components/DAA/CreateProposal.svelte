<script lang="ts">
  import { onMount } from "svelte";
  import { navigate } from "svelte-routing";
  import {
    daaContract,
    membershipContract,
    membershipStatusValue,
    signer,
  } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import type { Call, ProposalType } from "../../types/daa";
  import Navigation from "./Navigation.svelte";
  import AddCouncilMember from "./proposals/AddCouncilMember.svelte";
  import CallExtraOrdinaryAssembly from "./proposals/CallExtraOrdinaryAssembly.svelte";
  import DissolveAssociation from "./proposals/DissolveAssociation.svelte";
  import FreeText from "./proposals/FreeText.svelte";
  import RemoveCouncilMember from "./proposals/RemoveCouncilMember.svelte";
  import RemoveMember from "./proposals/RemoveMember.svelte";
  import RequestFunds from "./proposals/RequestFunds.svelte";

  const proposalTypes: ProposalType[] = [
    {
      component: RequestFunds,
      text: "Request funds",
    },
    {
      component: AddCouncilMember,
      text: "Add council member",
    },
    {
      component: RemoveCouncilMember,
      text: "Remove council member",
    },
    {
      component: RemoveMember,
      text: "Remove member",
    },
    {
      component: CallExtraOrdinaryAssembly,
      text: "Call extra ordinary assembly",
    },
    {
      component: DissolveAssociation,
      text: "Dissolve association",
    },
    {
      component: FreeText,
      text: "Free text",
    },
  ];

  let selected = 0;

  let calls: Call[] = [];
  let description = "";

  onMount(async () => {
    if ($signer === null || $membershipContract === null) {
      moveToVotesPage();
      return;
    }

    if ($membershipStatusValue != 3) {
      moveToVotesPage();
    }
  });

  function moveToVotesPage() {
    $error = "You are not allowed to view this page.";
    navigate("/daa/home");
  }

  async function createProposal() {
    $isSubmitting = true;

    let targets = [];
    let values = [];
    let transferCallData = [];

    calls.forEach((call) => {
      targets.push(call.target);
      values.push(call.value);
      transferCallData.push(call.transferCallData);
    });

    await $daaContract["propose(address[],uint256[],bytes[],string)"](
      targets,
      values,
      transferCallData,
      description
    );

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

  textarea {
    width: 100%;
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

    <svelte:component this={proposalTypes[selected].component} bind:calls />
  </div>

  <p>Description</p>

  <textarea class="box-sizing-border" bind:value={description} rows="10" />

  <button class="button4" on:click={() => createProposal()}
    >Create proposal</button
  >
</Navigation>
