<script lang="ts">
import Navigation from "../components/Navigation.svelte";
import { API } from "../ts/api";
import Spinner from "../components/Spinner.svelte";
import { formatDate, storeToken } from "../ts/services";
import { config, error, loadedSponsoredRepos, user } from "../ts/store";
import { faSignInAlt } from "@fortawesome/free-solid-svg-icons";
import Fa from "svelte-fa";

let promisePendingPayouts =API.payouts.pending("pending");
let promisePaidPayouts = API.payouts.pending("paid");
let promiseLimboPayouts= API.payouts.pending("limbo");
let promiseTime = API.payouts.time();
let promiseUsers = API.admin.users();

const handleFakeUsers = async (email: string) => {
  return API.payouts.fakeUser(email)
}

const handleFakePayment= async (email: string, seats:number) => {
  return API.payouts.fakePayment(email, seats)
}

const handleFakeContribution= async(json: string )=> {
  return API.payouts.fakeContribution(JSON.parse(json))
}

const handleWarp = async (hours: number) => {
  await API.user.timeWarp(hours);
  await API.authToken.timeWarp(hours);
  await refresh();
}

const payout = async (exchangeRate: number) => {
  await API.payouts.payout(exchangeRate)
}

let userEmail = ""
let exchangeRate = 0.0;
let seats = 1;

const d = new Date();
const datestring1 = formatDate(d);
d.setMonth(d.getMonth() -1)
const datestring2 = formatDate(d);

let json = `{
"startDate":"`+datestring2+`",
"endDate":"`+datestring1+`",
"name":"##name##",
"weights": [
 {"email":"tom@tom","weight":0.5},
 {"email":"sam@sam","weight":0.4}
]}`;

const refresh = async () => {
  promiseTime = API.payouts.time();
  promisePendingPayouts = API.payouts.pending("pending");
  promisePaidPayouts = API.payouts.pending("paid");
  promiseLimboPayouts= API.payouts.pending("limbo");
  promiseUsers = API.admin.users();
}

async function loginAs(email: string) {
  try {
    const res = await API.authToken.loginAs(email)
    storeToken(res);
    const u = await API.user.get();
    user.set(u);
    loadedSponsoredRepos.set(false);
  } catch (e) {
    $error = e;
  }
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


<Navigation>
  <h1 class="px-2">Admin</h1>
  <div class="container">
  {#await promiseTime}
    Time on the backend / UTC: ...
  {:then res}
    Time on the backend / UTC: {res.time}
  {/await}
  </div>

  {#if $config.env == "local"}
    <h2 class="px-2">Timewarp</h2>
    <div class="container">
      <button class="button2 m-2" on:click={() => handleWarp(1)}>
        Timewarp 1 hour
      </button>
      <button class="button2 m-2" on:click={() => handleWarp(24)}>
        Timewarp 1 day
      </button>
      <button class="button2 m-2" on:click={() => handleWarp(160)}>
        Timewarp 1 week
      </button>
      <button class="button2 m-2" on:click={() => handleWarp(600)}>
        Timewarp 25 days
      </button>
      <button class="button2 m-2" on:click={() => handleWarp(8640)}>
        Timewarp 360 days year
      </button>
    </div>
  {/if}

  <h2 class="px-2">Login as User</h2>
  <div class="container">
    {#await promiseUsers}
      <Spinner />
    {:then users}
      <table>
        <thead>
        <tr>
          <th>Email</th>
          <th>Enter</th>
        </tr>
        </thead>
        <tbody>
        {#if users && users.length > 1}
          {#each users as row}
            {#if $user.email !== row.email}
              <tr>
                <td>{row.email}</td>
                <td><span class="cursor-pointer" on:click="{() => loginAs(row.email)}">
                  <Fa icon="{faSignInAlt}" size="md" /></span>
                </td>
              </tr>
            {/if}
          {/each}
        {:else}
          <tr>
            <td colspan="2">No Data</td>
          </tr>
        {/if}
        </tbody>
      </table>
    {:catch err}
      {error.set(err)}
    {/await}
  </div>






  <h2>Fake User</h2>
  <button class="button2 py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleFakeUsers(userEmail)}>Add Fake User</button>
  Email: <input bind:value={userEmail}>
  <h2>Fake Payment</h2>
  <button class="button2 py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleFakePayment(userEmail, seats)}>Add Fake Payment</button>
  Email: <input bind:value={userEmail}> Seats: <input bind:value={seats}>
  <h2>Fake Contribution</h2>
  <button class="button2 py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleFakeContribution(json)}>Add Fake Contribution</button>
  Email: <textarea bind:value={json} rows="10" cols="50"></textarea>



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
      {#if res && res.length > 0}
        {#each res as row}
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
  {:catch err}
    {error.set(err)}
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
      {#if res && res.length > 0}
        {#each res as row}
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
  {:catch err}
    {error.set(err)}
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
      {#if res && res.length > 0}
        {#each res as row}
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
  {:catch err}
    {error.set(err)}
  {/await}

  <h2>Payout Action</h2>
  <button class="button2 py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={() => payout(exchangeRate)}>
    Payout
  </button>
  Exchange Rate: <input bind:value={exchangeRate}>

</Navigation>
