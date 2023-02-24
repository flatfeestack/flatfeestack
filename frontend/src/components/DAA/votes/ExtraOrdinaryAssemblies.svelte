<script lang="ts">
  import type { Event } from "ethers";
  import humanizeDuration from "humanize-duration";
  import { daaContract, userEthereumAddress } from "../../../ts/daaStore";
  import { isSubmitting } from "../../../ts/mainStore";
  import {
    extraOrdinaryAssemblyRequestProposalIds,
    proposalCreatedEvents,
    votesCasted,
  } from "../../../ts/proposalStore";
  import { futureBlockDate } from "../../../utils/futureBlockDate";
  import {
    executeProposal,
    queueProposal,
  } from "../../../utils/proposalFunctions";
  import VoteButtonGroup from "./VoteButtonGroup.svelte";

  let extraOrdinaryAssemblyRequests = [];
  export let currentBlockTimestamp: number | null = null;

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
      return;
    }

    const events = await Promise.all(
      $extraOrdinaryAssemblyRequestProposalIds.map(
        async (proposalId) =>
          await proposalCreatedEvents.get(proposalId.toString(), $daaContract)
      )
    );

    // ignore all proposals that are not pending, succeeded or queued
    const extraOrdinaryAssemblyRequestStates = await Promise.all(
      events.map(async (proposalCreatedEvent) => {
        const proposalId = proposalCreatedEvent.event.args[0];
        const proposalState = await $daaContract.state(proposalId);

        return {
          event: proposalCreatedEvent.event,
          proposalId,
          proposalState,
        };
      })
    );

    extraOrdinaryAssemblyRequests = await Promise.all(
      extraOrdinaryAssemblyRequestStates
        // ignore all proposals that are not pending, succeeded or queued
        .filter((intermediateObject) =>
          [1, 4, 5].includes(intermediateObject.proposalState)
        )
        .map(async (intermediateObject) => {
          const proposalEta = await $daaContract.proposalEta(
            intermediateObject.proposalId
          );
          const event: Event | undefined =
            $votesCasted === null
              ? undefined
              : $votesCasted.find(
                  (event) =>
                    event.args[1].toString() ===
                    intermediateObject.proposalId.toString()
                );

          return {
            calldatas: intermediateObject.event.args[5],
            canVote: $userEthereumAddress !== null && event === undefined,
            deadline: intermediateObject.event.args[7],
            description: intermediateObject.event.args[8],
            eta: proposalEta < currentBlockTimestamp ? 0 : proposalEta,
            id: intermediateObject.proposalId,
            proposer: intermediateObject.event.args[1],
            state: intermediateObject.proposalState,
            targets: intermediateObject.event.args[2],
            values: intermediateObject.event.args[3],
            voteValue: event === undefined ? null : event.args[2],
          };
        })
    );

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

  .button1 {
    border: 1px solid rgba(252, 165, 165, 1);
    background: rgba(252, 165, 165, 1);
  }

  .button3 {
    color: rgba(252, 165, 165, 1);
  }
</style>

{#each extraOrdinaryAssemblyRequests as extraOrdinaryAssemblyRequest, index}
  <div class="card-important">
    <h2>Active proposal for extra ordinary assembly</h2>

    <p>Proposer: {extraOrdinaryAssemblyRequest.proposer}</p>

    <p>{extraOrdinaryAssemblyRequest.description}</p>

    {#if extraOrdinaryAssemblyRequest.state === 1}
      Vote running until #{extraOrdinaryAssemblyRequest.deadline} (approx. {futureBlockDate(
        extraOrdinaryAssemblyRequest.deadline
      )}).

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
    {:else if extraOrdinaryAssemblyRequest.state === 4}
      Vote succeeded, proposal can be queued.

      {#if $userEthereumAddress !== null}
        <button
          on:click={() =>
            queueProposal(
              extraOrdinaryAssemblyRequest.targets,
              extraOrdinaryAssemblyRequest.values,
              extraOrdinaryAssemblyRequest.description,
              extraOrdinaryAssemblyRequest.calldatas
            )}
          class="py-2 button3">Queue Proposal Proposal</button
        >
      {/if}
    {:else if extraOrdinaryAssemblyRequest.state == 5}
      {#if extraOrdinaryAssemblyRequest.eta === 0}
        <p class="italic">The proposal is ready for execution!</p>

        {#if $userEthereumAddress !== null}
          <button
            on:click={() =>
              executeProposal(
                extraOrdinaryAssemblyRequest.targets,
                extraOrdinaryAssemblyRequest.values,
                extraOrdinaryAssemblyRequest.description,
                extraOrdinaryAssemblyRequest.calldatas
              )}
            class="button1">Execute proposal</button
          >
        {/if}
      {:else}
        <p class="italic">
          The proposal can be executed in {humanizeDuration(
            (extraOrdinaryAssemblyRequest.eta - currentBlockTimestamp) * 1000
          )}.
        </p>
        <button class="button1" disabled>Execute proposal</button>
      {/if}
    {/if}
  </div>
{/each}
