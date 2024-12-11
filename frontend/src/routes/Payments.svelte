<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { user, error } from "../ts/mainStore";
  import { API } from "../ts/api";
  import { onMount, onDestroy } from "svelte";
  import type { UserBalance } from "../types/backend";
  import { formatBalance } from "../ts/services";
  import PaymentSelection from "../components/PaymentSelection.svelte";

  let userBalances: UserBalance[] = [];
  let foundationBalances: UserBalance[] = [];
  let intervalId: ReturnType<typeof setInterval>;

  const fetchData = async () => {
    userBalances = await API.user.userBalance();
    foundationBalances = await API.user.foundationBalance();
  };

  onMount(async () => {
    try {
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

<style>
  caption {
    white-space: nowrap;
  }

  .history-container {
    width: 50%;
  }

  .balance-title {
    margin: 0 0 1rem 0;
  }

  @media screen and (max-width: 600px) {
    table {
      width: 100%;
    }
  }
</style>

<Navigation>
  <PaymentSelection />

  <h2 class="p-2 m-2">Payment History</h2>
  <div class="container">
    {#if userBalances}
      <div class="container-col2 m-2 history-container">
        <h3 class="balance-title">Your Current Sponsoring Balance: __</h3>
        <div class="container">
          <table>
            <caption>Sponsoring History</caption>
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
      </div>
    {/if}
    {#if foundationBalances && $user.multiplier}
      <div class="container-col2 m-2 history-container">
        <h3 class="balance-title">Current Multiplier Sponsoring Balance: __</h3>
        <div class="container">
          <table>
            <caption>Multiplier Sponsoring History</caption>
            <thead>
              <tr>
                <th>Balance</th>
                <th>Currency</th>
              </tr>
            </thead>
            <tbody>
              {#each foundationBalances as row}
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
      </div>
    {/if}
  </div>
</Navigation>
