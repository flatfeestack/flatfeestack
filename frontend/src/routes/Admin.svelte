<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { API } from "../ts/api";
  import Spinner from "../components/Spinner.svelte";
  import { formatDate, formatNowUTC, storeToken } from "../ts/services";
  import { config, error, loadedSponsoredRepos, user } from "../ts/mainStore";
  import {
    faSignInAlt,
    faCheck,
    faArrowsLeftRight,
  } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import type { Repos } from "../types/users";
  import Dots from "../components/Dots.svelte";

  //let promisePendingPayouts =API.payouts.payoutInfos();
  let promiseTime = API.payouts.time();
  let promiseUsers = API.admin.users();
  let showSuccess = false;

  /* Search */
  let search = "";
  //TODO: if we use github only, then the search name is unique and we don't need to spli
  //in case of change, make sure you split the repo according to the link id
  let repos: Repos[] = [];
  let isSearchSubmitting = false;
  let linkGitUrl = "";
  let rootUuid = null;
  let warning = "";
  const handleSearch = async () => {
    try {
      isSearchSubmitting = true;
      repos = await API.repos.searchName(search);
      if (repos.length == 0) {
        warning = "Repo [" + search + "] not found";
      } else {
        rootUuid = repos.find((e) => e.link === e.uuid).uuid;
      }
    } catch (e) {
      $error = e;
    } finally {
      isSearchSubmitting = false;
    }
  };
  async function handleLinkGitUrl() {
    try {
      repos = await API.repos.linkGitUrl(rootUuid, linkGitUrl);
      linkGitUrl = "";
    } catch (e) {
      $error = e;
    }
  }
  async function makeRoot(repoId: string) {
    try {
      repos = await API.repos.makeRoot(repoId, rootUuid);
    } catch (e) {
      $error = e;
    }
  }

  const handleFakeUsers = async (email: string) => {
    return API.payouts.fakeUser(email);
  };

  const handleFakePayment = async (email: string, seats: number) => {
    return API.payouts.fakePayment(email, seats);
  };

  const handleFakeContribution = async (json: string) => {
    return API.payouts.fakeContribution(JSON.parse(json));
  };

  const handleWarp = async (hours: number) => {
    const p1 = API.user.timeWarp(hours);
    const p2 = API.authToken.timeWarp(hours);
    const p3 = refresh();

    const res = await p2;
    storeToken(res);
    await p1;
    await p3;
  };

  const payout = async (exchangeRate: number) => {
    const res = await API.payouts.payout(exchangeRate);
    if (res.ok) {
      showSuccess = true;
    }
  };

  let userEmail = "";
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
</style>

<Navigation>
  <h2 class="px-2">Time</h2>
  <div class="container">
    {#await promiseTime}
      Time on the backend / UTC: ...<br />
      Time on the frontend / UTC: {formatNowUTC()}
    {:then res}
      Time on the backend / UTC: {res.time}<br />
      Time on the frontend / UTC: {formatNowUTC()}
    {/await}
  </div>

  {#if $config.env == "local" || $config.env == "dev"}
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
              {#if $user.email !== row}
                <tr>
                  <td>{row}</td>
                  <td
                    ><button
                      class="accessible-btn"
                      on:click={() => loginAs(row)}
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

  <h2 class="px-2">Link Repos</h2>
  <div class="p-2 m-2">
    <form class="flex" on:submit|preventDefault={handleSearch}>
      <input type="text" bind:value={search} />
      <button class="button1" type="submit" disabled={isSearchSubmitting}
        >Search{#if isSearchSubmitting}<Dots />{/if}</button
      >
    </form>
  </div>
  <div class="container">
    {#each repos as repos2, key (repos2.uuid)}
      <div>
        <table>
          <thead>
            <tr>
              <th>URL</th>
              <th>Git URL</th>
              <th>Source</th>
              <th>Root</th>
            </tr>
          </thead>
          {#each repos2.repos as repo, key (repo.uuid)}
            <tr>
              <td>{repo.url}</td>
              <td>{repo.gitUrl}</td>
              <td>{repo.source}</td>
              <td>
                {#if repo.uuid !== repos2.uuid}
                  <button
                    class="accessible-btn"
                    on:click={() => makeRoot(repo.uuid)}
                  >
                    <Fa icon={faArrowsLeftRight} size="md" />
                  </button>
                {:else}
                  <Fa icon={faCheck} size="md" />
                {/if}
              </td>
            </tr>
          {/each}
          <tr>
            <td colspan="4">
              <form class="flex" on:submit|preventDefault={handleLinkGitUrl}>
                <input
                  input-size="32"
                  id="address-input"
                  name="address"
                  type="text"
                  bind:value={linkGitUrl}
                  placeholder="Add Git URL"
                />
                <button class="button1" type="submit">Link Git URL</button>
              </form>
            </td>
          </tr>
        </table>
      </div>
    {/each}
    {#if warning != ""}{warning}{/if}
  </div>

  <h2>Fake User</h2>
  <button
    class="button2 py-2 px-3 bg-primary-500 rounded-md text-white"
    on:click={() => handleFakeUsers(userEmail)}>Add Fake User</button
  >
  Email: <input bind:value={userEmail} />
  <h2>Fake Payment</h2>
  <button
    class="button2 py-2 px-3 bg-primary-500 rounded-md text-white"
    on:click={() => handleFakePayment(userEmail, seats)}
    >Add Fake Payment</button
  >
  Email: <input bind:value={userEmail} /> Seats: <input bind:value={seats} />
  <h2>Fake Contribution</h2>
  <button
    class="button2 py-2 px-3 bg-primary-500 rounded-md text-white"
    on:click={() => handleFakeContribution(json)}>Add Fake Contribution</button
  >
  Email: <textarea bind:value={json} rows="10" cols="50" />

  <h2>Payout Action</h2>
  <button
    class="button2 py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75"
    on:click={() => payout(exchangeRate)}
  >
    Payout
  </button>
  Exchange Rate USD to ETH: <input bind:value={exchangeRate} />

  {#if showSuccess}
    <div class="p-2">Payment successful!</div>
  {/if}
</Navigation>
