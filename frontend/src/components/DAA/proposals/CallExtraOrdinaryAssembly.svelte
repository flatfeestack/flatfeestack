<script lang="ts">
  import { Interface } from "ethers/lib/utils";
  import { DAAABI } from "../../../contracts/DAA";
  import type { ProposalFormProps } from "../../../types/daa";
  import yup from "../../../utils/yup";

  interface $$Props extends ProposalFormProps {}
  export let calls: $$Props["calls"];

  let formValues = {
    proposedBlockNumber: 12345,
  };

  const daaInterface = new Interface(DAAABI);

  const schema = yup.object().shape({
    proposedBlockNumber: yup.number().required(),
  });

  $: {
    try {
      schema.validateSync(formValues, { abortEarly: false });
      updateCalldata();
    } catch (err) {
      // ignore errors for now
    }
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
