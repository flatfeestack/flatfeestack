<script type="ts">
  import Navigation from "../components/Navigation.svelte";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { error } from "../ts/store";
  import { Contributions } from "../types/users";
  import { formatDay, formatMUSD } from "../ts/services";

  let contributions: Contributions[] = [];

  onMount(async () => {
    try {
      const res = await API.user.contributions();
      contributions = res ? res : contributions;
    } catch (e) {
      $error = e;
    }
  });

</script>

<style></style>
<Navigation>
  <h1 class="px-2">Badges</h1>

  {#if contributions && contributions.length > 0}

    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Repository</th>
          <th>Contributor Email</th>
          <th>Contribution</th>
          <th>Realized</th>
          <th>Balance USD</th>
          <th>Date</th>
        </tr>
        </thead>
        <tbody>
        {#each contributions as contribution}
          <tr>
            <td>{contribution.repoName}</td>
            {#if contribution.contributorEmail}
              <td>{contribution.contributorEmail}</td>
              <td>{contribution.contributorWeight * 100}%</td>
              <td>
                {#if contribution.contributorUserId}
                  Realized
                {:else}
                  Unclaimed
                {/if}
              </td>
              <td>{formatMUSD(contribution.balance)}</td>
            {:else}
              <td colspan="4">Unprocessed: {formatMUSD(contribution.balanceRepo)} (analysis pending)</td>
            {/if}
            <td>{formatDay(new Date(contribution.day))}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="3">No Data</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
  {/if}


</Navigation>
