<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "./ts/api.ts";
  import { appState } from "ts/state.svelte.ts";
  import type { Contribution, ContributionSummary } from "./types/backend";
  import { formatDay, formatBalance } from "./ts/services.svelte.ts";
  /*import {
    Chart,
    LineController,
    LineElement,
    PointElement,
    BarController,
    BarElement,
    CategoryScale,
    LinearScale,
    Legend,
    Tooltip,
  } from "chart.js";
  import { Line, Bar } from "svelte-chartjs";*/
  //import { htmlLegendPlugin } from "utils";
  import '@fortawesome/fontawesome-free/css/all.min.css'
  import Main from "./Main.svelte";

  let contributionSummaries: ContributionSummary[] = [];
  let contributions: Contribution[] = [];
  let showGraph: string | undefined;
  let offset = 0;

  /*Chart.register(
    LineController,
    LineElement,
    PointElement,
    BarController,
    BarElement,
    CategoryScale,
    LinearScale,
    Legend,
    Tooltip
  );*/

  //https://www.chartjs.org/docs/latest/configuration/tooltip.html
  /*let dataOptions = {
    scales: {
      y: {
        ticks: {
          callback: function (value: number) {
            return (value * 100).toFixed(2) + "%";
          },
        },
      },
    },
    plugins: {
      tooltip: {
        boxPadding: 6,
        callbacks: {
          title: function (context) {
            let label = context[0].label || "";
            return "Git Metrics (3 Months) Until: " + label;
          },
          afterTitle: function (context) {
            let label = context[0].dataset.label || "";
            let start = label.indexOf(";");
            if (label && start > 0) {
              label = label.substring(start + 1);
              return JSON.parse(label).join(", ");
            }
            return label;
          },
          label: function (context) {
            let label = context.dataset.label || "";

            let start = label.indexOf(";");
            if (label && start > 0) {
              let start = label.indexOf(";");
              label = label.substring(0, start);
              label += ": ";
            }
            if (context.parsed.y !== null) {
              label += (context.parsed.y * 100).toFixed(2);
            } else {
              label += "0";
            }
            return label + "%";
          },
        },
      },
      htmlLegend: {
        // ID of the container to put the legend in
        containerID: "legend-container",
      },
      legend: {
        display: false,
      },
    },
  };*/

  onMount(async () => {
    try {
      const pr2 = API.user.contributionsSend();
      const pr3 = API.user.contributionsSummary();

      const res2 = await pr2;
      contributions = res2 || contributions;

      const res3 = await pr3;
      contributionSummaries = res3 || contributionSummaries;
    } catch (e) {
      appState.setError(e);
    }
  });
</script>

<style>
  @media screen and (max-width: 600px) {
    table {
      width: 100%;
    }
    table thead {
      border: none;
      clip: rect(0 0 0 0);
      height: 1px;
      margin: -1px;
      overflow: hidden;
      padding: 0;
      position: absolute;
      width: 1px;
    }

    table tr {
      border-bottom: 3px solid #fff;
      display: block;
    }

    table td {
      border-bottom: 1px solid #fff;
      display: block;
      font-size: 0.8em;
      text-align: right;
    }

    table td::before {
      content: attr(data-label);
      float: left;
      font-weight: bold;
      text-transform: uppercase;
    }

    table td:last-child {
      border-bottom: 0;
    }
  }
</style>

<Main>
  {#if contributionSummaries && contributionSummaries.length > 0}
    <h2 class="p-2 m-2">Supported Repositories</h2>
    <div class="container">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Description</th>
            <th>Unclaimed Sponsoring</th>
            <th>Graph</th>
          </tr>
        </thead>
        <tbody>
          {#each contributionSummaries as cs}
            <tr>
              <td data-label="Name"><a href={cs.repo.url}>{cs.repo.name}</a></td
              >
              <td
                data-label="Description"
                class={cs.repo.description ? "" : "no-desc"}
                >{cs.repo.description}</td
              >
              <td data-label="Unclaimed">
                {#each Object.entries(cs.currencyBalance) as [key, value]}{formatBalance(value, key)}{/each}
              </td>
              <td data-label="Graph">
                <div>
                  <button class="accessible-btn" aria-label={showGraph === cs.repo.uuid ? "Hide graph" : "Show graph"}
                          onclick={() => showGraph === cs.repo.uuid ? (showGraph = undefined) : (showGraph = cs.repo.uuid)}>
                    <i class="fas fa-plus" aria-hidden="true"></i>
                  </button>
                </div>
              </td>
            </tr>
            {#if showGraph === cs.repo.uuid}
              <tr id="bg-green1">
                <td colspan="6">
                  <div id="legend-container" ></div>
                  {#await API.repos.graph(cs.repo.uuid, offset)}
                    ...waiting
                  {:then data}
                    {#if data.days > 1}
                      <!--<Line
                        {data}
                        options={dataOptions}
                        plugins={[htmlLegendPlugin]}
                      />
                    {:else}
                      <Bar
                        {data}
                        options={dataOptions}
                        plugins={[htmlLegendPlugin]}
                      />-->
                    {/if}
                    {#if offset > 0}
                      <button
                        class="accessible-btn"
                        onclick={() => (offset -= 20)}
                      >
                        Previous 20 <i class="fa-arrow-left"></i>
                      </button>
                    {/if}
                    {#if data.total > offset + 20}
                      <button
                        class="accessible-btn"
                        onclick={() => (offset += 20)}
                      >
                        <i class="fa-arrow-right"></i> Next 20
                      </button>
                    {/if}
                  {:catch err}
                    {(appState.error = err.message)}
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
            <th>Realized</th>
            <th>Balance USD</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          {#each contributions as contribution}
            <tr>
              <td data-label="Respository">{contribution.repoName}</td>
              {#if contribution.contributorEmail}
                <td data-label="Email">{contribution.contributorEmail}</td>
                <td data-label="Realized">
                  {#if contribution.claimedAt}
                    Realized
                  {:else}
                    Unclaimed
                  {/if}
                </td>
                <td data-label="Balance"
                  >{formatBalance(
                    contribution.balance,
                    contribution.currency
                  )}</td
                >
              {:else}
                <td colspan="3"
                  >Unprocessed: {formatBalance(
                    contribution.balance,
                    contribution.currency
                  )} (analysis pending)</td
                >
              {/if}
              <td data-label="Date">{formatDay(new Date(contribution.day))}</td>
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
  <p class="p-2 m-2">
    Public URL: <a href="/badges/{appState.user.id}">/badges/{appState.user.id}</a>
  </p>
</Main>
