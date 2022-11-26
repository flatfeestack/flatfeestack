<script lang="ts">
  import { ethers } from "ethers";
  import { Interface } from "ethers/lib/utils";
  import yup from "../../../utils/yup";
  import { MembershipABI } from "../../../contracts/Membership";
  import type { ProposalFormProps } from "../../../types/daa";

  interface $$Props extends ProposalFormProps {}

  export let targets: $$Props["targets"];
  export let values: $$Props["values"];
  export let transferCallData: $$Props["transferCallData"];

  let formValues = {
    proposedChairman: "",
  };
  const membershipInterface = new Interface(MembershipABI);

  const schema = yup.object().shape({
    proposedChairman: yup.string().isEthereumAddress().required(),
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
    targets = [import.meta.env.VITE_MEMBERSHIP_CONTRACT_ADDRESS];
    transferCallData = [
      membershipInterface.encodeFunctionData("setChairman", [
        formValues.proposedChairman,
      ]),
    ];
  }
</script>

<label for="proposedChairman">Proposed chairman</label>
<input
  type="text"
  id="proposedChairman"
  name="proposedChairman"
  bind:value={formValues.proposedChairman}
  required
/>
{#await schema.validateAt("proposedChairman", formValues)}{:catch error}
  <p class="invalid" style="color:red">{error.errors[0]}</p>
{/await}
