<script lang="ts">
  import { fade } from "svelte/transition";
  import {
    error,
    latestThresholds,
    loadedLatestThresholds,
  } from "../ts/mainStore";
  import { onMount } from "svelte";
  import { API } from "../ts/api";

  let newThContributorCountLower: number;
  let newThContributorCountUpper: number;
  let newThCommitCountLower: number;
  let newThCommitCountUpper: number;
  let newThSponsorDonationLower: number;
  let newThSponsorDonationUpper: number;
  let newThRepoMultiplierLower: number;
  let newThRepoMultiplierUpper: number;
  let newThRepoStarCountLower: number;
  let newThRepoStarCountUpper: number;
  let newThActiveFFSUserCountLower: number;
  let newThActiveFFSUserCountUpper: number;

  async function getLatestThresholds() {
    try {
      $latestThresholds = await API.repos.getLatestHealthValueThresholds();
      $loadedLatestThresholds = true;
    } catch (e) {
      $error = e;
    }
  }

  function createHealthValueThreshold({
    ThContributorCount,
    ThCommitCount,
    ThSponsorDonation,
    ThRepoStarCount,
    ThRepoMultiplier,
    ThActiveFFSUserCount,
  }) {
    return {
      id: 0,
      createdAt: 0,
      ThContributorCount: { ...ThContributorCount },
      ThCommitCount: { ...ThCommitCount },
      ThSponsorDonation: { ...ThSponsorDonation },
      ThRepoStarCount: { ...ThRepoStarCount },
      ThRepoMultiplier: { ...ThRepoMultiplier },
      ThActiveFFSUserCount: { ...ThActiveFFSUserCount },
    };
  }

  function setNewThresholds(threshold: string): void {
    console.log("latest thresholds:", $latestThresholds);

    const newHealthValueThreshold = createHealthValueThreshold({
      ThContributorCount:
        threshold === "ThContributorCount" || threshold === "allThresholds"
          ? {
              upper: newThContributorCountUpper,
              lower: newThContributorCountLower,
            }
          : {
              upper: $latestThresholds.ThContributorCount.upper,
              lower: $latestThresholds.ThContributorCount.lower,
            },
      ThCommitCount:
        threshold === "ThCommitCount" || threshold === "allThresholds"
          ? { upper: newThCommitCountUpper, lower: newThCommitCountLower }
          : {
              upper: $latestThresholds.ThCommitCount.upper,
              lower: $latestThresholds.ThCommitCount.lower,
            },
      ThSponsorDonation:
        threshold === "ThSponsorDonation" || threshold === "allThresholds"
          ? {
              upper: newThSponsorDonationUpper,
              lower: newThSponsorDonationLower,
            }
          : {
              upper: $latestThresholds.ThSponsorDonation.upper,
              lower: $latestThresholds.ThSponsorDonation.lower,
            },
      ThRepoStarCount:
        threshold === "ThRepoStarCount" || threshold === "allThresholds"
          ? { upper: newThRepoStarCountUpper, lower: newThRepoStarCountLower }
          : {
              upper: $latestThresholds.ThRepoStarCount.upper,
              lower: $latestThresholds.ThRepoStarCount.lower,
            },
      ThRepoMultiplier:
        threshold === "ThRepoMultiplier" || threshold === "allThresholds"
          ? { upper: newThRepoMultiplierUpper, lower: newThRepoMultiplierLower }
          : {
              upper: $latestThresholds.ThRepoMultiplier.upper,
              lower: $latestThresholds.ThRepoMultiplier.lower,
            },
      ThActiveFFSUserCount:
        threshold === "ThActiveFFSUserCount" || threshold === "allThresholds"
          ? {
              upper: newThActiveFFSUserCountUpper,
              lower: newThActiveFFSUserCountLower,
            }
          : {
              upper: $latestThresholds.ThActiveFFSUserCount.upper,
              lower: $latestThresholds.ThActiveFFSUserCount.lower,
            },
    });

    console.log("new thresholds:", newHealthValueThreshold);

    // try {
    //   if (newDailyLimit >= 1) {
    //     API.user.setMultiplierDailyLimit(newDailyLimit);
    //     dailyLimit = newDailyLimit;
    //     $user.multiplierDailyLimit = newDailyLimit;
    //     newDailyLimit = "";
    //   } else {
    //     $error = "The daily limit must be a number greater than or equalt to 1";
    //   }
    // } catch (e) {
    //   $error = e;
    // }
  }

  onMount(async () => {
    if (!$loadedLatestThresholds) {
      await getLatestThresholds();
    }
  });
</script>

<style>
</style>

{#if $loadedLatestThresholds}
  <div class="container-col" transition:fade={{ duration: 500 }}>
    <h2>Threshold Settings</h2>
    <div class="container-col2 m-2">
      <p class="m-0">
        This settings allows you to customize the thresholds of the metrics used
        to assess repository health. For each metric, you can define lower and
        upper thresholds that determine how points are allocated:
      </p>
      <ul class="my-2">
        <li>
          Lower Threshold: The minimum value a repository must meet to earn
          points for this metric.
        </li>
        <li>
          Upper Threshold: The value at which a repository earns the maximum
          points for this metric.
        </li>
      </ul>
      <p class="m-0">
        These adjustments directly impact how repository health values are
        calculated.
      </p>
    </div>

    <div class="container-col2 m-2">
      <h3 class="mt-4">Thresholds for Contributor Count</h3>

      <div class="container justify-between">
        <div class="container-small">
          <div class="container-col2 mr-10">
            <label for="contributor-count-input-lower">lower:</label>
            <input
              id="contributor-count-input-lower"
              type="number"
              class="max-w20"
              bind:value={newThContributorCountLower}
              placeholder={String($latestThresholds.ThContributorCount.lower)}
            />
          </div>

          <div class="container-col2 mr-10">
            <label for="contributor-count-input-upper">upper:</label>
            <input
              id="contributor-count-input-upper"
              type="number"
              class="max-w20"
              bind:value={newThContributorCountUpper}
              placeholder={String($latestThresholds.ThContributorCount.upper)}
            />
          </div>
        </div>

        <button
          on:click={() => setNewThresholds("ThContributorCount")}
          class="ml-5 p-2 button1">Set Thresholds</button
        >
      </div>
    </div>

    <div class="container-col2 m-2">
      <h3 class="mt-4">Thresholds for Commit Count</h3>

      <div class="container justify-between">
        <div class="container-small">
          <div class="container-col2 mr-10">
            <label for="commit-count-input-lower">lower:</label>
            <input
              id="commit-count-input-lower"
              type="number"
              class="max-w20"
              bind:value={newThCommitCountLower}
              placeholder={String($latestThresholds.ThCommitCount.lower)}
            />
          </div>

          <div class="container-col2 mr-10">
            <label for="commit-count-input-upper">upper:</label>
            <input
              id="commit-count-input-upper"
              type="number"
              class="max-w20"
              bind:value={newThCommitCountUpper}
              placeholder={String($latestThresholds.ThCommitCount.upper)}
            />
          </div>
        </div>

        <button
          on:click={() => setNewThresholds("ThCommitCount")}
          class="ml-5 p-2 button1">Set Thresholds</button
        >
      </div>
    </div>

    <div class="container-col2 m-2">
      <h3 class="mt-4">Thresholds for Sponsoring Count</h3>

      <div class="container justify-between">
        <div class="container-small">
          <div class="container-col2 mr-10">
            <label for="sponsor-count-input-lower">lower:</label>
            <input
              id="sponsor-count-input-lower"
              type="number"
              class="max-w20"
              bind:value={newThSponsorDonationLower}
              placeholder={String($latestThresholds.ThSponsorDonation.lower)}
            />
          </div>

          <div class="container-col2 mr-10">
            <label for="sponsor-count-input-upper">upper:</label>
            <input
              id="sponsor-count-input-upper"
              type="number"
              class="max-w20"
              bind:value={newThSponsorDonationUpper}
              placeholder={String($latestThresholds.ThSponsorDonation.upper)}
            />
          </div>
        </div>

        <button
          on:click={() => setNewThresholds("ThSponsorDonation")}
          class="ml-5 p-2 button1">Set Thresholds</button
        >
      </div>
    </div>

    <div class="container-col2 m-2">
      <h3 class="mt-4">Thresholds for Multiplier Sponsoring Count</h3>

      <div class="container justify-between">
        <div class="container-small">
          <div class="container-col2 mr-10">
            <label for="multiplier-count-input-lower">lower:</label>
            <input
              id="multiplier-count-input-lower"
              type="number"
              class="max-w20"
              bind:value={newThRepoMultiplierLower}
              placeholder={String($latestThresholds.ThRepoMultiplier.lower)}
            />
          </div>

          <div class="container-col2 mr-10">
            <label for="multiplier-count-input-upper">upper:</label>
            <input
              id="multiplier-count-input-upper"
              type="number"
              class="max-w20"
              bind:value={newThRepoMultiplierUpper}
              placeholder={String($latestThresholds.ThRepoMultiplier.upper)}
            />
          </div>
        </div>

        <button
          on:click={() => setNewThresholds("ThRepoMultiplier")}
          class="ml-5 p-2 button1">Set Thresholds</button
        >
      </div>
    </div>

    <div class="container-col2 m-2">
      <h3 class="mt-4">Thresholds for Star Count</h3>

      <div class="container justify-between">
        <div class="container-small">
          <div class="container-col2 mr-10">
            <label for="star-count-input-lower">lower:</label>
            <input
              id="star-count-input-lower"
              type="number"
              class="max-w20"
              bind:value={newThRepoStarCountLower}
              placeholder={String($latestThresholds.ThRepoStarCount.lower)}
            />
          </div>

          <div class="container-col2 mr-10">
            <label for="star-count-input-upper">upper:</label>
            <input
              id="star-count-input-upper"
              type="number"
              class="max-w20"
              bind:value={newThRepoStarCountUpper}
              placeholder={String($latestThresholds.ThRepoStarCount.upper)}
            />
          </div>
        </div>

        <button
          on:click={() => setNewThresholds("ThRepoStarCount")}
          class="ml-5 p-2 button1">Set Thresholds</button
        >
      </div>
    </div>

    <div class="container-col2 m-2">
      <h3 class="mt-4">Thresholds for Active FFS User Count</h3>

      <div class="container justify-between">
        <div class="container-small">
          <div class="container-col2 mr-10">
            <label for="ffs-user-count-input-lower">lower:</label>
            <input
              id="ffs-user-count-input-lower"
              type="number"
              class="max-w20"
              bind:value={newThActiveFFSUserCountLower}
              placeholder={String($latestThresholds.ThActiveFFSUserCount.lower)}
            />
          </div>

          <div class="container-col2 mr-10">
            <label for="ffs-user-count-input-upper">upper:</label>
            <input
              id="ffs-user-count-input-upper"
              type="number"
              class="max-w20"
              bind:value={newThActiveFFSUserCountUpper}
              placeholder={String($latestThresholds.ThActiveFFSUserCount.upper)}
            />
          </div>
        </div>

        <button
          on:click={() => setNewThresholds("ThActiveFFSUserCount")}
          class="ml-5 p-2 button1">Set Thresholds</button
        >
      </div>
    </div>
  </div>
{:else}
  <p>failed to fetch thresholds</p>
{/if}
