<script lang="ts">
  import humanizeDuration from "humanize-duration";
  import { onDestroy } from "svelte";
  import { navigate } from "svelte-routing";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import {
    currentBlockTimestamp,
    daoConfig,
    daoContract,
  } from "../../ts/daoStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import { proposalStore, votingSlots } from "../../ts/proposalStore";
  import { checkUndefinedProvider } from "../../utils/ethHelpers";
  import {
    executeProposal,
    queueProposal,
  } from "../../utils/proposalFunctions";

  export let blockNumber: string;
  let proposals = [];

  checkUndefinedProvider();

  $: {
    if (
      $proposalStore === null ||
      $votingSlots === null ||
      $currentBlockTimestamp === null
    ) {
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

    const amountOfProposals =
      (await $daoContract.getNumberOfProposalsInVotingSlot(
        blockNumber
      )) as bigint;

    proposals = await Promise.all(
      [...Array(Number(amountOfProposals)).keys()].map(
        async (index: Number) => {
          const proposalId = await $daoContract.votingSlots(blockNumber, index);

          const [proposalState, proposalEta] = await Promise.all([
            $daoContract.state(proposalId),
            $daoContract.proposalEta(proposalId),
          ]);

          const proposal = await proposalStore.get(proposalId, $daoContract);

          return {
            calldatas: proposal.calldatas,
            description: proposal.description,
            eta: proposalEta < $currentBlockTimestamp ? 0 : proposalEta,
            id: proposalId.toString(),
            proposer: proposal.proposer,
            state: proposalState,
            targets: proposal.targets,
            values: proposal.values,
          };
        }
      )
    );

    $isSubmitting = false;
  }

  onDestroy(() => {
    $isSubmitting = false;
  });
</script>

<Navigation requiresChainId={$daoConfig?.chainId}>
  <h1 class="text-secondary-900">Execute proposals</h1>

  {#each proposals as proposal, i}
    <h2>Proposal {i + 1}</h2>

    <p>Proposer: {proposal.proposer}</p>

    <p>{proposal.description}</p>

    {#if proposal.state == 4}
      <button
        on:click={() =>
          queueProposal(
            proposal.targets,
            proposal.values,
            proposal.description,
            proposal.calldatas
          )}
        class="button4">Queue proposal for execution</button
      >
    {:else if proposal.state == 5}
      {#if proposal.eta === 0}
        <p class="italic">The proposal is ready for execution!</p>
        <button
          on:click={() =>
            executeProposal(
              proposal.targets,
              proposal.values,
              proposal.description,
              proposal.calldatas
            )}
          class="button4">Execute proposal</button
        >
      {:else}
        <p class="italic">
          The proposal can be executed in {humanizeDuration(
            (Number(proposal.eta) - $currentBlockTimestamp) * 1000
          )}.
        </p>
        <button class="button4" disabled>Execute proposal</button>
      {/if}
    {:else if proposal.state == 7}
      <p class="italic">Proposal has been executed.</p>
    {:else if proposal.state == 3}
      <p class="italic">
        The proposal cannot be executed as the vote didn't pass.
      </p>
    {:else}
      <p class="italic">
        The proposal cannot be executed as voting is still pending or running.
      </p>
    {/if}
  {/each}
</Navigation>
