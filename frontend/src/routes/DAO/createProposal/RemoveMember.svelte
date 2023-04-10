<script lang="ts">
  import { Interface } from "ethers/lib/utils";
  import { MembershipABI } from "$lib/contracts/Membership";
  import { membershipContract } from "$lib/ts/daoStore";
  import type { ProposalFormProps } from "$lib/types/dao";
  import truncateEthAddress from "$lib/utils/truncateEthereumAddress";
  import yup from "$lib/utils/yup";
  import Spinner from "$lib/components/Spinner.svelte";

  interface $$Props extends ProposalFormProps {}
  export let calls: $$Props["calls"];

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
    calls = [
      {
        target: import.meta.env.VITE_MEMBERSHIP_CONTRACT_ADDRESS,
        transferCallData: membershipInterface.encodeFunctionData(
          "removeMember",
          [formValues.memberToBeRemoved]
        ),
        value: 0,
      },
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
  {#await schema.validateAt("memberToBeRemoved", formValues) catch error}
    <p class="invalid" style="color:red">{error.errors[0]}</p>
  {/await}
{/if}
