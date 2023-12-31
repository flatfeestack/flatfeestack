<script lang="ts">
  import humanizeDuration from "humanize-duration";
  import { daoContract } from "../../../ts/daoStore";
  import { userEthereumAddress } from "../../../ts/ethStore";
  import { isSubmitting } from "../../../ts/mainStore";
  import {
    extraOrdinaryAssemblyRequestProposalIds,
    proposalStore,
    votesCasted,
    type Proposal,
  } from "../../../ts/proposalStore";
  import { futureBlockDate } from "../../../utils/futureBlockDate";
  import {
    executeProposal,
    queueProposal,
  } from "../../../utils/proposalFunctions";
  import VoteButtonGroup from "./VoteButtonGroup.svelte";
  import type { EventLog } from "ethers";

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

    const proposals: Proposal[] = await Promise.all(
      $extraOrdinaryAssemblyRequestProposalIds.map(
        async (proposalId) =>
          await proposalStore.get(proposalId.toString(), $daoContract)
      )
    );

    // ignore all proposals that are not pending, succeeded or queued
    const extraOrdinaryAssemblyRequestStates = await Promise.all(
      proposals.map(async (proposal) => {
        const proposalState = await $daoContract.state(proposal.id);

        return {
          ...proposal,
          state: proposalState,
        };
      })
    );

    extraOrdinaryAssemblyRequests = await Promise.all(
      extraOrdinaryAssemblyRequestStates
        // ignore all proposals that are not pending, succeeded or queued
        .filter((intermediateObject) =>
          [1, 4, 5].includes(intermediateObject.state)
        )
        .map(async (intermediateObject) => {
          const proposalEta = Number(
            (await $daoContract.proposalEta(intermediateObject.id)) as bigint
          );
          const event: EventLog | undefined =
            $votesCasted === null
              ? undefined
              : $votesCasted.find(
                  (event) =>
                    event.args[1].toString() ===
                    intermediateObject.id.toString()
                );

          return {
            calldatas: intermediateObject.calldatas,
            canVote: $userEthereumAddress !== null && event === undefined,
            deadline: intermediateObject.endBlock,
            description: intermediateObject.description,
            eta: proposalEta < currentBlockTimestamp ? 0 : proposalEta,
            id: intermediateObject.id,
            proposer: intermediateObject.proposer,
            state: intermediateObject.state,
            targets: intermediateObject.targets,
            values: intermediateObject.values,
            voteValue: event === undefined ? null : event.args[2],
          };
        })
    );

    $isSubmitting = false;
  }

  async function castVote(proposalId: string, voteValue: number) {
    await $daoContract.castVote(proposalId, voteValue);
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
