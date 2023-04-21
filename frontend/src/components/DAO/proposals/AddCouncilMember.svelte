<script lang="ts">
  import { membershipContract } from "../../../ts/daoStore";
  import type { ProposalFormProps } from "../../../types/dao";
  import yup from "../../../utils/yup";

  interface $$Props extends ProposalFormProps {}
  export let calls: $$Props["calls"];

  let formValues = {
    proposedCouncilMember: "",
  };

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
    calls = [
      {
        target: $membershipContract?.address,
        transferCallData: $membershipContract?.interface.encodeFunctionData(
          "addCouncilMember",
          [formValues.proposedCouncilMember]
        ),
        value: 0,
      },
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
{#await schema.validateAt("proposedCouncilMember", formValues) catch error}
  <p class="invalid" style="color:red">{error.errors[0]}</p>
{/await}
