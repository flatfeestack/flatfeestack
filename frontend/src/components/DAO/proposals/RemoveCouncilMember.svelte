<script lang="ts">
  import type { BigNumber } from "ethers";
  import { councilMembers, membershipContract } from "../../../ts/daoStore";
  import type { ProposalFormProps } from "../../../types/dao";
  import truncateEthAddress from "../../../utils/truncateEthereumAddress";
  import yup from "../../../utils/yup";
  import Spinner from "../../Spinner.svelte";

  interface $$Props extends ProposalFormProps {}
  export let calls: $$Props["calls"];

  let formValues = {
    councilMemberToBeRemoved: "",
  };
  let isLoading = false;
  let minimumCouncilMembers = 0;

  const schema = yup.object().shape({
    councilMemberToBeRemoved: yup.string().isEthereumAddress().required(),
  });

  $: {
    if ($councilMembers === null) {
      isLoading = true;
    } else {
      if (minimumCouncilMembers === 0) {
        isLoading = true;
        setMinimumCouncilMembers();
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
  }

  async function setMinimumCouncilMembers() {
    const result: BigNumber = await $membershipContract.minimumCouncilMembers();
    minimumCouncilMembers = result.toNumber();
  }

  function updateCalldata() {
    calls = [
      {
        target: $membershipContract?.address,
        transferCallData: $membershipContract?.interface.encodeFunctionData(
          "removeCouncilMember",
          [formValues.councilMemberToBeRemoved]
        ),
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

{#if isLoading}
  <Spinner />
{:else}
  <label for="councilMemberToBeRemoved">Member to be removed</label>
  <select
    name="councilMemberToBeRemoved"
    id="councilMemberToBeRemoved"
    bind:value={formValues.councilMemberToBeRemoved}
  >
    {#each $councilMembers as councilMember}
      <option value={String(councilMember)}>
        {truncateEthAddress(String(councilMember))}
      </option>
    {/each}
  </select>

  {#await schema.validateAt("councilMemberToBeRemoved", formValues) catch error}
    <p class="invalid" style="color:red">{error.errors[0]}</p>
  {/await}

  {#if minimumCouncilMembers > 0}
    <p class="combine-column italic">
      Note that the DAO requires at least {minimumCouncilMembers} council members.
      There won't be a validation if your proposal will result in less council members
      than the minimum amount, as there could be another proposal pending that will
      add an additional council member. But be aware of it as the execution of the
      proposal might fail.
    </p>
  {/if}
{/if}
