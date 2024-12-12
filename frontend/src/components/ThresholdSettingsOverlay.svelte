<script lang="ts">
  import { fade } from "svelte/transition";
  import {
    error,
    latestThresholds,
    loadedLatestThresholds,
    reloadHealthRepoCardKey,
    trustedRepos,
  } from "../ts/mainStore";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import type { HealthValueThreshold } from "../types/backend";

  let reloadKey = 0;

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
      id: $latestThresholds.id,
      createdAt: $latestThresholds.createdAt,
      ThContributorCount: { ...ThContributorCount },
      ThCommitCount: { ...ThCommitCount },
      ThSponsorDonation: { ...ThSponsorDonation },
      ThRepoStarCount: { ...ThRepoStarCount },
      ThRepoMultiplier: { ...ThRepoMultiplier },
      ThActiveFFSUserCount: { ...ThActiveFFSUserCount },
    };
  }

  async function updateHealthValueForRepoCards() {
    try {
      $trustedRepos = await API.repos.getTrusted();
    } catch (e) {
      $error = e;
    }

    reloadHealthRepoCardKey.update((key) => key + 1);
    setTimeout(async () => {
      reloadHealthRepoCardKey.update(() => 0);
    }, 500);
  }

  function emptyThresholdInputFields() {
    newThContributorCountLower = undefined;
    newThContributorCountUpper = undefined;
    newThCommitCountLower = undefined;
    newThCommitCountUpper = undefined;
    newThSponsorDonationLower = undefined;
    newThSponsorDonationUpper = undefined;
    newThRepoMultiplierLower = undefined;
    newThRepoMultiplierUpper = undefined;
    newThRepoStarCountLower = undefined;
    newThRepoStarCountUpper = undefined;
    newThActiveFFSUserCountLower = undefined;
    newThActiveFFSUserCountUpper = undefined;
  }

  function handleInputValidation(event) {
    event.target.value = event.target.value.replace(/[^0-9]/g, ""); // Store last valid value
  }

  async function setNewThresholds(threshold: string) {
    const newHealthValueThreshold: HealthValueThreshold =
      createHealthValueThreshold({
        ThContributorCount: {
          upper:
            threshold === "ThContributorCount" || threshold === "allThresholds"
              ? newThContributorCountUpper ||
                $latestThresholds.ThContributorCount.upper
              : $latestThresholds.ThContributorCount.upper,
          lower:
            threshold === "ThContributorCount" || threshold === "allThresholds"
              ? newThContributorCountLower ||
                $latestThresholds.ThContributorCount.lower
              : $latestThresholds.ThContributorCount.lower,
        },
        ThCommitCount: {
          upper:
            threshold === "ThCommitCount" || threshold === "allThresholds"
              ? newThCommitCountUpper || $latestThresholds.ThCommitCount.upper
              : $latestThresholds.ThCommitCount.upper,
          lower:
            threshold === "ThCommitCount" || threshold === "allThresholds"
              ? newThCommitCountLower || $latestThresholds.ThCommitCount.lower
              : $latestThresholds.ThCommitCount.lower,
        },
        ThSponsorDonation: {
          upper:
            threshold === "ThSponsorDonation" || threshold === "allThresholds"
              ? newThSponsorDonationUpper ||
                $latestThresholds.ThSponsorDonation.upper
              : $latestThresholds.ThSponsorDonation.upper,
          lower:
            threshold === "ThSponsorDonation" || threshold === "allThresholds"
              ? newThSponsorDonationLower ||
                $latestThresholds.ThSponsorDonation.lower
              : $latestThresholds.ThSponsorDonation.lower,
        },
        ThRepoStarCount: {
          upper:
            threshold === "ThRepoStarCount" || threshold === "allThresholds"
              ? newThRepoStarCountUpper ||
                $latestThresholds.ThRepoStarCount.upper
              : $latestThresholds.ThRepoStarCount.upper,
          lower:
            threshold === "ThRepoStarCount" || threshold === "allThresholds"
              ? newThRepoStarCountLower ||
                $latestThresholds.ThRepoStarCount.lower
              : $latestThresholds.ThRepoStarCount.lower,
        },
        ThRepoMultiplier: {
          upper:
            threshold === "ThRepoMultiplier" || threshold === "allThresholds"
              ? newThRepoMultiplierUpper ||
                $latestThresholds.ThRepoMultiplier.upper
              : $latestThresholds.ThRepoMultiplier.upper,
          lower:
            threshold === "ThRepoMultiplier" || threshold === "allThresholds"
              ? newThRepoMultiplierLower ||
                $latestThresholds.ThRepoMultiplier.lower
              : $latestThresholds.ThRepoMultiplier.lower,
        },
        ThActiveFFSUserCount: {
          upper:
            threshold === "ThActiveFFSUserCount" ||
            threshold === "allThresholds"
              ? newThActiveFFSUserCountUpper ||
                $latestThresholds.ThActiveFFSUserCount.upper
              : $latestThresholds.ThActiveFFSUserCount.upper,
          lower:
            threshold === "ThActiveFFSUserCount" ||
            threshold === "allThresholds"
              ? newThActiveFFSUserCountLower ||
                $latestThresholds.ThActiveFFSUserCount.lower
              : $latestThresholds.ThActiveFFSUserCount.lower,
        },
      });

    try {
      await API.repos.setNewHealthValueThresholds(newHealthValueThreshold);
    } catch (e) {
      $error = e;
      return;
    }

    $loadedLatestThresholds = false;
    reloadKey = reloadKey + 1;
    console.log("reloadKey", reloadKey);
    await getLatestThresholds();

    emptyThresholdInputFields();

    await updateHealthValueForRepoCards();
  }

  onMount(async () => {
    if (!$loadedLatestThresholds) {
      await getLatestThresholds();
    }
  });
</script>

<style>
  button#all-thresholds-button {
    margin: 1rem 0 0 auto;
  }
</style>

{#key reloadKey}
  {#if $loadedLatestThresholds}
    <div class="container-col" transition:fade={{ duration: 500 }}>
      <h2>Threshold Settings</h2>
      <div class="container-col2 m-2">
        <p class="m-0">
          This settings allows you to customize the thresholds of the metrics
          used to assess repository health. For each metric, you can define
          lower and upper thresholds that determine how points are allocated:
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

      <div class="container-small m-2">
        <button
          on:click={() => setNewThresholds("allThresholds")}
          class="button1"
          id="all-thresholds-button"
        >
          Set all Thresholds
        </button>
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                on:input={handleInputValidation}
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
                placeholder={String(
                  $latestThresholds.ThActiveFFSUserCount.lower
                )}
                on:input={handleInputValidation}
              />
            </div>

            <div class="container-col2 mr-10">
              <label for="ffs-user-count-input-upper">upper:</label>
              <input
                id="ffs-user-count-input-upper"
                type="number"
                class="max-w20"
                bind:value={newThActiveFFSUserCountUpper}
                placeholder={String(
                  $latestThresholds.ThActiveFFSUserCount.upper
                )}
                on:input={handleInputValidation}
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
    <p>updating thresholds...</p>
  {/if}
{/key}
