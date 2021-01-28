<script lang="ts">
import DashboardLayout from "../../layout/DashboardLayout.svelte";
import { API } from "src/api/api";
import Spinner from "../../components/UI/Spinner.svelte";
import ExchangeEntry from "../../components/ExchangeEntry.svelte";

let error: string;

const pendingPayouts = async () => {
  try {
    const res = await API.payouts.pending()
    console.log(res);
    return res;
  } catch (e) {
    console.log(e);
    error = e;
  }
}
let promisePayouts = pendingPayouts();

const time = async () => {
  const res = await API.payouts.time()
  return res;
}
let promiseTime = time();

const handleFakeUsers = async () => {
  return await API.payouts.fakeUser()
}

const handleWarp = async (hours: number) => {
  await API.payouts.timeWarp(hours)
  await API.auth.timeWarp(hours)
  promiseTime = time();
}

</script>

<DashboardLayout>
  <h1>Admin</h1>
  {#if error}
    <p class="p-2 bg-red-500 text-white inline-block">{error}</p>
  {/if}

  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={handleFakeUsers}>
    Insert 2 users and 2 repo
  </button>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={() => handleWarp(1)}>
    Timewarp 1 hour
  </button>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={() => handleWarp(24)}>
    Timewarp 1 day
  </button>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={() => handleWarp(160)}>
    Timewarp 1 week
  </button>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={() => handleWarp(600)}>
    Timewarp 25 days
  </button>

  {#await promiseTime}
    <Spinner />
  {:then res}
    Status: {res.status}, Server Datetime: {res.data.time}
  {/await}

  {#await promisePayouts}
    <Spinner />
  {:then res}
    {res.status}
    {#each res.data as pending}
      {pending.email_list}
    {/each}
  {/await}
</DashboardLayout>
