<script lang="ts">
  import {
    faCheck,
    faQuestion,
    faXmark,
  } from "@fortawesome/free-solid-svg-icons";
  import { Viewer } from "bytemd";
  import Fa from "svelte-fa";
  import { daaContract } from "../../ts/daaStore";
  import { isSubmitting } from "../../ts/mainStore";
  import { proposalCreatedEvents, votingSlots } from "../../ts/proposalStore";
  import Navigation from "./Navigation.svelte";

  export let blockNumber: Number;
  let proposals = [];
  let voteValues = {};

  $: {
    if ($proposalCreatedEvents === null || $votingSlots === null) {
      $isSubmitting = true;
    } else if (proposals.length === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    const amountOfProposals =
      await $daaContract.getNumberOfProposalsInVotingSlot(blockNumber);

    proposals = await Promise.all(
      [...Array(amountOfProposals.toNumber()).keys()].map(
        async (index: Number) => {
          const proposalId = (
            await $daaContract.votingSlots(blockNumber, index)
          ).toString();

          const event = $proposalCreatedEvents.find(
            (event) => event.args[0].toString() === proposalId
          );

          return {
            description: event.args[8],
            id: proposalId,
            proposer: event.args[1],
          };
        }
      )
    );

    proposals.forEach((proposal) => {
      voteValues = { ...voteValues, [proposal.id]: { value: 0, reason: "" } };
    });

    $isSubmitting = false;
  }

  function handleVoteValue(proposalId: string, voteValue: number) {
    voteValues[proposalId].value = voteValue;
  }
</script>

<style>
  .vote-container {
    align-items: center;
    display: flex;
    justify-content: space-between;
  }

  textarea {
    width: 100%;
  }
</style>

<Navigation>
  <h1 class="text-secondary-900">Cast votes</h1>

  {#each proposals as proposal, i}
    <h2>Proposal {i + 1}</h2>
    Proposer: {proposal.proposer}

    <Viewer value={proposal.description} />

    <div class="vote-container">
      <p>Your vote:</p>
      <div>
        <button
          class={voteValues[proposal.id].value === 0 ? "button1" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 0)}
        >
          <Fa icon={faXmark} size="sm" class="icon px-2" />
        </button>

        <button
          class={voteValues[proposal.id].value === 1 ? "button1" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 1)}
        >
          <Fa icon={faCheck} size="sm" class="icon px-2" />
        </button>

        <button
          class={voteValues[proposal.id].value === 2 ? "button1" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 2)}
        >
          <Fa icon={faQuestion} size="sm" class="icon px-2" />
        </button>
      </div>
    </div>

    <p>Reason (optional):</p>

    <div>
      <textarea
        class="box-sizing-border"
        bind:value={voteValues[proposal.id].reason}
        rows="10"
        cols="50"
      />
    </div>
  {/each}
</Navigation>
