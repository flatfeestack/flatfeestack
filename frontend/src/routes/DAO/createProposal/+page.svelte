<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { goto } from "$app/navigation";
  import {
    daoContract,
    membershipContract,
    membershipStatusValue,
    signer,
  } from "../../../ts/daoStore";
  import { error, isSubmitting } from "../../../ts/mainStore";
  import type { Call, ProposalType } from "../../../types/dao";
  import AddCouncilMember from "./AddCouncilMember.svelte";
  import CallExtraOrdinaryAssembly from "./CallExtraOrdinaryAssembly.svelte";
  import DissolveAssociation from "./DissolveAssociation.svelte";
  import FreeText from "./FreeText.svelte";
  import RemoveCouncilMember from "./RemoveCouncilMember.svelte";
  import RemoveMember from "./RemoveMember.svelte";
  import RequestFunds from "./RequestFunds.svelte";
  import checkUndefinedProvider from "../../../utils/checkUndefinedProvider";

  checkUndefinedProvider();

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
    goto("/dao");
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

    await $daoContract["propose(address[],uint256[],bytes[],string)"](
      targets,
      values,
      transferCallData,
      description
    );

    $isSubmitting = false;
    goto("/dao/votes");
  }

  onDestroy(() => {
    $isSubmitting = false;
  });
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
