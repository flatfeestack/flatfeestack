<script lang="ts">
  import { Interface } from "ethers/lib/utils";
  import humanizeDuration from "humanize-duration";
  import { DAAABI } from "../../../contracts/DAA";
  import { currentBlockNumber, daaContract } from "../../../ts/daaStore";
  import type { ProposalFormProps } from "../../../types/daa";
  import { secondsPerBlock } from "../../../utils/futureBlockDate";
  import yup from "../../../utils/yup";
  import Spinner from "../../Spinner.svelte";
  import { futureBlockDate } from "../../../utils/futureBlockDate";

  interface $$Props extends ProposalFormProps {}
  export let calls: $$Props["calls"];

  let formValues = {
    proposedBlockNumber: 12345,
  };

  let isLoading = true;
  let minimumBlockNumber = 0;

  interface ExtraOrdinaryAssemblyParameters {
    timelockMinimumDelay: number;
    votingPeriod: number;
    votingSlotAnnouncementPeriod: number;
  }

  let extraOrdinaryAssemblyParameters: ExtraOrdinaryAssemblyParameters | null =
    null;

  const daaInterface = new Interface(DAAABI);

  const schema = yup.object().shape({
    proposedBlockNumber: yup.number().min(minimumBlockNumber).required(),
  });

  $: {
    if ($daaContract === null) {
      isLoading = true;
    } else {
      if (extraOrdinaryAssemblyParameters === null) {
        isLoading = true;
        loadAssemblyParameters();
      } else {
        isLoading = false;

        try {
          schema.validateSync(formValues, { abortEarly: false });
          updateCalldata();
        } catch (err) {
          // ignore errors for now
        }
      }
    }

    try {
      schema.validateSync(formValues, { abortEarly: false });
      updateCalldata();
    } catch (err) {
      // ignore errors for now
    }
  }

  async function loadAssemblyParameters() {
    const [
      timelockMinimumDelay,
      extraAssemblyVotingPeriod,
      votingSlotAnnouncementPeriod,
    ] = await Promise.all([
      $daaContract.getMinDelay(),
      $daaContract.extraOrdinaryAssemblyVotingPeriod(),
      $daaContract.votingSlotAnnouncementPeriod(),
    ]);

    extraOrdinaryAssemblyParameters = {
      timelockMinimumDelay: timelockMinimumDelay.toNumber(),
      votingPeriod: extraAssemblyVotingPeriod.toNumber(),
      votingSlotAnnouncementPeriod: votingSlotAnnouncementPeriod.toNumber(),
    };

    console.log(extraOrdinaryAssemblyParameters);

    minimumBlockNumber =
      $currentBlockNumber +
      extraOrdinaryAssemblyParameters.timelockMinimumDelay / secondsPerBlock +
      extraOrdinaryAssemblyParameters.votingPeriod +
      extraOrdinaryAssemblyParameters.votingSlotAnnouncementPeriod;
  }

  function updateCalldata() {
    calls = [
      {
        target: import.meta.env.VITE_DAA_CONTRACT_ADDRESS,
        transferCallData: daaInterface.encodeFunctionData("setVotingSlot", [
          formValues.proposedBlockNumber,
        ]),
        value: 0,
      },
    ];
  }
</script>

<style>
  .combine-column {
    grid-column: 1 / 3;
  }
</style>

<label for="proposedBlockNumber">Proposed block number</label>
<input
  type="text"
  id="proposedBlockNumber"
  name="proposedBlockNumber"
  bind:value={formValues.proposedBlockNumber}
  required
/>
{#await schema.validateAt("proposedBlockNumber", formValues)}{:catch error}
  <p class="invalid" style="color:red">{error.errors[0]}</p>
{/await}

{#if isLoading}
  <Spinner />
{:else}
  <div class="combine-column italic">
    A proposal for an extra ordinary assembly does not belong to a voting slot
    and is published immediately. However, a few rules apply to the process.

    <ul>
      <li>
        Member can vote for {extraOrdinaryAssemblyParameters.votingPeriod} blocks
        about this proposal.
      </li>
      <li>
        The proposal needs to be queued for {humanizeDuration(
          extraOrdinaryAssemblyParameters.timelockMinimumDelay * 1000
        )}.
      </li>
      <li>
        A voting slot needs to be announced {extraOrdinaryAssemblyParameters.votingSlotAnnouncementPeriod}
        blocks in advance.
      </li>
    </ul>

    Given we're at block #{$currentBlockNumber}, the earliest you can request an
    extraordinary assembly would be at block #{minimumBlockNumber} (approx. {futureBlockDate(
      minimumBlockNumber,
      $currentBlockNumber
    )}). However, this would require that you create the proposal right now and
    queued and execute it immediately when ready. Therefore, calculate an
    additional day or two in (we calculate with 7200 blocks per day).
  </div>
{/if}
