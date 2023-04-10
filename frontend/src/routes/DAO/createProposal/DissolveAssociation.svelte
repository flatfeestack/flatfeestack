<script lang="ts">
  import { Interface } from "ethers/lib/utils";
  import { MembershipABI } from "$lib/contracts/Membership";
  import { membershipContract } from "$lib/ts/daoStore";
  import type { ProposalFormProps } from "$lib/types/dao";
  import truncateEthAddress from "$lib/utils/truncateEthereumAddress";
  import yup from "$lib/utils/yup";
  import Spinner from "$lib/components/Spinner.svelte";
  import { WalletABI } from "$lib/contracts/Wallet";
  import { DAOABI } from "$lib/contracts/DAO";

  interface $$Props extends ProposalFormProps {}
  export let calls: $$Props["calls"];

  let formValues = {
    liquidator: "",
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
  const walletInterface = new Interface(WalletABI);
  const daoInterface = new Interface(DAOABI);

  const schema = yup.object().shape({
    liquidator: yup.string().isEthereumAddress().required(),
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
        target: import.meta.env.VITE_WALLET_CONTRACT_ADDRESS,
        transferCallData: walletInterface.encodeFunctionData("liquidate", [
          formValues.liquidator,
        ]),
        value: 0,
      },
      {
        target: import.meta.env.VITE_MEMBERSHIP_CONTRACT_ADDRESS,
        transferCallData:
          membershipInterface.encodeFunctionData("lockMembership"),
        value: 0,
      },
      {
        target: import.meta.env.VITE_DAO_CONTRACT_ADDRESS,
        transferCallData: daoInterface.encodeFunctionData("dissolveDAO"),
        value: 0,
      },
    ];
  }
</script>

{#if isLoading}
  <Spinner />
{:else}
  <label for="liquidator">Liquidator</label>
  <select name="liquidator" id="liquidator" bind:value={formValues.liquidator}>
    {#each members as member}
      <option value={member}>
        {truncateEthAddress(member)}
      </option>
    {/each}
  </select>
  {#await schema.validateAt("liquidator", formValues) catch error}
    <p class="invalid" style="color:red">{error.errors[0]}</p>
  {/await}
{/if}
