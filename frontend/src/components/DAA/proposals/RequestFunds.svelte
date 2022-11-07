<script lang="ts">
  import { ethers } from "ethers";
  import { Interface } from "ethers/lib/utils";
  import yup from "../../../utils/yup";
  import { WalletABI } from "../../../contracts/Wallet";
  import type { ProposalFormProps } from "../../../types/daa";

  // as described in https://github.com/sveltejs/svelte/issues/7605#issuecomment-1156000553
  interface $$Props extends ProposalFormProps {}

  export let targets: $$Props["targets"];
  export let values: $$Props["values"];
  export let transferCallData: $$Props["transferCallData"];

  const currencies = ["ETH", "Wei"];
  const walletInterface = new Interface(WalletABI);

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
    values = [0];
    targets = [import.meta.env.VITE_WALLET_CONTRACT_ADDRESS];
    transferCallData = walletInterface.encodeFunctionData("increaseAllowance", [
      formValues.targetWalletAddress,
      ethers.utils.parseUnits(
        String(formValues.amount),
        currencies[formValues.selectedCurrency] === "ETH" ? 18 : 1
      ),
    ]);
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
{#await schema.validateAt("targetWalletAddress", formValues)}{:catch error}
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
{#await schema.validateAt("amount", formValues)}{:catch error}
  <p class="invalid" style="color:red">{error.errors[0]}</p>
{/await}
