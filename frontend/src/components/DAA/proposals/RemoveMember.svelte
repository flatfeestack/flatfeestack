<script lang="ts">
  import { Interface } from "ethers/lib/utils";
  import { MembershipABI } from "../../../contracts/Membership";
  import { membershipContract } from "../../../ts/daaStore";
  import type { ProposalFormProps } from "../../../types/daa";
  import truncateEthAddress from "../../../utils/truncateEthereumAddress";
  import yup from "../../../utils/yup";
  import Spinner from "../../Spinner.svelte";

  interface $$Props extends ProposalFormProps {}

  export let targets: $$Props["targets"];
  export let values: $$Props["values"];
  export let transferCallData: $$Props["transferCallData"];

  let formValues = {
    memberToBeRemoved: "",
  };
  let isLoading = true;
  let members: string[] | null = null;

  $: {
    if ($membershipContract === null) {
      isLoading = true;
    } else if (members === null) {
      prepareView();
    }
  }

  async function prepareView() {
    const amountOfMembers = await $membershipContract.getMembersLength();

    members = await Promise.all(
      [...Array(amountOfMembers.toNumber()).keys()].map(
        async (index: Number) => {
          return await $membershipContract.members(index);
        }
      )
    );

    isLoading = false;
  }

  const membershipInterface = new Interface(MembershipABI);

  const schema = yup.object().shape({
    memberToBeRemoved: yup.string().isEthereumAddress().required(),
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
      membershipInterface.encodeFunctionData("removeMember", [
        formValues.memberToBeRemoved,
      ]),
    ];
  }
</script>

{#if isLoading}
  <Spinner />
{:else}
  <label for="memberToBeRemoved">Member to be removed</label>
  <select
    name="memberToBeRemoved"
    id="memberToBeRemoved"
    bind:value={formValues.memberToBeRemoved}
  >
    {#each members as member}
      <option value={member}>
        {truncateEthAddress(member)}
      </option>
    {/each}
  </select>
  {#await schema.validateAt("memberToBeRemoved", formValues)}{:catch error}
    <p class="invalid" style="color:red">{error.errors[0]}</p>
  {/await}
{/if}
