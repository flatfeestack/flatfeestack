<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error } from "../ts/mainStore";
  import { API } from "../ts/api";
  import { onMount, onDestroy } from "svelte";
  import type { UserBalance } from "../types/backend";
  import { formatBalance } from "../ts/services";
  import PaymentSelection from "../components/PaymentSelection.svelte";

  let userBalances: UserBalance[] = [];
  let intervalId: number;

  const fetchData = async () => {
    userBalances = await API.user.userBalance();
  };

  onMount(async () => {
    try {
      // Fetch data immediately on component mount
      await fetchData();
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
  <PaymentSelection />

  {#if userBalances}
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
