<script lang="ts">
  import { API } from "../ts/api";
  import {
    error,
    loadedLatestThresholds,
    latestThresholds,
  } from "../ts/mainStore";
  import type { Repo, RepoMetrics } from "../types/backend";
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

  const externalMetric1Points = 1.1;
  const externalMetric2Points = 1.2;
  const internalMetric1Points = 2.1;
  const internalMetric2Points = 2.2;
  const internalMetric3Points = 2.3;

  let repoMetrics: RepoMetrics;

  let externalMetric1: number;
  let externalMetric2: number;
  let internalMetric1: number;
  let internalMetric2: number;
  let internalMetric3: number;

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
            // `${externalMetric1} External Metric 1`,
            // `${externalMetric2} External Metric 2`,
            // `${internalMetric1} Internal Metric 1`,
            // `${internalMetric2} Internal Metric 2`,
            // `${internalMetric3} Internal Metric 3`,
            `External Metric 1`,
            `External Metric 2`,
            `Internal Metric 1`,
            `Internal Metric 2`,
            `Internal Metric 3`,
          ],
          datasets: [
            {
              data: [
                externalMetric1Points,
                externalMetric2Points,
                internalMetric1Points,
                internalMetric2Points,
                internalMetric3Points,
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
    width: 52vw;
  }
  .repo-metrics-chart-div {
    height: 10vh;
    width: 52vw;
  }
  #health-value-chart,
  #repo-metrics-chart-em1,
  #repo-metrics-chart-em2 {
    height: 100%;
    width: 100%;
  }
</style>

<div class="container-col">
  <h2>Assessment of {repo.name}</h2>
  <div class="container-col2 m-2">
    <h3>
      Total Health Value: <strong>{repo.healthValue}</strong>
    </h3>
    <p>
      External Metric 1: <strong>{externalMetric1}</strong>
      <br />
      External Metric 2: <strong>{externalMetric2}</strong>
    </p>
    <p>
      Internal Metric 1: <strong>{internalMetric1}</strong>
      <br />
      Internal Metric 2: <strong>{internalMetric2}</strong>
      <br />
      Internal Metric 3: <strong>{internalMetric3}</strong>
    </p>
  </div>

  <div id="health-value-chart-div" class="container m-2">
    <canvas id="health-value-chart" />
  </div>

  <p class="mt-4">Active contributors in the last three months</p>
  <div class="repo-metrics-chart-div m-2 container">
    <canvas id="repo-metrics-chart-em1" />
  </div>

  <p class="mt-4">Commits in the last three months</p>
  <div class="repo-metrics-chart-div container m-2">
    <canvas id="repo-metrics-chart-em2" />
  </div>

  <p class="mt-4">Total active sponsoring for the repo</p>
  <div class="repo-metrics-chart-div container m-2">
    <canvas id="repo-metrics-chart-im1" />
  </div>

  <p class="mt-4">Total active multiplier sponsoring for Repo</p>
  <div class="repo-metrics-chart-div container m-2">
    <canvas id="repo-metrics-chart-im2" />
  </div>

  <p class="mt-4">Total active sponsoring by acredited developers</p>
  <div class="repo-metrics-chart-div container m-2">
    <canvas id="repo-metrics-chart-im3" />
  </div>
</div>
