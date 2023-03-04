<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error, userBalances } from "../ts/mainStore";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import type { Repos } from "../types/users";
  import { connectWs, formatDate, formatBalance } from "../ts/services";
  import PaymentSelection from "../components/PaymentSelection.svelte";

  let sponsoredRepos: Repos[] = [];
  let invite_email;
  let isSubmitting = false;

  onMount(async () => {
    try {
      const pr1 = connectWs();
      const pr2 = API.user.getSponsored();
      const res2 = await pr2;
      sponsoredRepos = res2 === null ? [] : res2;
      await pr1;
    } catch (e) {
      $error = e;
    }
  });
</script>

<Navigation>
  <h2 class="p-2 m-2">Sponsor Summary</h2>

  <div class="grid-2">
    <p class="nobreak">Selected Projects: <span class="bold m-4">{sponsoredRepos.length} projects</span></p>
  </div>

  <PaymentSelection />

  {#if $userBalances && $userBalances.userBalances}
    <h2 class="p-2 m-2">Payment History</h2>
    <div class="container">
      <table>
        <thead>
          <tr>
            <th>Balance</th>
            <th>Currency</th>
            <th>Date</th>
            <th>Type</th>
          </tr>
        </thead>
        <tbody>
          {#each $userBalances.userBalances as row}
            <tr>
              <td>{formatBalance(row.balance, row.currency)}</td>
              <td>{row.currency}</td>
              <td>{formatDate(new Date(row.createdAt))}</td>
              <td title="Payment cycle Id: {row.paymentCycleId}"
                >{row.balanceType}</td
              >
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
