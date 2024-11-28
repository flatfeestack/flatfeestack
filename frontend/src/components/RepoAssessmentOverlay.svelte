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
      console.log("thresholds: ", $latestThresholds);
      console.log("contributors: ", $latestThresholds.ThContributorCount);
      console.log("commit: ", $latestThresholds.ThCommitCount);
      console.log("sponsordonation: ", $latestThresholds.ThSponsorDonation);
      console.log("multiplier: ", $latestThresholds.ThRepoMultiplier);
      console.log("starCount: ", $latestThresholds.ThRepoStarCount);
    } catch (e) {
      $error = e;
    }
  }

  async function getRepoHealthMetrics() {
    try {
      repoMetrics = await API.repos.getRepoMetricsById(repo.uuid);
      console.log("repo metrics: ", repoMetrics);
      console.log("repo metric contrib:", repoMetrics.contributorcount);
      console.log("repo metric commit:", repoMetrics.commitcount);
    } catch (e) {
      $error = e;
    }
  }

  function initRepoMetricsChart() {
    console.log("repo metric commit:", repoMetrics.commitcount);
    const ctxEM1 = (
      document.getElementById("repo-metrics-chart-em1") as HTMLCanvasElement
    )?.getContext("2d");
    const ctxEM2 = (
      document.getElementById("repo-metrics-chart-em2") as HTMLCanvasElement
    )?.getContext("2d");
    const ctxIM1 = (
      document.getElementById("repo-metrics-chart-im1") as HTMLCanvasElement
    )?.getContext("2d");
    const ctxIM2 = (
      document.getElementById("repo-metrics-chart-im2") as HTMLCanvasElement
    )?.getContext("2d");
    const ctxIM3 = (
      document.getElementById("repo-metrics-chart-im3") as HTMLCanvasElement
    )?.getContext("2d");

    if (ctxEM1) {
      Chart.register(
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        Tooltip,
        Legend,
        Title
      );
      repoMetricsChartExternal1 = new Chart(ctxEM1, {
        type: "bar",
        data: {
          labels: [" "],
          datasets: [
            {
              label: "Contributor Count",
              data: [repoMetrics.contributorcount],
              backgroundColor: ["rgba(75, 192, 192, 0.6)"],
              borderColor: ["rgba(75, 192, 192, 1)"],
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
              max:
                $latestThresholds.ThContributorCount.lower +
                $latestThresholds.ThContributorCount.upper,
            },
          },
          animation: {
            onComplete: function () {
              const chartInstance = repoMetricsChartExternal1;
              const xScale = chartInstance.scales.x;
              const yScale = chartInstance.scales.y;
              const ctx = chartInstance.ctx;

              // Save context state
              ctx.save();
              ctx.lineWidth = 2;
              ctx.setLineDash([5, 5]); // Dashed line

              // Draw the red line at lower Threshold
              const lowerThreshold = xScale.getPixelForValue(
                $latestThresholds.ThContributorCount.lower
              );
              ctx.strokeStyle = "red"; // Set color to red
              ctx.beginPath();
              ctx.moveTo(lowerThreshold, yScale.top); // Top of the y-axis
              ctx.lineTo(lowerThreshold, yScale.bottom); // Bottom of the y-axis
              ctx.stroke();

              // Draw the green line at upper Threshold
              const upperThreshold = xScale.getPixelForValue(
                $latestThresholds.ThContributorCount.upper
              );
              ctx.strokeStyle = "green"; // Set color to green
              ctx.beginPath();
              ctx.moveTo(upperThreshold, yScale.top); // Top of the y-axis
              ctx.lineTo(upperThreshold, yScale.bottom); // Bottom of the y-axis
              ctx.stroke();

              // Restore context state
              ctx.restore();
            },
          },
        },
        // plugins: [thresholdPlugin],
      });
    }
    if (ctxEM2) {
      Chart.register(
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        Tooltip,
        Legend,
        Title
      );
      repoMetricsChartExternal2 = new Chart(ctxEM2, {
        type: "bar",
        data: {
          labels: [" "],
          datasets: [
            {
              label: "Commit Count",
              data: [repoMetrics.commitcount],
              backgroundColor: ["rgba(153, 102, 255, 0.6)"],
              borderColor: ["rgba(153, 102, 255, 1)"],
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
              max:
                $latestThresholds.ThCommitCount.lower +
                $latestThresholds.ThCommitCount.upper,
            },
          },
          animation: {
            onComplete: function () {
              const chartInstance = repoMetricsChartExternal2;
              const xScale = chartInstance.scales.x;
              const yScale = chartInstance.scales.y;
              const ctx = chartInstance.ctx;

              // Save context state
              ctx.save();
              ctx.lineWidth = 2;
              ctx.setLineDash([5, 5]); // Dashed line

              // Draw the red line at lower Threshold
              const lowerThreshold = xScale.getPixelForValue(
                $latestThresholds.ThCommitCount.lower
              );
              ctx.strokeStyle = "red"; // Set color to red
              ctx.beginPath();
              ctx.moveTo(lowerThreshold, yScale.top); // Top of the y-axis
              ctx.lineTo(lowerThreshold, yScale.bottom); // Bottom of the y-axis
              ctx.stroke();

              // Draw the green line at upper Threshold
              const upperThreshold = xScale.getPixelForValue(
                $latestThresholds.ThCommitCount.upper
              );
              ctx.strokeStyle = "green"; // Set color to green
              ctx.beginPath();
              ctx.moveTo(upperThreshold, yScale.top); // Top of the y-axis
              ctx.lineTo(upperThreshold, yScale.bottom); // Bottom of the y-axis
              ctx.stroke();

              // Restore context state
              ctx.restore();
            },
          },
        },
        // plugins: [thresholdPlugin],
      });
    }
    if (ctxIM1) {
      Chart.register(
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        Tooltip,
        Legend,
        Title
      );
      repoMetricsChartInternal1 = new Chart(ctxIM1, {
        type: "bar",
        data: {
          labels: [" "],
          datasets: [
            {
              label: "Commit Count",
              data: [repoMetrics.repostarcount],
              backgroundColor: ["rgba(255, 159, 64, 0.6)"],
              borderColor: ["rgba(255, 159, 64, 1)"],
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
            // x: {
            //   min: thInternalMetric1.lower,
            //   max: thInternalMetric1.upper,
            // },
          },
        },
        // plugins: [thresholdPlugin],
      });
    }
    if (ctxIM2) {
      Chart.register(
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        Tooltip,
        Legend,
        Title
      );
      repoMetricsChartInternal2 = new Chart(ctxIM2, {
        type: "bar",
        data: {
          labels: [" "],
          datasets: [
            {
              label: "Commit Count",
              data: [repoMetrics.repomultipliercount],
              backgroundColor: ["rgba(255, 99, 132, 0.6)"],
              borderColor: ["rgba(255, 99, 132, 1)"],
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
            // x: {
            //   min: thInternalMetric1.lower,
            //   max: thInternalMetric1.upper,
            // },
          },
        },
        // plugins: [thresholdPlugin],
      });
    }
    if (ctxIM3) {
      Chart.register(
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        Tooltip,
        Legend,
        Title
      );
      repoMetricsChartInternal3 = new Chart(ctxIM3, {
        type: "bar",
        data: {
          labels: [" "],
          datasets: [
            {
              label: "Commit Count",
              data: [repoMetrics.sponsorcount],
              backgroundColor: ["rgba(54, 162, 235, 0.6)"],
              borderColor: ["rgba(54, 162, 235, 1)"],
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
            // x: {
            //   min: thInternalMetric1.lower,
            //   max: thInternalMetric1.upper,
            // },
          },
        },
        // plugins: [thresholdPlugin],
      });
    }
  }

  function initHealthValueChart() {
    const ctx = (
      document.getElementById("health-value-chart") as HTMLCanvasElement
    )?.getContext("2d");

    if (ctx) {
      // const thresholds = [1.1, 1.1, 2.2, 2.2, 2.2];

      // const thresholdPlugin = {
      //   id: 'thresholdPlugin',
      //   afterDatasetsDraw(chart: any) {
      //     const { ctx, scales } = chart;
      //     const xScale = scales.x;
      //     const yScale = scales.y;
      //
      //     ctx.save();
      //     thresholds.forEach((threshold, index) => {
      //       const xCenter = xScale.getPixelForValue(index); // Center of the bar
      //       const yThreshold = yScale.getPixelForValue(threshold); // Y position for threshold
      //
      //       ctx.beginPath();
      //       ctx.strokeStyle = threshold > 2 ? 'red' : 'green'; // Red for > 2, green otherwise
      //       ctx.lineWidth = 2;
      //       ctx.moveTo(xCenter - xScale.getPixelForTickWidth() / 2, yThreshold);
      //       ctx.lineTo(xCenter + xScale.getPixelForTickWidth() / 2, yThreshold);
      //       ctx.stroke();
      //     });
      //     ctx.restore();
      //   },
      // }

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
        // plugins: [thresholdPlugin],
      });
      // repoMetricsChart = new Chart(ctx, {
      //   type: 'bar',
      //   data: {
      //     labels: [
      //       `External Metric 1`,
      //     ],
      //     datasets: [
      //       {
      //         data: [
      //           externalMetric1,
      //         ],
      //         backgroundColor: [
      //           'rgba(75, 192, 192, 0.6)',
      //         ],
      //         borderColor: [
      //           'rgba(75, 192, 192, 1)',
      //         ],
      //         borderWidth: 1,
      //       },
      //     ],
      //   },
      //   options: {
      //     indexAxis: 'y',
      //     responsive: true,
      //     plugins: {
      //       legend: {
      //         display: false,
      //       },
      //       tooltip: {
      //         enabled: true,
      //       },
      //     },
      //     scales: {
      //       y: {
      //         beginAtZero: true,
      //       },
      //     },
      //   },
      //   // plugins: [thresholdPlugin],
      // });
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
