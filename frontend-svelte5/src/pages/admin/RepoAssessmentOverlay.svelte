<script lang="ts">


  import {
    BarController,
    BarElement,
    CategoryScale,
    Chart,
    Legend,
    LinearScale,
    Title,
    Tooltip,
  } from "chart.js";
  import { onDestroy, onMount } from "svelte";
  import { fade } from "svelte/transition";
  import type {PartialHealthValues, Repo, RepoMetrics} from "../../types/backend.ts";
  import {appState} from "../../ts/state.svelte";
  import {API} from "../../ts/api.ts";

  export let repo: Repo;

  let repoMetrics: RepoMetrics;
  let partialHealthValues: PartialHealthValues;
  let lastAnalysed: string;

  let contributorValue: number;
  let commitValue: number;
  let sponsorValue: number;
  let multiplierSponsorValue: number;
  let starValue: number;
  let activeFFSUserValue: number;

  let contributorCount: number;
  let commitCount: number;
  let sponsorCount: number;
  let multiplierSponsorCount: number;
  let starCount: number;
  let activeFFSUserCount: number;

  let healthValueChart: Chart | null = null;
  let contributorCountChart: Chart | null = null;
  let commitCountChart: Chart | null = null;
  let sponsorCountChart: Chart | null = null;
  let multiplierSponsorCountChart: Chart | null = null;
  let starCountChart: Chart | null = null;
  let activeFFSUserChart: Chart | null = null;

  async function getLatestThresholds() {
    try {
      appState.latestThresholds = await API.repos.getLatestHealthValueThresholds();
      appState.loadedLatestThresholds = true;
    } catch (e) {
      appState.setError(e);
    }
  }

  function setRepoMetricsVariables() {
    contributorCount = repoMetrics.contributorcount;
    commitCount = repoMetrics.commitcount;
    sponsorCount = repoMetrics.sponsorcount;
    multiplierSponsorCount = repoMetrics.repomultipliercount;
    starCount = repoMetrics.repostarcount;
    activeFFSUserCount = repoMetrics.activeffsusercount;
    lastAnalysed = repoMetrics.createdat;
  }

  async function getRepoHealthMetrics() {
    try {
      repoMetrics = await API.repos.getRepoMetricsById(repo.uuid);
      setRepoMetricsVariables();
    } catch (e) {
      appState.setError(e);
    }
  }

  function setPartialHealthValues() {
    contributorValue = partialHealthValues.contributorvalue;
    commitValue = partialHealthValues.commitvalue;
    sponsorValue = partialHealthValues.sponsorvalue;
    multiplierSponsorValue = partialHealthValues.repomultipliervalue;
    starValue = partialHealthValues.repostarvalue;
    activeFFSUserValue = partialHealthValues.activeffsuservalue;
  }

  async function getPartialHealthValues() {
    try {
      partialHealthValues = await API.repos.getPartialHealthValues(repo.uuid);
      setPartialHealthValues();
    } catch (e) {
      appState.setError(e);
    }
  }

  function initRepoMetricsChart() {
    const chartConfigs = [
      {
        chart: contributorCountChart,
        id: "repo-metrics-chart-em1",
        label: "Contributor Count",
        data: [repoMetrics.contributorcount],
        backgroundColor: ["rgba(255, 159, 64, 0.6)"],
        borderColor: ["rgba(255, 159, 64, 1)"],
        thresholds: appState.latestThresholds.ThContributorCount,
      },
      {
        chart: commitCountChart,
        id: "repo-metrics-chart-em2",
        label: "Commit Count",
        data: [repoMetrics.commitcount],
        backgroundColor: ["rgba(255, 99, 132, 0.6)"],
        borderColor: ["rgba(255, 99, 132, 1)"],
        thresholds: appState.latestThresholds.ThCommitCount,
      },
      {
        chart: sponsorCountChart,
        id: "repo-metrics-chart-im1",
        label: "Sponsors with Donation",
        data: [repoMetrics.sponsorcount], // change back to repoMetrics.sponsorcount
        backgroundColor: ["rgba(75, 192, 192, 0.6)"],
        borderColor: ["rgba(75, 192, 192, 1)"],
        thresholds: appState.latestThresholds.ThSponsorDonation,
      },
      {
        chart: multiplierSponsorCountChart,
        id: "repo-metrics-chart-im2",
        label: "Multiplier Sponsors",
        data: [repoMetrics.repomultipliercount], // change back to repoMetrics.repomultipliercount
        backgroundColor: ["rgba(153, 102, 255, 0.6)"],
        borderColor: ["rgba(153, 102, 255, 1)"],
        thresholds: appState.latestThresholds.ThRepoMultiplier,
      },
      {
        chart: starCountChart,
        id: "repo-metrics-chart-im3",
        label: "User w/o Donation",
        data: [repoMetrics.repostarcount], // change back to repoMetrics.repostarcount
        backgroundColor: ["rgba(54, 162, 235, 0.6)"],
        borderColor: ["rgba(54, 162, 235, 1)"],
        thresholds: appState.latestThresholds.ThRepoStarCount,
      },
      {
        chart: activeFFSUserChart,
        id: "repo-metrics-chart-im4",
        label: "Active FFS User",
        data: [repoMetrics.activeffsusercount],
        backgroundColor: ["rgba(255, 102, 178, 0.6)"],
        borderColor: ["rgba(255, 102, 178, 1)"],
        thresholds: appState.latestThresholds.ThActiveFFSUserCount,
      },
    ];

    chartConfigs.forEach((config) => {
      const ctx = (
        document.getElementById(config.id) as HTMLCanvasElement
      )?.getContext("2d");

      if (ctx) {
        Chart.register(
          BarController,
          BarElement,
          CategoryScale,
          LinearScale,
          Tooltip,
          Legend,
          Title
        );
        config.chart = new Chart(ctx, {
          type: "bar",
          data: {
            labels: [" "],
            datasets: [
              {
                label: config.label,
                data: config.data,
                backgroundColor: config.backgroundColor,
                borderColor: config.borderColor,
                borderWidth: 1,
              },
            ],
          },
          options: {
            indexAxis: "y",
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
              legend: {
                display: false,
              },
              tooltip: {
                enabled: true,
              },
            },
            scales: {
              y: {
                beginAtZero: true,
              },
              x: {
                min: 0,
                max: config.thresholds.lower + config.thresholds.upper,
              },
            },
            animation: {
              onComplete: function () {
                const chartInstance = config.chart;
                const xScale = chartInstance.scales.x;
                const yScale = chartInstance.scales.y;
                const ctx = chartInstance.ctx;

                ctx.save();
                ctx.lineWidth = 2;
                ctx.setLineDash([5, 5]); // Dashed line

                const lowerThreshold = xScale.getPixelForValue(
                  config.thresholds.lower
                );
                ctx.strokeStyle = "red";
                ctx.beginPath();
                ctx.moveTo(lowerThreshold, yScale.top);
                ctx.lineTo(lowerThreshold, yScale.bottom);
                ctx.stroke();

                const upperThreshold = xScale.getPixelForValue(
                  config.thresholds.upper
                );
                ctx.strokeStyle = "green";
                ctx.beginPath();
                ctx.moveTo(upperThreshold, yScale.top);
                ctx.lineTo(upperThreshold, yScale.bottom);
                ctx.stroke();

                ctx.restore();
              },
            },
          },
        });
      }
    });
  }

  function initHealthValueChart() {
    const ctx = (
      document.getElementById("health-value-chart") as HTMLCanvasElement
    )?.getContext("2d");

    if (ctx) {
      Chart.register(
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        Tooltip,
        Legend,
        Title
      );
      healthValueChart = new Chart(ctx, {
        type: "bar",
        data: {
          labels: [
            `Contributor Value`,
            `Commit Value`,
            `Sponsor Value`,
            `Multiplier Sponsor Value`,
            `Star Value`,
            `Active FFS User Value`,
          ],
          datasets: [
            {
              data: [
                contributorValue,
                commitValue,
                sponsorValue, // change back to sponsorValue
                multiplierSponsorValue, // change back to multiplierSponsorValue
                starValue, // change back to starValue
                activeFFSUserValue,
              ],
              backgroundColor: [
                "rgba(255, 159, 64, 0.6)",
                "rgba(255, 99, 132, 0.6)",
                "rgba(75, 192, 192, 0.6)",
                "rgba(153, 102, 255, 0.6)",
                "rgba(54, 162, 235, 0.6)",
                "rgba(255, 102, 178, 0.6)",
              ],
              borderColor: [
                "rgba(255, 159, 64, 1)",
                "rgba(255, 99, 132, 1)",
                "rgba(75, 192, 192, 1)",
                "rgba(153, 102, 255, 1)",
                "rgba(54, 162, 235, 1)",
                "rgba(255, 102, 178, 1)",
              ],
              borderWidth: 2,
            },
          ],
        },
        options: {
          indexAxis: "y",
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              display: false,
            },
            tooltip: {
              enabled: true,
            },
          },
          scales: {
            y: {
              beginAtZero: true,
            },
            x: {
              min: 0,
              max: 2,
            },
          },
        },
      });
    }
  }

  onMount(async () => {
    if (!appState.loadedLatestThresholds) {
      await getLatestThresholds();
    }
    await getRepoHealthMetrics();
    await getPartialHealthValues();
    initHealthValueChart();
    initRepoMetricsChart();
    // asyncDataLoaded = true;
  });

  onDestroy(() => {
    if (healthValueChart) healthValueChart.destroy();
    if (contributorCountChart) contributorCountChart.destroy();
    if (commitCountChart) commitCountChart.destroy();
    if (sponsorCountChart) sponsorCountChart.destroy();
    if (multiplierSponsorCountChart) multiplierSponsorCountChart.destroy();
    if (starCountChart) starCountChart.destroy();
    if (activeFFSUserChart) activeFFSUserChart.destroy();
  });
</script>

<style>
  h2,
  h3 {
    margin-bottom: 0;
  }

  .container-col {
    padding: 0;
  }
  #health-value-chart-div {
    height: 40vh;
    width: auto;
  }
  .repo-metrics-chart-div {
    height: 10vh;
    width: auto;
  }
  #health-value-chart,
  #repo-metrics-chart-em1,
  #repo-metrics-chart-em2 {
    height: 100%;
    width: 100%;
  }
  thead tr {
    background-color: var(--primary-700);
  }
  tbody tr:nth-of-type(even) {
    background-color: var(--secondary-100);
  }
</style>

<div class="container-col" transition:fade={{ duration: 500 }}>
  <h2>Assessment of {repo.name}</h2>
  <div class="container-col2 m-2">
    <h3>
      Overview - Composition of Health Value: {repo.healthValue}
    </h3>
    <table class="m-4">
      <thead>
        <tr>
          <th>Exact Values of Repo Metrics Analysis</th>
          <th class="text-center">Partial Health Values</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td
            >Corresponding points for <strong>{contributorCount}</strong> contributors:</td
          >
          <td class="text-center"><strong>{contributorValue}</strong></td>
        </tr>
        <tr>
          <td
            >Corresponding points for <strong>{commitCount}</strong> commits:</td
          >
          <td class="text-center"><strong>{commitValue}</strong></td>
        </tr>
        <tr>
          <td
            >Corresponding points for <strong>{sponsorCount}</strong> active sponsors:</td
          >
          <td class="text-center"><strong>{sponsorValue}</strong></td>
        </tr>
        <tr>
          <td
            >Corresponding points for <strong>{multiplierSponsorCount}</strong> active
            multiplier sponsors:</td
          >
          <td class="text-center"><strong>{multiplierSponsorValue}</strong></td>
        </tr>
        <tr>
          <td>Corresponding points for <strong>{starCount}</strong> stars:</td>
          <td class="text-center"><strong>{starValue}</strong></td>
        </tr>
        <tr>
          <td
            >Corresponding points for <strong>{activeFFSUserCount}</strong> active
            FFS user in this repo:</td
          >
          <td class="text-center"><strong>{activeFFSUserValue}</strong></td>
        </tr>
        <tr>
          <td><strong>Total</strong></td>
          <td class="text-center"><strong>{repo.healthValue}</strong></td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="container-col2 m-2">
    <h3>Partial Health Values</h3>

    <div id="health-value-chart-div" class="container">
      <canvas id="health-value-chart" />
    </div>
  </div>

  <div class="container-col2 m-2">
    <h3>Exact Values of Repo Metrics Analysis and Associated Thresholds</h3>

    <p class="mtrl-4">
      <strong>Contributor Count:</strong> Active contributors in the last three months
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-em1" />
    </div>

    <p class="mtrl-4">
      <strong>Commit Count:</strong> Commits in the last three months
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-em2" />
    </div>

    <p class="mtrl-4">
      <strong>Sponsoring Count:</strong> total amount of currently active sponsoring
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-im1" />
    </div>

    <p class="mtrl-4">
      <strong>Multiplier Sponsoring Count:</strong> total amount of currently active
      multiplier sponsoring
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-im2" />
    </div>

    <p class="mtrl-4">
      <strong>Star Count:</strong> total recieved stars by users <br /> (only users
      that donated on FlatFeeStack in the last 3 months are counted)
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-im3" />
    </div>

    <p class="mtrl-4">
      <strong>Active FFS User Count:</strong> total active FlatFeeStack user contributing
      to this repository
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-im4" />
    </div>
  </div>

  <div class="container-col2 m-2">
    <h3>Last Analysis of Repository</h3>
    <div class="container justify-between">
      <p>Date: {lastAnalysed}</p>
    </div>
  </div>
</div>
