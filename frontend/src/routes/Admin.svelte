<script lang="ts">
  import { faSignInAlt } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import { navigate } from "svelte-routing";
  import Navigation from "../components/Navigation.svelte";
  import Spinner from "../components/Spinner.svelte";
  import { API } from "../ts/api";
  import { config, error, loadedSponsoredRepos, user } from "../ts/mainStore";
  import { formatDate, formatNowUTC, storeToken } from "../ts/services";

  //let promisePendingPayouts =API.payouts.payoutInfos();
  let promiseTime = API.payouts.time();
  let promiseUsers = API.admin.users();
  let showSuccess = false;

  const handleFakeUsers = async () => {
    try {
      await API.payouts.fakeUser(fakeUserEmail);
    } catch (e) {
      $error = e;
    }
  };

  const handleFakePayment = async () => {
    try {
      await API.payouts.fakePayment(fakePaymentEmail, seats);
    } catch (e) {
      $error = e;
    }
  };

  const handleFakeContribution = async () => {
    try {
      await API.payouts.fakeContribution(JSON.parse(json));
    } catch (e) {
      $error = e;
    }
  };

  const handleWarp = async (hours: number) => {
    try {
      const p1 = API.user.timeWarp(hours);
      const p2 = API.authToken.timeWarp(hours);
      const p3 = refresh();

      const res = await p2;
      storeToken(res);
      await p1;
      await p3;
    } catch (e) {
      $error = e;
    }
  };

  const payout = async () => {
    try {
      const res = await API.payouts.payout(exchangeRate);
      if (res.ok) {
        showSuccess = true;
      }
    } catch (e) {
      $error = e;
    }
  };

  let fakeUserEmail = "";
  let fakePaymentEmail = "";
  let exchangeRate = 0.0;
  let seats = 1;

  const d = new Date();
  const datestring1 = formatDate(d);
  d.setMonth(d.getMonth() - 1);
  const datestring2 = formatDate(d);

  let json =
    `{
"startDate":"` +
    datestring2 +
    `",
"endDate":"` +
    datestring1 +
    `",
"name":"##name##",
"weights": [
 {"email":"tom@tom","weight":0.5},
 {"email":"sam@sam","weight":0.4}
]}`;

  const refresh = async () => {
    promiseTime = API.payouts.time();
    //promisePendingPayouts = API.payouts.pending("pending");
    //promisePaidPayouts = API.payouts.pending("paid");
    //promiseLimboPayouts= API.payouts.pending("limbo");
    promiseUsers = API.admin.users();
  };

  async function loginAs(email: string) {
    try {
      const res = await API.authToken.loginAs(email);
      storeToken(res);
      const u = await API.user.get();
      user.set(u);
      loadedSponsoredRepos.set(false);
      navigate("/user/search");
    } catch (e) {
      $error = e;
    }
  }
</script>

<style>
  table,
  th,
  td {
    border: 1px solid black;
    border-collapse: collapse;
  }
  table {
    background: #eee;
    width: 50%;
    text-align: center;
  }
  .form-single label,
  input {
    display: inline-block;
  }
  .form-single button {
    display: block;
  }
  .ml-2 {
    margin-left: 0.5rem;
  }
  .mr-2 {
    margin-right: 0.5rem;
  }
  .mb-2 {
    margin-bottom: 0.5rem;
  }
</style>

<Navigation>
  <h2 class="p-2 m-2">Time</h2>
  <div class="container m-2 p-2">
    {#await promiseTime}
      Time on the backend / UTC: ...<br />
      Time on the frontend / UTC: {formatNowUTC()}
    {:then res}
      Time on the backend / UTC: {res.time}<br />
      Time on the frontend / UTC: {formatNowUTC()}
    {/await}
  </div>

  {#if $config.env == "local" || $config.env == "dev"}
    <h2 class="p-2 m-2">Timewarp</h2>
    <div class="container">
      <button class="button1 m-2" on:click={() => handleWarp(1)}>
        Timewarp 1 hour
      </button>
      <button class="button1 m-2" on:click={() => handleWarp(24)}>
        Timewarp 1 day
      </button>
      <button class="button1 m-2" on:click={() => handleWarp(160)}>
        Timewarp 1 week
      </button>
      <button class="button1 m-2" on:click={() => handleWarp(600)}>
        Timewarp 25 days
      </button>
      <button class="button1 m-2" on:click={() => handleWarp(8640)}>
        Timewarp 360 days year
      </button>
    </div>
  {/if}

  <h2 class="p-2 m-2">Login as User</h2>
  <div class="container m-2 p-2">
    {#await promiseUsers}
      <Spinner />
    {:then userEmails}
      <table>
        <thead>
          <tr>
            <th>Email</th>
            <th>Enter</th>
          </tr>
        </thead>
        <tbody>
          {#if userEmails && userEmails.length > 1}
            {#each userEmails as userEmail}
              {#if $user.email !== userEmail}
                <tr>
                  <td>{userEmail}</td>
                  <td
                    ><button
                      class="accessible-btn"
                      on:click={() => loginAs(userEmail)}
                    >
                      <Fa icon={faSignInAlt} size="md" /></button
                    >
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

  <h2 class="p-2 m-2">Fake User</h2>
  <div class="container m-2 p-2">
    <form class="flex form-single" on:submit|preventDefault={handleFakeUsers}>
      <label class="mr-2" for="fake-user">Email:</label>
      <input
        class="ml-2"
        name="fake-user"
        type="text"
        bind:value={fakeUserEmail}
      />
      <button class="button1 mt-2 mb-2" type="submit">Add Fake User</button>
    </form>
  </div>

  <h2 class="p-2 m-2">Fake Payment</h2>
  <div class="container m-2 p-2">
    <form class="flex form-single" on:submit|preventDefault={handleFakePayment}>
      <div class="mt-2 mb-2">
        <label class="mr-2" for="fake-payment-email">Email:</label>
        <input
          class="ml-2"
          type="text"
          name="fake-payment-email"
          bind:value={fakePaymentEmail}
        />
      </div>
      <div class="mt-2 mb-2">
        <label class="mr-2" for="fake-payment-seats">Seats:</label>
        <input
          class="ml-2"
          type="text"
          name="fake-payment-seats"
          bind:value={seats}
        />
      </div>

      <button class="button1 mt-2 mb-2" type="submit">Add Fake Payment</button>
    </form>
  </div>

  <h2 class="p-2 m-2">Fake Contribution</h2>
  <div class="container m-2 p-2">
    <form
      class="flex form-single"
      on:submit|preventDefault={handleFakeContribution}
    >
      <label class="mr-2" for="fake-contribution">Contribution:</label>
      <textarea
        name="fake-contribution"
        class="ml-2"
        bind:value={json}
        rows="10"
        cols="50"
      />
      <button class="button1 mt-2 mb-2" type="submit"
        >Add Fake Contribution</button
      >
    </form>
  </div>

  <h2 class="p-2 m-2">Payout Action</h2>
  <div class="container m-2 p-2">
    <form class="flex form-single" on:submit|preventDefault={payout}>
      <label class="mr-2" for="fake-payout">Exchange Rate USD to ETH: </label>
      <input
        class="ml-2"
        name="fake-payout"
        type="text"
        bind:value={exchangeRate}
      />

      <button class="button1 mt-2 mb-2 disabled:opacity-75" type="submit">
        Payout
      </button>

      {#if showSuccess}
        <div class="p-2 m-2">Payment successful!</div>
      {/if}
    </form>
  </div>
</Navigation>
