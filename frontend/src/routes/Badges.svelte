<script type="ts">
  import Navigation from "../components/Navigation.svelte";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { error, firstTime} from "../ts/store";
  import type { Contributions } from "../types/users";
  import { formatDay, formatMUSD } from "../ts/services";
  import confetti from "canvas-confetti";



  let contributions: Contributions[] = [];
  let canvas;

  onMount(async () => {
    try {
      let pr1;
      if ($firstTime) {
        pr1 = confetti.create(<HTMLCanvasElement>canvas, {
          resize: true,
          useWorker: true,
        })({ particleCount: 200, spread: 500 });
      }
      firstTime.set(false);
      const pr2 = API.user.contributions();
      const res2 = await pr2;
      contributions = res2 ? res2 : contributions;
      if (pr1) {
        await pr1;
      }
    } catch (e) {
      $error = e;
    }
  });

</script>

<style>
  canvas {
      position: absolute;
      width: 80%;
      height: 70%;
  }
</style>

<Navigation>
  <h1 class="px-2">Badges</h1>
  {#if firstTime}
    <canvas bind:this="{canvas}"></canvas>
  {/if}
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
