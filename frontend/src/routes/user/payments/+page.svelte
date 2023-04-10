<script lang="ts">
  import { error, userBalances } from "$lib/ts/mainStore";
  import { API } from "$lib/ts/api";
  import { onMount } from "svelte";
  import type { Repo } from "$lib/types/users";
  import { connectWs, formatDate, formatBalance } from "$lib/ts/services";
  import PaymentSelection from "$lib/components/PaymentSelection.svelte";

  let sponsoredRepos: Repo[] = [];

  onMount(async () => {
    try {
      const pr1 = connectWs();
      const pr2 = API.user.getSponsored();
      const res2 = await pr2;
      sponsoredRepos = res2 || [];
      await pr1;
    } catch (e) {
      $error = e as string;
    }
  });
</script>

<h2 class="p-2 m-2">Sponsor Summary</h2>

<div class="grid-2">
  <p class="nobreak">
    Selected Projects: <span class="bold m-4"
      >{sponsoredRepos.length} projects</span
    >
  </p>
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
