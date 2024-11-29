<script lang="ts">
  import { API } from "../ts/api";
  import {
    error,
    loadedLatestThresholds,
    latestThresholds,
  } from "../ts/mainStore";
  import type {
    Repo,
    RepoMetrics,
    PartialHealthValues,
  } from "../types/backend";
  import {
    Chart,
    BarElement,
    CategoryScale,
    LinearScale,
    Tooltip,
    Legend,
    BarController,
    Title,
    Animation,
  } from "chart.js";
  import { onMount, onDestroy } from "svelte";

  export let repo: Repo;

  let repoMetrics: RepoMetrics;
  let partialHealthValues: PartialHealthValues;

  let externalMetricHV1: number;
  let externalMetricHV2: number;
  let internalMetricHV1: number;
  let internalMetricHV2: number;
  let internalMetricHV3: number;

  let healthValueChart: Chart | null = null;
  let repoMetricsChartExternal1: Chart | null = null;
  let repoMetricsChartExternal2: Chart | null = null;
  let repoMetricsChartInternal1: Chart | null = null;
  let repoMetricsChartInternal2: Chart | null = null;
  let repoMetricsChartInternal3: Chart | null = null;

  async function getLatestThresholds() {
    try {
      $latestThresholds = await API.repos.getLatestHealthValueThresholds();
      $loadedLatestThresholds = true;
    } catch (e) {
      $error = e;
    }
  }

  async function getRepoHealthMetrics() {
    try {
      repoMetrics = await API.repos.getRepoMetricsById(repo.uuid);
    } catch (e) {
      $error = e;
    }
  }

  async function getPartialHealthValues() {
    try {
      partialHealthValues = await API.repos.getPartialHealthValues(repo.uuid);
      externalMetricHV1 = partialHealthValues.contributorvalue;
      externalMetricHV2 = partialHealthValues.commitvalue;
      internalMetricHV1 = partialHealthValues.sponsorvalue;
      internalMetricHV2 = partialHealthValues.repomultipliervalue;
      internalMetricHV3 = partialHealthValues.repostarvalue;
    } catch (e) {
      $error = e;
    }
  }

  function initRepoMetricsChart() {
    const chartConfigs = [
      {
        chart: repoMetricsChartExternal1,
        id: "repo-metrics-chart-em1",
        label: "Contributor Count",
        data: [repoMetrics.contributorcount],
        backgroundColor: ["rgba(75, 192, 192, 0.6)"],
        borderColor: ["rgba(75, 192, 192, 1)"],
        thresholds: $latestThresholds.ThContributorCount,
      },
      {
        chart: repoMetricsChartExternal2,
        id: "repo-metrics-chart-em2",
        label: "Commit Count",
        data: [repoMetrics.commitcount],
        backgroundColor: ["rgba(153, 102, 255, 0.6)"],
        borderColor: ["rgba(153, 102, 255, 1)"],
        thresholds: $latestThresholds.ThCommitCount,
      },
      {
        chart: repoMetricsChartInternal1,
        id: "repo-metrics-chart-im1",
        label: "Sponsors with Donation",
        data: [repoMetrics.sponsorcount],
        backgroundColor: ["rgba(255, 159, 64, 0.6)"],
        borderColor: ["rgba(255, 159, 64, 1)"],
        thresholds: $latestThresholds.ThSponsorDonation,
      },
      {
        chart: repoMetricsChartInternal2,
        id: "repo-metrics-chart-im2",
        label: "Multiplier Sponsors",
        data: [repoMetrics.repomultipliercount],
        backgroundColor: ["rgba(255, 99, 132, 0.6)"],
        borderColor: ["rgba(255, 99, 132, 1)"],
        thresholds: $latestThresholds.ThRepoMultiplier,
      },
      {
        chart: repoMetricsChartInternal3,
        id: "repo-metrics-chart-im3",
        label: "User w/o Donation",
        data: [repoMetrics.repostarcount],
        backgroundColor: ["rgba(54, 162, 235, 0.6)"],
        borderColor: ["rgba(54, 162, 235, 1)"],
        thresholds: $latestThresholds.ThRepoStarCount,
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
            `External Metric 1`,
            `External Metric 2`,
            `Internal Metric 1`,
            `Internal Metric 2`,
            `Internal Metric 3`,
          ],
          datasets: [
            {
              data: [
                externalMetricHV1,
                externalMetricHV2,
                internalMetricHV1,
                internalMetricHV2,
                internalMetricHV3,
              ],
              backgroundColor: [
                "rgba(75, 192, 192, 0.6)",
                "rgba(153, 102, 255, 0.6)",
                "rgba(255, 159, 64, 0.6)",
                "rgba(255, 99, 132, 0.6)",
                "rgba(54, 162, 235, 0.6)",
                // externalMetric1 > 2 ? 'red' : 'green',
                // externalMetric2 > 2 ? 'red' : 'green',
                // internalMetric1 > 2 ? 'red' : 'green',
                // internalMetric2 > 2 ? 'red' : 'green',
                // internalMetric3 > 2 ? 'red' : 'green',
              ],
              borderColor: [
                "rgba(75, 192, 192, 1)",
                "rgba(153, 102, 255, 1)",
                "rgba(255, 159, 64, 1)",
                "rgba(255, 99, 132, 1)",
                "rgba(54, 162, 235, 1)",
              ],
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
          },
        },
      });
    }
  }

  onMount(async () => {
    if (!$loadedLatestThresholds) {
      await getLatestThresholds();
    }
    await getRepoHealthMetrics();
    await getPartialHealthValues();
    initHealthValueChart();
    initRepoMetricsChart();
  });

  onDestroy(() => {
    if (healthValueChart) healthValueChart.destroy();
    if (repoMetricsChartExternal1) repoMetricsChartExternal1.destroy();
    if (repoMetricsChartExternal2) repoMetricsChartExternal2.destroy();
    if (repoMetricsChartInternal1) repoMetricsChartInternal1.destroy();
    if (repoMetricsChartInternal2) repoMetricsChartInternal2.destroy();
    if (repoMetricsChartInternal3) repoMetricsChartInternal3.destroy();
  });
</script>

<style>
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
  table,
  td {
    border: 1px solid var(--primary-700);
  }
  table {
    border-collapse: collapse;
  }
</style>

<div class="container-col">
  <h2>Assessment of {repo.name}</h2>
  <div class="container-col2 m-2">
    <h3>
      Composition of Health Value: <strong>{repo.healthValue}</strong>
    </h3>

    <div id="health-value-chart-div" class="container">
      <canvas id="health-value-chart" />
    </div>

    <table class="m-4">
      <tr>
        <td>Corresponding Points for External Metric 1:</td>
        <td><strong>{externalMetricHV1}</strong></td>
      </tr>
      <tr>
        <td>Corresponding Points for External Metric 2:</td>
        <td><strong>{externalMetricHV2}</strong></td>
      </tr>
      <tr>
        <td>Corresponding Points for Internal Metric 1:</td>
        <td><strong>{internalMetricHV1}</strong></td>
      </tr>
      <tr>
        <td>Corresponding Points for Internal Metric 2:</td>
        <td><strong>{internalMetricHV2}</strong></td>
      </tr>
      <tr>
        <td>Corresponding Points for Internal Metric 3:</td>
        <td><strong>{internalMetricHV3}</strong></td>
      </tr>
    </table>
  </div>

  <div class="container-col2 m-2">
    <h3>Exact Values of Repo Metrics Analysis and Associated Thresholds</h3>

    <p class="mtrl-4">
      External Metric 1: Active contributors in the last three months
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-em1" />
    </div>

    <p class="mtrl-4">External Metric 2: Commits in the last three months</p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-em2" />
    </div>

    <p class="mtrl-4">
      Internal Metric 1: Total active sponsoring for the repo
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-im1" />
    </div>

    <p class="mtrl-4">
      Internal Metric 2: Total active multiplier sponsoring for Repo
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-im2" />
    </div>

    <p class="mtrl-4">
      Internal Metric 3: Total active sponsoring by acredited developers
    </p>
    <div class="repo-metrics-chart-div container">
      <canvas id="repo-metrics-chart-im3" />
    </div>
  </div>
</div>
