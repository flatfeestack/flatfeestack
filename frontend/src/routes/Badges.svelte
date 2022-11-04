<script type="ts">
  import Navigation from "../components/Navigation.svelte";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { error, user} from "../ts/store";
  import type { Contributions, Repos } from "../types/users";
  import { formatDay, formatBalance } from "../ts/services";
  import {Line, Bar} from "svelte-chartjs";
  import Fa from "svelte-fa";
  import { faPlus, faArrowRight, faArrowLeft } from "@fortawesome/free-solid-svg-icons";
  import {htmlLegendPlugin} from "../ts/utils";

  let repos: Repos[] = [];
  let contributions: Contributions[] = [];
  let canvas;
  let showGraph;
  let offset = 0;

 //https://www.chartjs.org/docs/latest/configuration/tooltip.html
  let dataOptions={plugins: {
    tooltip: {
      boxPadding: 6,
      callbacks: {
        title: function(context) {
          let label = context[0].label || '';
          return "Git Metrics (3 Months) Until: " + label;
        },
        afterTitle: function(context) {
          let label = context[0].dataset.label || '';
          let start = label.indexOf(";")
          if (label && start > 0) {
            label = label.substring(start+1)
            return JSON.parse(label).join(", ")
          }
          return label;
        },
        label: function(context) {
          let label = context.dataset.label || '';

          let start = label.indexOf(";")
          if (label && start > 0) {
            let start = label.indexOf(";")
            label = label.substring(0, start)
            label += ': ';
          }
          if (context.parsed.y !== null) {
            label += (context.parsed.y * 100).toFixed(2);
          } else {
            label += "0"
          }
          return label + "%";
        }
      }
    },
      htmlLegend: {
        // ID of the container to put the legend in
        containerID: 'legend-container',
      },
      legend: {
        display: false,
      }
    }}

  onMount(async () => {
    try {
      let pr1;
      const pr2 = API.user.contributionsSend();
      const pr3 = API.user.contributionsSummary();

      const res2 = await pr2;
      contributions = res2 ? res2 : contributions;

      const res3 = await pr3;
      repos = res3 ? res3 : repos;

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
  {#if repos && repos.length > 0}
    <h2 class="p-2 m-2">Supported Repositories</h2>
    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Name</th>
          <th>URL</th>
          <th>Repos</th>
          <th>Description</th>
          <th>Unclaimed Sponsoring</th>
          <th>Graph</th>
        </tr>
        </thead>
        <tbody>
        {#each repos as repo}
          <tr>
            <td>{repo.repos[0].name}</td>
            <td><a href="{repo.repos[0].url}">{repo.repos[0].url}</a></td>
            <td>
              {#each repo.repos as r2}
                <a href="{r2.gitUrl}">{r2.gitUrl}</a>
              {/each}
            </td>
            <td>{repo.repos[0].description}</td>
            <td>{#each Object.entries(repo.balances) as [key, value]}{formatBalance(value, key)}{/each}</td>
            <td>
              <div class="cursor-pointer" on:click="{() => showGraph === repo.uuid? showGraph = undefined : showGraph = repo.uuid}">
                <Fa icon="{faPlus}" size="md"/>
              </div>
            </td>
          </tr>
          {#if showGraph === repo.uuid}
            <tr id="bg-green1">
              <td colspan="6">
                <div id="legend-container"></div>
                {#await API.repos.graph(repo.uuid, offset)}
                  ...waiting
                {:then data}
                  {#if data.days > 1}
                    <Line data={data} options="{dataOptions}" plugins="{[htmlLegendPlugin]}"/>
                  {:else}
                    <Bar data={data} options="{dataOptions}" plugins="{[htmlLegendPlugin]}"/>
                  {/if}
                  {#if offset > 0}
                    <span class="cursor-pointer" on:click="{() => offset-=20}">
                      Previous 20 <Fa icon="{faArrowLeft}" size="md"/>
                    </span>
                  {/if}
                  {#if data.total > offset + 20}
                    <span class="cursor-pointer" on:click="{() => offset+=20}">
                      <Fa icon="{faArrowRight}" size="md"/> Next 20
                    </span>
                  {/if}
                {:catch err}
                  {$error = err.message}
                {/await}
              </td>
            </tr>
          {/if}
        {:else}
          <tr>
            <td colspan="5">No Data</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
  {:else}
    <p class="p-2 m-2">No supported repositories yet.</p>
  {/if}


  {#if contributions && contributions.length > 0}
    <h2 class="px-2">Actual Contribution</h2>
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
              <td>{formatBalance(contribution.balance, "TODO")}</td>
            {:else}
              <td colspan="4">Unprocessed: {formatBalance(contribution.balanceRepo, "TODO")} (analysis pending)</td>
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
  {:else}
    <p class="p-2 m-2">No contributions yet.</p>
  {/if}
  <p class="p-2 m-2">Public URL: <a href="/badges/{$user.id}">/badges/{$user.id}"</a></p>

</Navigation>
