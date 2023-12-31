<script lang="ts">
  import humanizeDuration from "humanize-duration";
  import { HTTPError } from "ky";
  import { onDestroy } from "svelte";
  import { Link } from "svelte-routing";
  import type { Unsubscriber } from "svelte/store";
  import Call from "../../components/DAO/Call.svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import { API } from "../../ts/api";
  import {
    currentBlockTimestamp,
    daoContract,
    membershipStatusValue,
  } from "../../ts/daoStore";
  import { error } from "../../ts/mainStore";
  import { proposalStore, type Proposal } from "../../ts/proposalStore";
  import type { Post } from "../../types/forum";
  import { futureBlockDate } from "../../utils/futureBlockDate";

  export let proposalId: string;

  let discussion: Post | null = undefined;
  let isLoading = true;
  let membershipStatusNumber: number;
  let membershipStatusUnsubscriber: Unsubscriber;
  let proposal: Proposal;
  let proposalEta = 0;
  let proposalState = 0;

  $: {
    if ($currentBlockTimestamp === null || $daoContract === null) {
      isLoading = true;
    } else {
      prepareView();
    }
  }

  async function prepareView() {
    await Promise.all([getDiscussion(), getProposal()]);

    isLoading = false;
  }

  async function getDiscussion(): Promise<void> {
    try {
      discussion = await API.forum.getPostByProposalId(proposalId);
    } catch (e) {
      if (e instanceof HTTPError && e.response.status === 404) {
        discussion = null;
      } else {
        $error = e.message;
      }
    }
  }

  async function getProposal(): Promise<void> {
    try {
      [proposal, proposalState] = await Promise.all([
        proposalStore.get(proposalId, $daoContract),
        $daoContract.state(proposalId),
      ]);

      if (proposalState === 5) {
        proposalEta = Number(
          (await $daoContract.proposalEta(proposalId)) as bigint
        );
        proposalEta = proposalEta < $currentBlockTimestamp ? 0 : proposalEta;
      }
    } catch (e) {
      $error = e.message;
    }
  }

  membershipStatusUnsubscriber = membershipStatusValue.subscribe(
    (value: bigint | null) => {
      if (value === null) {
        membershipStatusNumber = 0;
      } else {
        membershipStatusNumber = Number(value);
      }
    }
  );

  onDestroy(() => {
    if (membershipStatusUnsubscriber !== undefined) {
      membershipStatusUnsubscriber();
    }
  });

  const isDaoMember = () => membershipStatusNumber === 3;
</script>

<Navigation>
  {#if isLoading}
    <Spinner />
  {:else}
    <h2>Details about proposal</h2>

    <p class="bold">Proposer</p>
    <p>{proposal.proposer}</p>

    <p class="bold">Description</p>
    <p class="whitespace-pre-wrap">{proposal.description}</p>

    <p class="bold">State</p>
    {#if proposalState === 0}
      <p>
        Vote of proposal is scheduled for #{proposal.startBlock}
        {#await futureBlockDate(proposal.startBlock)}(approx ...){:then date}(approx
          {date}){/await}
      </p>
    {:else if proposalState === 1}
      <p>
        Vote for proposal is open until #{proposal.endBlock}
        {#await futureBlockDate(proposal.endBlock)}(approx ...){:then date}(approx
          {date}){/await}.
        {#if isDaoMember()}
          <Link to={`/dao/castVotes/${proposal.startBlock}`}>Cast vote!</Link>
        {/if}
      </p>
    {:else if proposalState === 2}
      <p>The proposal has been cancelled.</p>
    {:else if proposalState === 3}
      <p>The proposal has been denied.</p>
    {:else if proposalState === 4}
      <p>
        The vote for the proposal was successful.
        {#if isDaoMember()}
          <Link to={`/dao/executeProposals/${proposal.startBlock}`}
            >Enqueue proposal</Link
          >
        {/if}
      </p>
    {:else if proposalState === 5}
      {#if proposalEta === 0}
        <p>
          The proposal is ready for execution.
          {#if isDaoMember()}
            <Link to={`/dao/executeProposals/${proposal.startBlock}`}
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

    <p class="bold">Discussion</p>
    {#if discussion === null}
      <p>No active discussion is known for this proposal.</p>
    {:else}
      <Link to={`/dao/discussion/${discussion.id}`}>{discussion.title}</Link>
    {/if}

    <p class="bold mb-2">Calls</p>
    {#each proposal.targets as target, index}
      <Call
        calldata={String(proposal.calldatas[index])}
        {index}
        {target}
        value={Number(proposal.values[index])}
      />
    {/each}
  {/if}
</Navigation>
