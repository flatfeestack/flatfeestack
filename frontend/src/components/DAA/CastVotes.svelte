<script lang="ts">
  import {
    faCheck,
    faQuestion,
    faXmark,
  } from "@fortawesome/free-solid-svg-icons";
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

  interface Votes {
    abstainVotes: number;
    againstVotes: number;
    forVotes: number;
  }

  interface VoteValuesContainer {
    [key: string]: VoteValues;
  }

  interface VotesContainer {
    [key: string]: Votes;
  }

  export let blockNumber: string;
  let proposals = [];
  let voteValues: VoteValuesContainer = {};
  let votes: VotesContainer = {};
  let hasAnyVotes = false;

  $: {
    if ($votingSlots === null) {
      $isSubmitting = true;
    } else if (proposals.length === 0) {
      $isSubmitting = true;
      prepareView();
    }
  }

  async function prepareView() {
    if (!$votingSlots.includes(Number(blockNumber))) {
      $error = "Invalid voting slot.";
      navigate("/daa/votes");
    }

    const votingPower = await $daaContract.getVotes(
      $userEthereumAddress,
      blockNumber
    );
    if (votingPower.toNumber() < 1) {
      $error = "You are not allowed to vote in this cycle.";
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

          const event = await proposalCreatedEvents.get(
            proposalId,
            $daaContract
          );

          return {
            description: event.event.args[8],
            id: event.proposalId,
            proposer: event.event.args[1],
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

    for (const proposal of proposals) {
      const { againstVotes, forVotes, abstainVotes } =
        await $daaContract.proposalVotes(proposal.id);
      votes = {
        ...votes,
        [proposal.id]: {
          abstainVotes,
          againstVotes,
          forVotes,
        },
      };
    }

    proposals.forEach((proposal) => {
      const event: Event | undefined = votesCasted.find(
        (event) => event.args[1].toString() === proposal.id
      );

      if (event === undefined) {
        voteValues = {
          ...voteValues,
          [proposal.id]: { value: -1, reason: "", canVote: true },
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
      if (!value.canVote || value.value === -1) {
        continue;
      }

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

    <p>Proposer: {proposal.proposer}</p>

    <p>{proposal.description}</p>

    <div>
      <p>State of the vote:</p>
      <ul>
        <li>For votes: {votes[proposal.id].forVotes}</li>
        <li>Against votes: {votes[proposal.id].againstVotes}</li>
        <li>Abstain votes: {votes[proposal.id].abstainVotes}</li>
      </ul>
    </div>

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
