<script lang="ts">
import { API } from "../api/api";
import { onMount } from "svelte";
let address = "";
let error;

async function update(e) {
  e.preventDefault();
  try {
    if (!address || !address.match(/^0x[a-fA-F0-9]{40}$/g)) {
      console.log("error not valid eth address");
      error = "Invalid ethereum address";
    }
    await API.user.updatePayoutAddress({ chain_id: "ETH", address });
    error = "";
  } catch (e) {
    error = String(e);
    console.log(e);
  }
}

async function fetchAddress() {
  try {
    const res = await API.user.getPayoutAddresses();

    const addr = res.data.data?.find((a) => a.chain_id === "ETH");
    if (addr) {
      address = addr.address;
    }
  } catch (e) {
    console.log(e);
  }
}

onMount(async () => {
  await fetchAddress();
});
</script>

<form class="flex items-end" on:submit="{update}">
  <div class="w-64">
    <label class="block text-grey-darker text-sm font-bold mb-2 w-full">Ethereum
      Address</label>
    <input type="text" step="0.0001" class="input" bind:value="{address}" />
  </div>
  <div><button type="submit" class="button ml-5">Update Address</button></div>
</form>
{#if error}
  <div class="flex mt-5">
    <p class="bg-red-500 p-2 block text-white rounded">{error}</p>
  </div>
{/if}
