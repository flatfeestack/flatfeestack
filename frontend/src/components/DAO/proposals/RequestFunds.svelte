<script lang="ts">
  import { ethers } from "ethers";
  import { walletContract } from "../../../ts/daoStore";
  import type { ProposalFormProps } from "../../../types/dao";
  import yup from "../../../utils/yup";

  // as described in https://github.com/sveltejs/svelte/issues/7605#issuecomment-1156000553
  interface $$Props extends ProposalFormProps {}
  export let calls: $$Props["calls"];

  const currencies = ["ETH", "Wei"];

  let formValues = {
    amount: 0,
    selectedCurrency: 0,
    targetWalletAddress: "",
  };

  const schema = yup.object().shape({
    amount: yup.number().min(0).required(),
    targetWalletAddress: yup.string().isEthereumAddress().required(),
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
        target: $walletContract?.address,
        transferCallData: $walletContract?.interface.encodeFunctionData(
          "increaseAllowance",
          [
            formValues.targetWalletAddress,
            ethers.utils.parseUnits(
              String(formValues.amount),
              currencies[formValues.selectedCurrency] === "ETH" ? 18 : 1
            ),
          ]
        ),
        value: 0,
      },
    ];
  }
</script>

<style>
  .invalid {
    grid-column: 1 / 3;
    margin: 0;
    text-align: right;
  }
</style>

<label for="targetWalletAddress">Target wallet address</label>
<input
  type="text"
  name="targetWalletAddress"
  bind:value={formValues.targetWalletAddress}
  required
/>
{#await schema.validateAt("targetWalletAddress", formValues) catch error}
  <p class="invalid" style="color:red">{error.errors[0]}</p>
{/await}

<label for="amount">Amount</label>
<div>
  <input
    type="number"
    min="0"
    step="any"
    name="amount"
    bind:value={formValues.amount}
    required
  />
  <select bind:value={formValues.selectedCurrency}>
    {#each currencies as currency, index}
      <option value={index}>
        {currency}
      </option>
    {/each}
  </select>
</div>
{#await schema.validateAt("amount", formValues) catch error}
  <p class="invalid" style="color:red">{error.errors[0]}</p>
{/await}
