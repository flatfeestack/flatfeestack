<script lang="ts">
  import { Viewer } from "bytemd";
  import type { Event } from "ethers";
  import { daaContract, userEthereumAddress } from "../../../ts/daaStore";
  import { isSubmitting } from "../../../ts/mainStore";
  import {
    extraOrdinaryAssemblyRequestProposalIds,
    proposalCreatedEvents,
    votesCasted,
  } from "../../../ts/proposalStore";
  import { futureBlockDate } from "../../../utils/futureBlockDate";
  import VoteButtonGroup from "./VoteButtonGroup.svelte";

  let extraOrdinaryAssemblyRequests = [];
  export let currentBlockNumber: number | null = null;

  // subscribe to the store for notification when the user decides to connect their wallet
  extraOrdinaryAssemblyRequestProposalIds.subscribe(
    async (_extraOrdinaryAssemblyRequestProposalIds) => {
      await createRunningExtraOrdinaryAssemblyVotes();
    }
  );

  userEthereumAddress.subscribe(async (_userEthereumAddress) => {
    await createRunningExtraOrdinaryAssemblyVotes();
  });

  votesCasted.subscribe(async (_votesCasted) => {
    await createRunningExtraOrdinaryAssemblyVotes();
  });

  async function createRunningExtraOrdinaryAssemblyVotes() {
    if ($extraOrdinaryAssemblyRequestProposalIds === null) {
      console.log("hello");
      return;
    }

    const events = await Promise.all(
      $extraOrdinaryAssemblyRequestProposalIds.map(
        async (proposalId) =>
          await proposalCreatedEvents.get(proposalId.toString(), $daaContract)
      )
    );

    extraOrdinaryAssemblyRequests = events
      .filter((event) => event.event.args[7].toNumber() > currentBlockNumber)
      .map((proposalCreatedEvent) => {
        const event: Event | undefined =
          $votesCasted === null
            ? undefined
            : $votesCasted.find(
                (event) =>
                  event.args[1].toString() ===
                  proposalCreatedEvent.event.args[0].toString()
              );

        return {
          canVote: $userEthereumAddress !== null && event === undefined,
          deadline: proposalCreatedEvent.event.args[7],
          description: proposalCreatedEvent.event.args[8],
          id: proposalCreatedEvent.event.args[0],
          proposer: proposalCreatedEvent.event.args[1],
          voteValue: event === undefined ? null : event.args[2],
        };
      });

    $isSubmitting = false;
  }

  async function castVote(proposalId: string, voteValue: number) {
    await $daaContract.castVote(proposalId, voteValue);
  }
</script>

<style>
  .card-important {
    margin-top: 1rem;
    padding: 1rem;
    box-shadow: 0 4px 8px 0 rgba(252, 165, 165, 1);
    color: rgba(252, 165, 165, 1);
  }

  .vote-container {
    align-items: center;
    display: flex;
    justify-content: space-between;
  }
</style>

{#each extraOrdinaryAssemblyRequests as extraOrdinaryAssemblyRequest, index}
  <div class="card-important">
    <h2>Active proposal for extra ordinary assembly</h2>

    Proposer: {extraOrdinaryAssemblyRequest.proposer}
    Vote running until #{extraOrdinaryAssemblyRequest.deadline}
    {#await futureBlockDate(extraOrdinaryAssemblyRequest.deadline, currentBlockNumber) then futureDate}
      (approx. {futureDate})
    {/await}.

    <Viewer value={extraOrdinaryAssemblyRequest.description} />

    {#if $userEthereumAddress !== null}
      <div class="vote-container">
        <p>Your vote:</p>
        <VoteButtonGroup
          disabled={!extraOrdinaryAssemblyRequest.canVote}
          onClick={castVote}
          proposalId={extraOrdinaryAssemblyRequest.id}
          voteValue={extraOrdinaryAssemblyRequest.voteValue}
        />
      </div>
    {/if}
  </div>
{/each}
