<script lang="ts">
  import Spinner from "../../components/Spinner.svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import {
    proposalCreatedEvents,
    type ProposalCreatedEvent,
  } from "../../ts/proposalStore";
  import {
    currentBlockTimestamp,
    daoContract,
    membershipStatusValue,
  } from "../../ts/daoStore";
  import { futureBlockDate } from "../../utils/futureBlockDate";
  import { Link } from "svelte-routing";
  import humanizeDuration from "humanize-duration";

  export let proposalId: string;

  let proposal: ProposalCreatedEvent;
  let isLoading = true;
  let proposalEta = 0;
  let proposalState = 0;

  $: {
    if ($daoContract === null) {
      isLoading = true;
    } else {
      prepareView();
    }
  }

  async function prepareView() {
    [proposal, proposalState] = await Promise.all([
      proposalCreatedEvents.get(proposalId, $daoContract),
      $daoContract.state(proposalId),
    ]);

    if (proposalState === 5) {
      proposalEta = await $daoContract.proposalEta(proposalId);
    }

    isLoading = false;
  }
</script>

<Navigation>
  {#if isLoading}
    <Spinner />
  {:else}
    <h2>Details about proposal</h2>

    <p class="bold">Proposer</p>
    <p>{proposal.event.args[1]}</p>

    <p class="bold">Description</p>
    <p>{proposal.event.args[8]}</p>

    <p class="bold">State</p>
    {#if proposalState === 0}
      <p>
        Vote of proposal is scheduled for #{proposal.event.args[6]}
        {#await futureBlockDate(proposal.event.args[6])}(approx ...){:then date}(approx
          {date}){/await}
      </p>
    {:else if proposalState === 1}
      <p>
        Vote for proposal is open until #{proposal.event.args[7]}
        {#await futureBlockDate(proposal.event.args[7])}(approx ...){:then date}(approx
          {date}){/await}.
        {#if $membershipStatusValue === 3}
          <Link to={`/dao/castVotes/${proposal.event.args[6]}`}>Cast vote!</Link
          >
        {/if}
      </p>
    {:else if proposalState === 2}
      <p>The proposal has been cancelled.</p>
    {:else if proposalState === 3}
      <p>The proposal has been denied.</p>
    {:else if proposalState === 4}
      <p>
        The vote for the proposal was successful.
        {#if $membershipStatusValue === 3}
          <Link to={`/dao/executeProposals/${proposal.event.args[6]}`}
            >Enqueue proposal</Link
          >
        {/if}
      </p>
    {:else if proposalState === 5}
      {#if proposalEta === 0}
        <p>
          The proposal is ready for execution.
          {#if $membershipStatusValue === 3}
            <Link to={`/dao/executeProposals/${proposal.event.args[6]}`}
              >Execute proposal</Link
            >
          {/if}
        </p>
      {:else}
        The proposal can be executed in {humanizeDuration(
          (proposalEta - $currentBlockTimestamp) * 1000
        )}.
      {/if}
    {:else if proposalState === 6}
      <p>The time to execute the proposal has expired.</p>
    {:else if proposalState === 7}
      <p>The proposal has been executed.</p>
    {/if}
  {/if}
</Navigation>
