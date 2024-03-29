<script lang="ts">
  import {
    faCheck,
    faQuestion,
    faXmark,
  } from "@fortawesome/free-solid-svg-icons";
  import type { EventLog } from "ethers";
  import { onDestroy, onMount } from "svelte";
  import Fa from "svelte-fa";
  import { navigate } from "svelte-routing";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import {
    daoConfig,
    daoContract,
    membershipStatusValue,
  } from "../../ts/daoStore";
  import { userEthereumAddress } from "../../ts/ethStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import { proposalStore, votingSlots } from "../../ts/proposalStore";
  import { checkUndefinedProvider } from "../../utils/ethHelpers";

  interface VoteValues {
    canVote: boolean;
    reason: string;
    value: bigint;
  }

  interface Votes {
    abstainVotes: bigint;
    againstVotes: bigint;
    forVotes: bigint;
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

  checkUndefinedProvider();

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
      navigate("/dao/votes");
    }

    const votingPower = (await $daoContract.getVotes(
      $userEthereumAddress,
      blockNumber
    )) as bigint;
    if (votingPower < 1n) {
      $error = "You are not allowed to vote in this cycle.";
      navigate("/dao/votes");
    }

    const amountOfProposals =
      (await $daoContract.getNumberOfProposalsInVotingSlot(
        blockNumber
      )) as bigint;

    proposals = await Promise.all(
      [...Array(Number(amountOfProposals)).keys()].map(
        async (index: Number) => {
          const proposalId = (
            await $daoContract.votingSlots(blockNumber, index)
          ).toString();

          const proposal = await proposalStore.get(proposalId, $daoContract);

          return {
            description: proposal.description,
            id: proposal.id,
            proposer: proposal.proposer,
          };
        }
      )
    );

    const votesCasted = (await $daoContract.queryFilter(
      $daoContract.filters.VoteCast(
        $userEthereumAddress,
        null,
        null,
        null,
        null
      )
    )) as EventLog[];

    for (const proposal of proposals) {
      const { againstVotes, forVotes, abstainVotes } =
        await $daoContract.proposalVotes(proposal.id);
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
      const event: EventLog | undefined = votesCasted.find(
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

  function handleVoteValue(proposalId: string, voteValue: bigint) {
    voteValues[proposalId].value = voteValue;
  }

  async function castVotes() {
    for (const [key, value] of Object.entries(voteValues)) {
      if (!value.canVote || value.value === -1n) {
        continue;
      }

      if (value.reason.trim() === "") {
        await $daoContract.castVote(key, value.value);
      } else {
        await $daoContract.castVoteWithReason(key, value.value, value.reason);
      }
    }
  }

  onMount(async () => {
    if ($membershipStatusValue !== 3n) {
      $error = "You are not allowed to view this page.";
      navigate("/dao/home");
    }
  });

  onDestroy(() => {
    $isSubmitting = false;
  });
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

<Navigation requiresChainId={$daoConfig?.chainId}>
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
          class={voteValues[proposal.id].value === 0n ? "button4" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 0n)}
        >
          <Fa icon={faXmark} size="sm" class="icon px-2" />
        </button>

        <button
          disabled={!voteValues[proposal.id].canVote}
          class={voteValues[proposal.id].value === 1n ? "button4" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 1n)}
        >
          <Fa icon={faCheck} size="sm" class="icon px-2" />
        </button>

        <button
          disabled={!voteValues[proposal.id].canVote}
          class={voteValues[proposal.id].value === 2n ? "button4" : "button3"}
          on:click={() => handleVoteValue(proposal.id, 2n)}
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

  <button disabled={!hasAnyVotes} on:click={() => castVotes()} class="button4"
    >Cast votes</button
  >
</Navigation>
