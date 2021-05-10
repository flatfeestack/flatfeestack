<script lang="ts">
import Navigation from "../components/Navigation.svelte";
import { API } from "../ts/api";
import Spinner from "../components/Spinner.svelte";

let promisePendingPayouts =API.payouts.pending("pending");
let promisePaidPayouts = API.payouts.pending("paid");
let promiseLimboPayouts= API.payouts.pending("limbo");
let promiseTime = API.payouts.time();

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
  console.log("PPPPPAAAYYYY")
  await API.payouts.payout(exchangeRate)
}

let userEmail = ""
let exchangeRate = 0.0;
let seats = 1;

//https://stackoverflow.com/questions/3552461/how-to-format-a-javascript-date
const d = new Date();
const datestring1 = d.getFullYear() + "-" + ("0"+(d.getMonth()+1)).slice(-2) + "-" +
  ("0" + d.getDate()).slice(-2) + " " + ("0" + d.getHours()).slice(-2) + ":" + ("0" + d.getMinutes()).slice(-2);


const datestring2 = d.getFullYear()  + "-" + ("0"+(d.getMonth())).slice(-2) + "-" +
  ("0" + d.getDate()).slice(-2) + " " + ("0" + d.getHours()).slice(-2) + ":" + ("0" + d.getMinutes()).slice(-2);


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
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleWarp(1)}>
    Timewarp 1 hour
  </button>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleWarp(24)}>
    Timewarp 1 day
  </button>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleWarp(160)}>
    Timewarp 1 week
  </button>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleWarp(600)}>
    Timewarp 25 days
  </button>

  <h2>Fake User</h2>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleFakeUsers(userEmail)}>Add Fake User</button>
  Email: <input bind:value={userEmail}>
  <h2>Fake Payment</h2>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleFakePayment(userEmail, seats)}>Add Fake Payment</button>
  Email: <input bind:value={userEmail}> Seats: <input bind:value={seats}>
  <h2>Fake Contribution</h2>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white" on:click={() => handleFakeContribution(json)}>Add Fake Contribution</button>
  Email: <textarea bind:value={json} rows="10" cols="50"></textarea>

  {#await promiseTime}
    <h2>Time: ...</h2>
  {:then res}
    <h2>Time: {res.time}</h2>
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
  {:catch error}
    {error}
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
  {/await}

  <h2>Payout Action</h2>
  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={() => payout(exchangeRate)}>
    Payout
  </button>
  Exchange Rate: <input bind:value={exchangeRate}>

</Navigation>
