<script lang="ts">
  import { navigate } from "svelte-routing";
  import {
    currentBlockTimestamp,
    daaContract,
    membershipStatusValue,
    provider,
  } from "../../ts/daaStore";
  import { isSubmitting } from "../../ts/mainStore";
  import { proposalCreatedEvents, votingSlots } from "../../ts/proposalStore";
  import formatDateTime from "../../utils/formatDateTime";
  import { futureBlockDate } from "../../utils/futureBlockDate";
  import Navigation from "./Navigation.svelte";
  import ExtraOrdinaryAssemblies from "./votes/ExtraOrdinaryAssemblies.svelte";

  let viewVotingSlots: VotingSlotsContainer = {};
  let slotCloseTime: number = 0;
  let currentBlockNumber: number = 0;
  let currentTime: string = "";
  let votingPeriod: number = 0;

  enum VotingSlotState {
    ProposalsOpen,
    ProposalFreeze,
    VotingOpen,
    ExecutionPhase,
  }

  interface VotingSlot {
    blockDate: string;
    id: number;
    proposalInfos: ProposalInfo[];
    votingSlotState: VotingSlotState;
  }

  interface VotingSlotsContainer {
    [key: number]: VotingSlot;
  }

  interface BlockInfo {
    blockNumber: number;
    blockDate: string;
  }

  interface ProposalInfo {
    proposalId: string;
    proposalDescription: string;
  }

  $: {
    if (
      $daaContract === null ||
      $votingSlots === null ||
      $currentBlockTimestamp === null
    ) {
      $isSubmitting = true;
    } else if (Object.keys(viewVotingSlots).length === 0) {
      $isSubmitting = true;
      prepareView();
    }
  }

  async function prepareView() {
    slotCloseTime = (await $daaContract.slotCloseTime()).toNumber();
    currentBlockNumber = await $provider.getBlockNumber();
    currentTime = formatDateTime(new Date($currentBlockTimestamp * 1000));
    votingPeriod = (await $daaContract.votingPeriod()).toNumber();

    await createVotingSlots();

    $isSubmitting = false;
  }

  async function createVotingSlots() {
    $votingSlots.forEach(async (votingSlotBlock: number, index) => {
      const blockInfo = await createBlockInfo(votingSlotBlock);
      const proposalInfos = await createProposalInfo(blockInfo.blockNumber);
      const votingSlotState = await getVotingSlotState(blockInfo.blockNumber);

      viewVotingSlots = {
        ...viewVotingSlots,
        [blockInfo.blockNumber]: {
          proposalInfos,
          blockDate: blockInfo.blockDate,
          id: index + 1,
          votingSlotState,
        },
      };
    });

    viewVotingSlots = Object.keys(viewVotingSlots)
      .sort()
      .reverse()
      .reduce((obj, key) => {
        obj[key] = viewVotingSlots[key];
        return obj;
      }, {});
  }

  async function createBlockInfo(
    futureBlockNumber: number
  ): Promise<BlockInfo> {
    if (futureBlockNumber <= currentBlockNumber) {
      const blockTimestamp = (await $provider.getBlock(futureBlockNumber))
        .timestamp;
      return {
        blockNumber: futureBlockNumber,
        blockDate: formatDateTime(new Date(blockTimestamp * 1000)),
      };
    } else {
      return {
        blockNumber: futureBlockNumber,
        blockDate: futureBlockDate(futureBlockNumber, currentBlockNumber),
      };
    }
  }

  async function createProposalInfo(
    blockNumber: number
  ): Promise<ProposalInfo[]> {
    const number = (
      await $daaContract.getNumberOfProposalsInVotingSlot(blockNumber)
    ).toNumber();
    let proposalInfos: ProposalInfo[] = [];
    for (let i = 0; i < number; i++) {
      const proposalId = (
        await $daaContract.votingSlots(blockNumber, i)
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
    const event = await proposalCreatedEvents.get(proposalId, $daaContract);
    return event.event.args[8];
  }

  async function getVotingSlotState(
    votingSlotBlockNumber: number
  ): Promise<VotingSlotState> {
    if (currentBlockNumber < votingSlotBlockNumber) {
      if (currentBlockNumber < votingSlotBlockNumber - slotCloseTime) {
        return VotingSlotState.ProposalsOpen;
      } else {
        return VotingSlotState.ProposalFreeze;
      }
    } else {
      if (currentBlockNumber > votingSlotBlockNumber + votingPeriod) {
        return VotingSlotState.ExecutionPhase;
      } else {
        return VotingSlotState.VotingOpen;
      }
    }
  }
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
  <p>Last updated (block): #{currentBlockNumber}</p>
  <p>
    Last updated (time): Current-Time: {currentTime}

    <ExtraOrdinaryAssemblies
      {currentBlockNumber}
      currentBlockTimestamp={$currentBlockTimestamp}
    />

    {#each Object.entries(viewVotingSlots).reverse() as [blockNumber, slotInfo]}
      <div class="card">
        <h2 class="text-secondary-900">
          Voting slot #{slotInfo.id}
        </h2>

        {#if slotInfo.votingSlotState === VotingSlotState.ProposalsOpen}
          <p>
            Proposal creation open until #{Number(blockNumber) - slotCloseTime}
          </p>
          <p>Voting scheduled for #{blockNumber}</p>
        {:else if slotInfo.votingSlotState === VotingSlotState.ProposalFreeze}
          <p>
            Proposal creation closed since #{Number(blockNumber) -
              slotCloseTime}
          </p>
          <p>Voting scheduled for #{blockNumber}</p>
        {:else if slotInfo.votingSlotState === VotingSlotState.VotingOpen}
          <p>Voting open until #{Number(blockNumber) + votingPeriod}</p>
        {:else}
          <p>Voting closed since #{Number(blockNumber) + votingPeriod}</p>
        {/if}

        {#if slotInfo.proposalInfos.length > 0}
          <ul>
            {#each slotInfo.proposalInfos as proposalInfo, i}
              <li>Proposal {i + 1}: {proposalInfo.proposalDescription}</li>
            {/each}
          </ul>
        {:else}
          <p class="italic">No proposals submitted.</p>
        {/if}

        {#if $membershipStatusValue == 3}
          {#if slotInfo.votingSlotState == VotingSlotState.ProposalsOpen}
            <button
              on:click={() => navigate("/daa/createProposal")}
              class="py-2 button3">Create Proposal</button
            >
          {:else if slotInfo.votingSlotState == VotingSlotState.VotingOpen}
            <button
              on:click={() => navigate(`/daa/castVotes/${blockNumber}`)}
              class="py-2 button3">Vote</button
            >
          {:else if slotInfo.votingSlotState == VotingSlotState.ExecutionPhase}
            <button
              on:click={() => navigate(`/daa/executeProposals/${blockNumber}`)}
              class="py-2 button3">Execute proposals</button
            >
          {/if}
        {/if}
      </div>
    {/each}
  </p></Navigation
>
