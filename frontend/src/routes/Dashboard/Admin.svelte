<script lang="ts">
import DashboardLayout from "./DashboardLayout.svelte";
import { API } from "./../../ts/api";
import Spinner from "../../components/Spinner.svelte";

let promisePendingPayouts =API.payouts.pending("pending");
let promisePaidPayouts = API.payouts.pending("paid");
let promiseLimboPayouts= API.payouts.pending("limbo");
let promiseTime = API.payouts.time();

const handleFakeUsers = async () => {
  return await API.payouts.fakeUser()
}

const handleWarp = async (hours: number) => {
  await API.user.timeWarp(hours);
  await API.authToken.timeWarp(hours);
  await refresh();
}

const payout = async (exchangeRate: number) => {
  await API.payouts.payout(exchangeRate)
}

let exchangeRate = 0.0;

const refresh = async () => {
  promiseTime = API.payouts.time();
  promisePendingPayouts = API.payouts.pending("pending");
  promisePaidPayouts = API.payouts.pending("paid");
  promiseLimboPayouts= API.payouts.pending("limbo");
}

</script>

<style>
    table, th, td {
        border: 1px solid black;
        border-collapse: collapse;
    }
    table {
        background: #eee;
        width: 50%;
        text-align: center;
    }
</style>


<DashboardLayout>
  <h1 class="px-2">Admin</h1>

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
    <h2>Time: ...</h2>
  {:then res}
    <h2>Time: {res.data.time}</h2>
  {/await}

  <h2>Pending Payouts</h2>
  {#await promisePendingPayouts}
    <Spinner />
  {:then res}
    <table>
      <thead>
      <tr>
        <th>ETH Address</th>
        <th>µ&nbsp;USD</th>
        <th>Email(s)</th>
        <th>Monthly Repo ID(s)</th>
      </tr>
      </thead>
      <tbody>
      {#if res.data}
        {#each res.data as row}
          <tr>
            <td>{row.payout_eth}</td>
            <td>{row.balance}</td>
            <td>{row.email_list}</td>
            <td>{row.monthly_user_payout_id_list}</td>
          </tr>
        {:else}
          <tr><td colspan="4">No Data</td></tr>
        {/each}
      {:else}
        <tr><td colspan="4">No Data</td></tr>
      {/if}
      </tbody>
    </table>
  {/await}

  <h2>Paid Payouts</h2>
  {#await promisePaidPayouts}
    <Spinner />
  {:then res}
    <table>
      <thead>
      <tr>
        <th>ETH Address</th>
        <th>µ&nbsp;USD</th>
        <th>Email(s)</th>
        <th>Monthly Repo ID(s)</th>
      </tr>
      </thead>
      <tbody>
      {#if res.data}
        {#each res.data as row}
          <tr>
            <td>{row.payout_eth}</td>
            <td>{row.balance}</td>
            <td>{row.email_list}</td>
            <td>{row.monthly_user_payout_id_list}</td>
          </tr>
        {:else}
          <tr><td colspan="4">No Data</td></tr>
        {/each}
      {:else}
        <tr><td colspan="4">No Data</td></tr>
      {/if}
      </tbody>
    </table>
  {/await}

  <h2>Limbo Payouts</h2>
  {#await promiseLimboPayouts}
    <Spinner />
  {:then res}
    <table>
      <thead>
      <tr>
        <th>ETH Address</th>
        <th>µ&nbsp;USD</th>
        <th>Email(s)</th>
        <th>Monthly Repo ID(s)</th>
      </tr>
      </thead>
      <tbody>
      {#if res.data}
        {#each res.data as row}
          <tr>
            <td>{row.payout_eth}</td>
            <td>{row.balance}</td>
            <td>{row.email_list}</td>
            <td>{row.monthly_user_payout_id_list}</td>
          </tr>
        {:else}
          <tr><td colspan="4">No Data</td></tr>
        {/each}
      {:else}
        <tr><td colspan="4">No Data</td></tr>
      {/if}
      </tbody>
    </table>
  {/await}

  <h2>Payout Action</h2>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={() => payout(exchangeRate)}>
    Payout
  </button>
  Exchange Rate: <input bind:value={exchangeRate}>

</DashboardLayout>
