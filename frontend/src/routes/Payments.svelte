<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error, userBalances } from "../ts/store";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import type {Repo} from "../types/users";
  import { connectWs, formatDate} from "../ts/services";
  import PaymentSelection from "../components/PaymentSelection.svelte";

  let sponsoredRepos: Repo[] = [];
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
    <label class="nobreak">Selected Projects:</label>
    <div class="container">
      <span class="bold">{sponsoredRepos.length} projects</span>
    </div>
  </div>

  {#if $userBalances && $userBalances.total}
    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Currency</th>
          <th>My Balance</th>
        </tr>
        </thead>
        <tbody>
        {#each $userBalances.total as row}
          <tr>
            <td>{row.currency}</td>
            <td>{row.balance}</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <PaymentSelection/>

  {#if $userBalances && $userBalances.userBalances}
    <h2 class="p-2 m-2">Payment History</h2>
    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Payment Cycle</th>
          <th>Balance</th>
          <th>Currency</th>
          <th>Type</th>
          <th>Day</th>
        </tr>
        </thead>
        <tbody>

        {#each $userBalances.userBalances as row}
          <tr>
            <td>{row.paymentCycleId}</td>
            <td>{row.balance}</td>
            <td>{row.currency}</td>
            <td>{row.balanceType}</td>
            <td>{formatDate(new Date(row.createdAt))}</td>
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
