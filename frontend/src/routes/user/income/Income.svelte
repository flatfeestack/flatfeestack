<script lang="ts">
  import { API } from "../ts/api";
  import { error } from "../ts/mainStore";
  import { formatBalance, formatDate, timeSince } from "../ts/services";
</script>

<h2 class="p-2 m-2">Income</h2>
<p class="p-2 m-2">
  If you are an open-source contributor, and someone sponsored the respective
  repository, you can claim it here.
</p>

<div class="container">
  <table>
    <thead>
      <tr>
        <th>Repository</th>
        <th>From</th>
        <th>Balance</th>
        <th>Currency</th>
        <th>Realized</th>
        <th>Date</th>
      </tr>
    </thead>
    <tbody>
      {#await API.user.contributionsRcv()}
        ...waiting
      {:then contributions}
        {#each contributions as contribution}
          <tr>
            <td><a href={contribution.repoUrl}>{contribution.repoName}</a></td>
            <td
              >{contribution.sponsorName
                ? contribution.sponsorName
                : "[no name]"}</td
            >
            <td>{formatBalance(contribution.balance, contribution.currency)}</td
            >
            <td>{contribution.currency}</td>
            <td>
              {#if contribution.contributorUserId}
                Realized
              {:else}
                Unclaimed
              {/if}
            </td>
            <td title={formatDate(new Date(contribution.day))}>
              {timeSince(new Date(contribution.day), new Date())} ago
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6">No Data</td>
          </tr>
        {/each}
      {:catch err}
        {($error = err.message)}
      {/await}
    </tbody>
  </table>
</div>
