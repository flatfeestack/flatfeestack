<script lang="ts">
import DashboardLayout from "../../layout/DashboardLayout.svelte";
import { API } from "src/api/api";
import Spinner from "../../components/UI/Spinner.svelte";
import ExchangeEntry from "../../components/ExchangeEntry.svelte";

let error: string;

const fetchData = async () => {
  try {
    const res = await API.exchanges.get();
    return res.data;
  } catch (e) {
    console.log(e);
    error = e;
  }
};
let promise = fetchData();

const updateExchange = async () => {
  try {
  } catch (e) {
    console.log(e);
  }
};
</script>

<DashboardLayout>
  <h1>Admin</h1>
  {#if error}
    <p class="p-2 bg-red-500 text-white inline-block">{error}</p>
  {/if}
  {#await promise}
    <Spinner />
  {:then res}
    {#each res as exchange}
      <ExchangeEntry exchange="{exchange}" />
    {/each}
  {/await}
</DashboardLayout>
