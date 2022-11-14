<script lang="ts">
  import { navigate } from "svelte-routing";
  import { daaContract, provider } from "../../ts/daaStore";
  import { isSubmitting } from "../../ts/mainStore";
  import { proposalCreatedEvents, votingSlots } from "../../ts/proposalStore";
  import formatDateTime from "../../utils/formatDateTime";
  import Navigation from "./Navigation.svelte";

  let futureVotingSlots: VotingSlot[] = [];
  let pastVotingSlots: VotingSlot[] = [];
  let slotCloseTime: number = 0;
  let currentBlockNumber: number = 0;
  let currentTime: string = "";
  let votingPeriod: number = 0;

  interface VotingSlot {
    blockInfo: BlockInfo;
    proposalInfos: ProposalInfo[];
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
      $proposalCreatedEvents === null ||
      $votingSlots === null
    ) {
      $isSubmitting = true;
    } else if (futureVotingSlots.length === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    slotCloseTime = (await $daaContract.slotCloseTime()).toNumber();
    currentBlockNumber = await $provider.getBlockNumber();
    const currentBlockTimestamp = (await $provider.getBlock(currentBlockNumber))
      .timestamp;
    currentTime = formatDateTime(new Date(currentBlockTimestamp * 1000));
    votingPeriod = (await $daaContract.votingPeriod()).toNumber();

    await createVotingSlots();

    $isSubmitting = false;
  }

  async function createVotingSlots() {
    $votingSlots.forEach(async (blockNumberForSlot: number) => {
      const blockInfo = await createBlockInfo(blockNumberForSlot);
      const proposalInfos = await createProposalInfo(blockInfo.blockNumber);
      if (blockInfo.blockNumber + votingPeriod < currentBlockNumber) {
        pastVotingSlots = [...pastVotingSlots, { proposalInfos, blockInfo }];
      } else {
        futureVotingSlots = [
          ...futureVotingSlots,
          { proposalInfos, blockInfo },
        ];
      }
    });
  }

  async function createBlockInfo(blockNumber: number): Promise<BlockInfo> {
    const secondsPerBlock = 13.3;
    if (blockNumber <= currentBlockNumber) {
      const blockTimestamp = (await $provider.getBlock(blockNumber)).timestamp;
      return {
        blockNumber,
        blockDate: formatDateTime(new Date(blockTimestamp * 1000)),
      };
    } else {
      const blockDifference = blockNumber - currentBlockNumber;
      const timeDifference = Math.abs(blockDifference * secondsPerBlock);
      const currentBlockTimestamp = (
        await $provider.getBlock(currentBlockNumber)
      ).timestamp;
      let date = new Date(currentBlockTimestamp * 1000);
      date.setSeconds(date.getSeconds() + timeDifference);
      return { blockNumber, blockDate: formatDateTime(date) };
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
    const event = $proposalCreatedEvents.find(
      (event) => event.args[0].toString() === proposalId
    );
    return event.args[8];
  }
</script>

<style>
  .card {
    display: flex;
    margin-top: 1rem;
    padding: 1rem;
    box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2);
  }
</style>

<Navigation>
  {#if futureVotingSlots.length > 0}
    <h2>Next Voting Windows</h2>
  {/if}
  {#each futureVotingSlots as slot, i}
    <div class="card">
      <div>
        <div>Voting Start</div>
        <div>#{slot.blockInfo.blockNumber}</div>
        <div>≈{slot.blockInfo.blockDate}</div>
        {#if currentBlockNumber >= slot.blockInfo.blockNumber && currentBlockNumber < slot.blockInfo.blockNumber + votingPeriod}
          <button on:click={() => navigate(`/daa/castVotes/${slot.blockInfo.blockNumber}`)} class="py-2 button3">Vote</button>
        {/if}
        {#if currentBlockNumber < slot.blockInfo.blockNumber - slotCloseTime}
          <button
            on:click={() => navigate("/daa/createProposal")}
            class="py-2 button3">Create Proposal</button
          >
        {/if}
      </div>
      <div>
        {#each slot.proposalInfos as proposalInfo, i}
          <li>Proposal {i + 1}: {proposalInfo.proposalDescription}</li>
        {/each}
      </div>
    </div>
  {/each}
  {#if pastVotingSlots.length > 0}
    <h2>Past Voting Windows</h2>
  {/if}
  {#each pastVotingSlots as slot, i}
    <div class="card">
      <div>
        <div>Voting Start</div>
        <div>#{slot.blockInfo.blockNumber}</div>
        <div>{slot.blockInfo.blockDate}</div>
      </div>
      <div>
        {#each slot.proposalInfos as proposalInfo, i}
          <li>Proposal {i + 1}: {proposalInfo.proposalDescription}</li>
        {/each}
      </div>
    </div>
  {/each}

  <p>Current-Block: #{currentBlockNumber} Current-Time: {currentTime}</p>
</Navigation>
