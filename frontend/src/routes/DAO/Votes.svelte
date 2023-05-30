<script lang="ts">
  import { onDestroy } from "svelte";
  import { Link, navigate } from "svelte-routing";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import ExtraOrdinaryAssemblies from "../../components/DAO/votes/ExtraOrdinaryAssemblies.svelte";
  import {
    currentBlockNumber,
    currentBlockTimestamp,
    daoConfig,
    daoContract,
    membershipStatusValue,
  } from "../../ts/daoStore";
  import { provider } from "../../ts/ethStore";
  import { isSubmitting } from "../../ts/mainStore";
  import { proposalCreatedEvents, votingSlots } from "../../ts/proposalStore";
  import {
    checkUndefinedProvider,
    ensureSameChainId,
  } from "../../utils/ethHelpers";
  import formatDateTime from "../../utils/formatDateTime";
  import { futureBlockDate } from "../../utils/futureBlockDate";
  import truncateString from "../../utils/truncateString";

  let viewVotingSlots: VotingSlot[] = [];
  let slotCloseTime: number = 0;
  let currentTime: string = "";
  let votingPeriod: number = 0;

  enum VotingSlotState {
    ProposalsOpen,
    ProposalFreeze,
    VotingOpen,
    ExecutionPhase,
  }

  interface VotingSlotDates {
    proposalCreationOpenBlockNumber: number;
    proposalCreationOpenDate: string;
    votingStartBlockNumber: number;
    votingStartDate: string;
    votingEndBlockNumber: number;
    votingEndDate: string;
  }

  interface VotingSlot {
    dates: VotingSlotDates;
    id: number;
    proposalInfos: ProposalInfo[];
    votingSlotState: VotingSlotState;
  }

  interface ProposalInfo {
    proposalId: string;
    proposalDescription: string;
  }

  checkUndefinedProvider();

  $: {
    ensureSameChainId($daoConfig?.chainId);

    if (
      $currentBlockNumber === null ||
      $currentBlockTimestamp === null ||
      $daoContract === null ||
      $votingSlots === null
    ) {
      $isSubmitting = true;
    } else if (viewVotingSlots.length === 0) {
      $isSubmitting = true;
      prepareView();
    }
  }

  async function prepareView() {
    slotCloseTime = (await $daoContract.slotCloseTime()).toNumber();
    currentTime = formatDateTime(new Date($currentBlockTimestamp * 1000));
    votingPeriod = (await $daoContract.votingPeriod()).toNumber();

    await createVotingSlots();

    $isSubmitting = false;
  }

  async function createVotingSlots() {
    const creationOfSlots = $votingSlots.map(
      async (votingSlotBlock: number, index) => {
        const blockInfo = await createBlockInfo(votingSlotBlock);
        const proposalInfos = await createProposalInfo(
          blockInfo.votingStartBlockNumber
        );
        const votingSlotState = await getVotingSlotState(
          blockInfo.votingStartBlockNumber
        );

        return {
          proposalInfos,
          dates: blockInfo,
          id: $votingSlots.length - index,
          votingSlotState,
        };
      }
    );

    viewVotingSlots = await Promise.all(creationOfSlots);
  }

  async function getDateForBlock(blockNumber): Promise<string> {
    if (blockNumber <= $currentBlockNumber) {
      const blockTimestamp = (await $provider.getBlock(blockNumber)).timestamp;
      return formatDateTime(new Date(blockTimestamp * 1000));
    } else {
      return futureBlockDate(blockNumber);
    }
  }

  async function createBlockInfo(
    votingStartBlockNumber: number
  ): Promise<VotingSlotDates> {
    const dates = {
      proposalCreationOpenBlockNumber: votingStartBlockNumber - slotCloseTime,
      votingStartBlockNumber,
      votingEndBlockNumber: votingStartBlockNumber + votingPeriod,
    };

    return {
      ...dates,
      proposalCreationOpenDate: await getDateForBlock(
        dates.proposalCreationOpenBlockNumber
      ),
      votingStartDate: await getDateForBlock(dates.votingStartBlockNumber),
      votingEndDate: await getDateForBlock(dates.votingEndBlockNumber),
    };
  }

  async function createProposalInfo(
    blockNumber: number
  ): Promise<ProposalInfo[]> {
    const number = (
      await $daoContract.getNumberOfProposalsInVotingSlot(blockNumber)
    ).toNumber();
    let proposalInfos: ProposalInfo[] = [];
    for (let i = 0; i < number; i++) {
      const proposalId = (
        await $daoContract.votingSlots(blockNumber, i)
      ).toString();
      const proposalDescription = await loadProposalDescription(proposalId);
      proposalInfos.push({
        proposalId,
        proposalDescription,
      });
    }
    return proposalInfos;
  }

  async function loadProposalDescription(proposalId: string): Promise<string> {
    const event = await proposalCreatedEvents.get(proposalId, $daoContract);
    return event.event.args[8];
  }

  async function getVotingSlotState(
    votingSlotBlockNumber: number
  ): Promise<VotingSlotState> {
    if ($currentBlockNumber < votingSlotBlockNumber) {
      if ($currentBlockNumber < votingSlotBlockNumber - slotCloseTime) {
        return VotingSlotState.ProposalsOpen;
      } else {
        return VotingSlotState.ProposalFreeze;
      }
    } else {
      if ($currentBlockNumber > votingSlotBlockNumber + votingPeriod) {
        return VotingSlotState.ExecutionPhase;
      } else {
        return VotingSlotState.VotingOpen;
      }
    }
  }

  onDestroy(() => {
    $isSubmitting = false;
  });
</script>

<style>
  .card {
    margin-top: 1rem;
    padding: 1rem;
    box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2);
  }

  p,
  h2 {
    margin: 0.25rem;
  }
</style>

<Navigation>
  <p>Last updated (block): #{$currentBlockNumber}</p>
  <p>
    Last updated (time): Current-Time: {currentTime}

    <ExtraOrdinaryAssemblies currentBlockTimestamp={$currentBlockTimestamp} />

    {#each viewVotingSlots as slotInfo}
      <div class="card">
        <h2 class="text-secondary-900">
          Voting slot #{slotInfo.id}
        </h2>

        {#if slotInfo.votingSlotState === VotingSlotState.ProposalsOpen}
          <p>
            Proposal creation open until #{slotInfo.dates
              .proposalCreationOpenBlockNumber} (approx. {slotInfo.dates
              .proposalCreationOpenDate})
          </p>
          <p>
            Voting scheduled for #{slotInfo.dates.votingStartBlockNumber} (approx.
            {slotInfo.dates.votingStartDate})
          </p>
        {:else if slotInfo.votingSlotState === VotingSlotState.ProposalFreeze}
          <p>
            Proposal creation closed since #{slotInfo.dates
              .proposalCreationOpenBlockNumber} ({slotInfo.dates
              .proposalCreationOpenDate})
          </p>
          <p>
            Voting start scheduled for #{slotInfo.dates.votingStartBlockNumber} (approx.
            {slotInfo.dates.votingStartDate})
          </p>

          <p>
            Voting end scheduled for #{slotInfo.dates.votingEndBlockNumber} (approx.
            {slotInfo.dates.votingEndDate})
          </p>
        {:else if slotInfo.votingSlotState === VotingSlotState.VotingOpen}
          <p>
            Voting open until #{slotInfo.dates.votingEndBlockNumber} (approx. {slotInfo
              .dates.votingEndDate})
          </p>
        {:else}
          <p>
            Voting closed since #{slotInfo.dates.votingEndBlockNumber} ({slotInfo
              .dates.votingEndDate})
          </p>
        {/if}

        {#if slotInfo.proposalInfos.length > 0}
          <ul>
            {#each slotInfo.proposalInfos as proposalInfo, i}
              <li>
                <Link to="/dao/proposals/{proposalInfo.proposalId}"
                  >Proposal {i + 1}: {truncateString(
                    proposalInfo.proposalDescription,
                    30
                  )}</Link
                >
              </li>
            {/each}
          </ul>
        {:else}
          <p class="italic">No proposals submitted.</p>
        {/if}

        {#if $membershipStatusValue == 3}
          {#if slotInfo.votingSlotState == VotingSlotState.ProposalsOpen}
            <button
              on:click={() => navigate("/dao/createProposal")}
              class="py-2 button3">Create Proposal</button
            >
          {:else if slotInfo.votingSlotState == VotingSlotState.VotingOpen && slotInfo.proposalInfos.length > 0}
            <button
              on:click={() =>
                navigate(
                  `/dao/castVotes/${slotInfo.dates.votingStartBlockNumber}`
                )}
              class="py-2 button3">Vote</button
            >
          {:else if slotInfo.votingSlotState == VotingSlotState.ExecutionPhase && slotInfo.proposalInfos.length > 0}
            <button
              on:click={() =>
                navigate(
                  `/dao/executeProposals/${slotInfo.dates.votingStartBlockNumber}`
                )}
              class="py-2 button3">Execute proposals</button
            >
          {/if}
        {/if}
      </div>
    {/each}
  </p>
</Navigation>
