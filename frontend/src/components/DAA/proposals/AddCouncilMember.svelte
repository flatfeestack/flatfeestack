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
    proposedCouncilMember: "",
  };
  const membershipInterface = new Interface(MembershipABI);

  const schema = yup.object().shape({
    proposedCouncilMember: yup.string().isEthereumAddress().required(),
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
      membershipInterface.encodeFunctionData("addCouncilMember", [
        formValues.proposedCouncilMember,
      ]),
    ];
  }
</script>

<label for="proposedCouncilMember">Proposed council member</label>
<input
  type="text"
  id="proposedCouncilMember"
  name="proposedCouncilMember"
  bind:value={formValues.proposedCouncilMember}
  required
/>
{#await schema.validateAt("proposedCouncilMember", formValues)}{:catch error}
  <p class="invalid" style="color:red">{error.errors[0]}</p>
{/await}
