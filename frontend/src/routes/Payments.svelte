<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error } from "../ts/mainStore";
  import { API } from "../ts/api";
  import { onMount, onDestroy } from "svelte";
  import type { Repo, UserBalance } from "../types/backend";
  import { formatDate, formatBalance } from "../ts/services";
  import PaymentSelection from "../components/PaymentSelection.svelte";

  let sponsoredRepos: Repo[] = [];
  let userBalances: UserBalance[] = [];
  let intervalId;

  const fetchData = async () => {
    userBalances = await API.user.userBalance();
  };

  onMount(async () => {
    try {
      // Fetch data immediately on component mount
      const pr1 = fetchData();
      const pr2 = API.user.getSponsored();
      const res2 = await pr2;
      sponsoredRepos = res2 || [];
      await pr1;
      intervalId = setInterval(fetchData, 5000); // Poll every 5 seconds
    } catch (e) {
      $error = e;
    }
  });

  onDestroy(() => {
    clearInterval(intervalId); // Clear interval on component unmount to prevent memory leaks
  });

</script>

<Navigation>
  <h2 class="p-2 m-2">Sponsor Summary</h2>

  <div class="grid-2">
    <p class="nobreak">
      Selected Projects: <span class="bold m-4"
        >{sponsoredRepos.length} projects</span
      >
    </p>
  </div>

  <PaymentSelection />

  {#if userBalances }
    <h2 class="p-2 m-2">Payment History</h2>
    <div class="container">
      <table>
        <thead>
          <tr>
            <th>Balance</th>
            <th>Currency</th>
          </tr>
        </thead>
        <tbody>
          {#each userBalances as row}
            <tr>
              <td>{formatBalance(row.balance, row.currency)}</td>
              <td>{row.currency}</td>
            </tr>
          {:else}
            <tr>
              <td colspan="5">No Data</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</Navigation>
