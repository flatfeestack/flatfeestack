<script lang="ts">
  import {
    faCheck,
    faQuestion,
    faXmark,
  } from "@fortawesome/free-solid-svg-icons";
  import { Viewer } from "bytemd";
  import type { Event } from "ethers";
  import Fa from "svelte-fa";
  import { navigate } from "svelte-routing";
  import { daaContract, userEthereumAddress } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import { proposalCreatedEvents, votingSlots } from "../../ts/proposalStore";
  import Navigation from "./Navigation.svelte";

  interface VoteValues {
    canVote: boolean;
    reason: string;
    value: number;
  }

  interface VoteValuesContainer {
    [key: string]: VoteValues;
  }

  export let blockNumber: Number;
  let proposals = [];
  let voteValues: VoteValuesContainer = {};
  let hasAnyVotes = false;

  $: {
    if ($proposalCreatedEvents === null || $votingSlots === null) {
      $isSubmitting = true;
    } else if (proposals.length === 0) {
      $isSubmitting = true;
      prepareView();
    }
  }

  async function prepareView() {
    if (!$votingSlots.includes(blockNumber)) {
      $error = "Invalid voting slot.";
      navigate("/daa/votes");
    }

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

    const votesCasted = await $daaContract.queryFilter(
      $daaContract.filters.VoteCast(
        $userEthereumAddress,
        null,
        null,
        null,
        null
      )
    );

    proposals.forEach((proposal) => {
      const event: Event | undefined = votesCasted.find(
        (event) => event.args[1].toString() === proposal.id
      );

      if (event === undefined) {
        voteValues = {
          ...voteValues,
          [proposal.id]: { value: 0, reason: "", canVote: true },
        };
        hasAnyVotes = true;
      } else {
        voteValues = {
          ...voteValues,
          [proposal.id]: {
            value: event.args[2],
            reason: event.args[4],
            canVote: false,
          },
        };
      }
    });

    $isSubmitting = false;
  }

  function handleVoteValue(proposalId: string, voteValue: number) {
    voteValues[proposalId].value = voteValue;
  }

  async function castVotes() {
    for (const [key, value] of Object.entries(voteValues)) {
      if (value.reason.trim() === "") {
        await $daaContract.castVote(key, value.value);
      } else {
        await $daaContract.castVoteWithReason(key, value.value, value.reason);
      }
    }
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
          disabled={!voteValues[proposal.id].canVote}
          class={voteValues[proposal.id].value === 0 ? "button1" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 0)}
        >
          <Fa icon={faXmark} size="sm" class="icon px-2" />
        </button>

        <button
          disabled={!voteValues[proposal.id].canVote}
          class={voteValues[proposal.id].value === 1 ? "button1" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 1)}
        >
          <Fa icon={faCheck} size="sm" class="icon px-2" />
        </button>

        <button
          disabled={!voteValues[proposal.id].canVote}
          class={voteValues[proposal.id].value === 2 ? "button1" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 2)}
        >
          <Fa icon={faQuestion} size="sm" class="icon px-2" />
        </button>
      </div>
    </div>

    {#if voteValues[proposal.id].canVote}
      <p>Reason (optional):</p>

      <div>
        <textarea
          class="box-sizing-border"
          bind:value={voteValues[proposal.id].reason}
          rows="10"
          cols="50"
        />
      </div>
    {:else if voteValues[proposal.id].reason.trim() == ""}
      <p>Reason: (no reason given)</p>
    {:else}
      <p>Reason: {voteValues[proposal.id].reason}</p>
    {/if}
  {/each}

  <button disabled={!hasAnyVotes} on:click={() => castVotes()} class="button1"
    >Cast votes</button
  >
</Navigation>
