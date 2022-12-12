<script lang="ts">
  import { Interface } from "ethers/lib/utils";
  import { DAAABI } from "../../../contracts/DAA";
  import type { ProposalFormProps } from "../../../types/daa";
  import yup from "../../../utils/yup";

  interface $$Props extends ProposalFormProps {}

  export let targets: $$Props["targets"];
  export let values: $$Props["values"];
  export let transferCallData: $$Props["transferCallData"];

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
    values = [0];
    targets = [import.meta.env.VITE_DAA_CONTRACT_ADDRESS];
    transferCallData = [
      daaInterface.encodeFunctionData("setVotingSlot", [
        formValues.proposedBlockNumber,
      ]),
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
