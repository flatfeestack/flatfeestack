<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { API } from "../../ts/api";
  import {
    faCaretUp,
    faClose,
    faGear,
    type IconDefinition,
  } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  import {
    error,
    isSubmitting,
    loadedTrustedRepos,
    trustedRepos,
    reloadAdminSearchKey,
    reposWaitingForNewAnalysis,
  } from "../../ts/mainStore";
  import type { Repo } from "../../types/backend";
  import Dots from "../../components/Dots.svelte";
  import Navigation from "../../components/Navigation.svelte";
  import AdminSearchResult from "../../components/AdminSearchResult.svelte";
  import TrustedRepoCard from "../../components/HealthyRepoCard.svelte";
  import ThresholdSettingsOverlay from "../../components/ThresholdSettingsOverlay.svelte";

  let icon: IconDefinition;
  let search = "";
  let healthyRepoSearch: string = "";
  let searchRepos: Repo[] = [];
  let isSearchSubmitting = false;
  let sortingFunction: (a: Repo, b: Repo) => number;
  let sortingTitle: string;
  let amountOfShownRepos: number = 50;
  let intervalId: any;
  let thresholdSettingsOverlayVisible: boolean = false;

  $: isSearchDisabled = search.trim().length === 0 || isSearchSubmitting;

  $: sortedTrustedRepos = $trustedRepos
    .slice()
    .filter((repo) =>
      repo.name?.toLowerCase().includes(healthyRepoSearch.trim().toLowerCase())
    )
    .sort(sortingFunction)
    .slice(0, amountOfShownRepos);

  $: {
    if ($reposWaitingForNewAnalysis.length > 0) {
      startBackendPolling();
    } else {
      stopBackendPolling();
    }
  }

  async function startBackendPolling() {
    if ($reloadAdminSearchKey === 0) {
      if (intervalId) clearInterval(intervalId);
      intervalId = setInterval(async () => {
        reloadAdminSearchKey.update((key) => key + 1);
        try {
          searchRepos = await API.repos.search(search);
        } catch (e) {
          $error = e;
        }
      }, 5000);
    }
  }

  function stopBackendPolling() {
    if ($reposWaitingForNewAnalysis.length === 0) {
      clearInterval(intervalId);
      intervalId = null;
      reloadAdminSearchKey.update(() => 0);
    } else {
    }
  }

  const handleSearch = async () => {
    try {
      isSearchSubmitting = true;
      searchRepos = await API.repos.search(search);
    } catch (e) {
      $error = e;
    } finally {
      isSearchSubmitting = false;
    }
  };

  function sortByDate(a: Repo, b: Repo, ascending: boolean = true): number {
    const dateA = new Date(a.trustAt).getTime();
    const dateB = new Date(b.trustAt).getTime();
    return ascending ? dateA - dateB : dateB - dateA;
  }
  function sortByDateAsc(a: Repo, b: Repo) {
    return sortByDate(a, b, true);
  }
  function sortByDateDesc(a: Repo, b: Repo) {
    return sortByDate(a, b, false);
  }

  function sortByScore(a: Repo, b: Repo, ascending: boolean = true): number {
    const scoreA = a.healthValue;
    const scoreB = b.healthValue;
    return ascending ? scoreA - scoreB : scoreB - scoreA;
  }
  function sortByScoreAsc(a: Repo, b: Repo) {
    return sortByScore(a, b, true);
  }
  function sortByScoreDesc(a: Repo, b: Repo) {
    return sortByScore(a, b, false);
  }

  // Set default sorting function
  sortingFunction = sortByDateDesc;
  sortingTitle = "Date - Recently Added";

  function showOverlay() {
    thresholdSettingsOverlayVisible = true;
  }

  function hideOverlay() {
    thresholdSettingsOverlayVisible = false;
  }

  onMount(async () => {
    if (!$loadedTrustedRepos) {
      try {
        $isSubmitting = true;
        $trustedRepos = await API.repos.getTrusted();
        $loadedTrustedRepos = true;
      } catch (e) {
        $error = e;
      } finally {
        $isSubmitting = false;
      }
    }
  });

  onDestroy(async () => {
    reloadAdminSearchKey.update(() => 0);
    if (intervalId) clearInterval(intervalId);
    reposWaitingForNewAnalysis.update(() => []);
  });
</script>

<style>
  .grid-healthy-repos {
    display: grid;
    grid-template-columns: 100%;
    grid-template-rows: auto calc(
        100vh - 2rem - 1rem - (1.2rem * 1.2) - 1rem - (1.1rem * 1.2) - 319px
      );
    height: calc(100vh - 2rem - 1rem - (1.2rem * 1.2) - 1rem - (1.1rem * 1.2));
  }

  .healty-repos-div {
    grid-column: 1 / 2;
    grid-row: 1 / 2;
  }

  .add-healthy-repos-div {
    grid-column: 1 / 2;
    grid-row: 2 / 3;
    height: calc(
      100vh - 2rem - 1rem - (1.2rem * 1.2) - 1rem - (1.1rem * 1.2) - 319px
    );
    overflow-y: auto;
    overflow-x: hidden;
    -webkit-overflow-scrolling: touch;
  }

  .add-healthy-repos-div::-webkit-scrollbar {
    width: 0.5rem;
    background: #f1f1f1;
  }

  .add-healthy-repos-div::-webkit-scrollbar-thumb {
    background: #888;
    border-radius: 4px;
  }

  .add-healthy-repos-div::-webkit-scrollbar-thumb:hover {
    background: #555;
  }

  .cards-overflow-x {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    width: 100%;
    overflow-x: auto;
    overflow-y: hidden;
    -webkit-overflow-scrolling: touch;
  }
  .cards-overflow-x::-webkit-scrollbar {
    height: 0.5rem;
    background: #f1f1f1;
  }

  .cards-overflow-x::-webkit-scrollbar-thumb {
    background: #888;
    border-radius: 4px;
  }

  .cards-overflow-x::-webkit-scrollbar-thumb:hover {
    background: #555;
  }

  button#threshold-setting-button {
    padding: 0.25rem;
    font-size: 2rem;
    height: 2.6rem;
    width: 2.6rem;
    border: none;
    transition: color 0.5s, background-color 0.5s;
  }
  button#threshold-setting-button:hover {
    color: var(--secondary-100);
    background: var(--secondary-300);
  }

  .dropdown-content-sort {
    min-width: 16rem;
    max-width: 20rem;
  }
  .dropdown-content-amount-filter {
    min-width: 4.5rem;
    max-width: 4.5rem;
  }

  #amount-drop-button {
    min-width: 4.5rem;
    max-width: 4.5rem;
  }

  .separator-y {
    height: 3rem;
    width: 2px;
    background-color: var(--secondary-300);
    border-radius: 5px;
  }

  .search-container {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
  }

  .threshold-overlay {
    position: fixed;
    display: block;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: rgba(
      221,
      221,
      221,
      0.3
    ); /* secondary-200 with 30% opacity */
    z-index: 2;
  }

  .overlay-container {
    position: absolute;
    width: 60vw;
    height: 90vh;
    background-color: white;
    color: black;
    overflow-y: auto;
    margin: 5vh 20vw;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    border-radius: 10px;
  }

  #close-overlay-button {
    position: absolute;
    top: 5vh;
    right: 21vw;
    z-index: 3;
  }

  @media screen and (min-width: 2000px) {
    .overlay-container {
      width: 1185px;
      margin: 5vh calc((100vw - 1185px) / 2);
    }
    #close-overlay-button {
      right: calc(((100vw - 1185px) / 2) + 1vw);
    }
  }

  @media screen and (max-width: 600px) {
    form {
      flex-direction: column;
    }
    form button {
      margin: 0.5rem 0;
    }
    h2,
    p {
      word-break: break-word;
    }
  }
</style>

<Navigation>
  <div class="grid-healthy-repos">
    <div class="healty-repos-div">
      <div class="container-small justify-between">
        <h2 class="p-2 m-2">Healthy Repositories</h2>
        <button
          class="button3"
          id="threshold-setting-button"
          title="Threshold Settings Button"
          on:click={showOverlay}><Fa icon={faGear} class="m-0 p-0" /></button
        >
      </div>

      <div class="container-col2 m-4">
        <div class="container-small justify-between w-100">
          <div class="container-small">
            <h3 class="m-2">Sort by</h3>
            <div class="container-small m-2 dropdown">
              <button class="button1 drop-button" id="drop-button"
                ><Fa icon={faCaretUp} /> {sortingTitle}</button
              >
              <div class="dropdown-content dropdown-content-sort">
                <button
                  on:click={() => {
                    sortingFunction = sortByDateDesc;
                    sortingTitle = "Date - Recently Added";
                  }}>Date - Recently Added</button
                >
                <button
                  on:click={() => {
                    sortingFunction = sortByDateAsc;
                    sortingTitle = "Date - First Added";
                  }}>Date - First Added</button
                >
                <button
                  on:click={() => {
                    sortingFunction = sortByScoreDesc;
                    sortingTitle = "Score - high to low";
                  }}>Score - high to low</button
                >
                <button
                  on:click={() => {
                    sortingFunction = sortByScoreAsc;
                    sortingTitle = "Score - low to high";
                  }}>Score - low to high</button
                >
              </div>
            </div>
          </div>

          <div class="container-small">
            <input
              type="text"
              bind:value={healthyRepoSearch}
              placeholder="Search Healthy Repos"
            />

            <div class="container-small m-2 dropdown">
              <button class="button1 drop-button" id="amount-drop-button"
                ><Fa icon={faCaretUp} /> {amountOfShownRepos}</button
              >
              <div class="dropdown-content dropdown-content-amount-filter">
                <button
                  on:click={() => {
                    amountOfShownRepos = 10;
                  }}>10</button
                >
                <button
                  on:click={() => {
                    amountOfShownRepos = 25;
                  }}>25</button
                >
                <button
                  on:click={() => {
                    amountOfShownRepos = 50;
                  }}>50</button
                >
                <button
                  on:click={() => {
                    amountOfShownRepos = 100;
                  }}>100</button
                >
              </div>
            </div>
          </div>
        </div>

        {#if $trustedRepos.length > 0}
          <div class="cards-overflow-x">
            {#each sortedTrustedRepos as repo, key (repo.uuid)}
              <TrustedRepoCard {repo} />
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <div class="add-healthy-repos-div">
      <h2 class="p-2 m-2">Add new Healthy Repositories</h2>

      <div class="container-col2 m-4">
        <form class="flex m-2" on:submit|preventDefault={handleSearch}>
          <input
            type="text"
            bind:value={search}
            placeholder="Search all Repos"
          />
          <button class="button1 ml-5" type="submit" disabled={isSearchDisabled}
            >Search{#if isSearchSubmitting}<Dots />{/if}</button
          >
        </form>
        {#if searchRepos?.length > 0}
          <div class="search-container m-2">
            <h3 class="m-2">Results</h3>
            <div class="wrap-container">
              {#each searchRepos as repo, key (repo.uuid)}
                {#key $reloadAdminSearchKey}
                  <AdminSearchResult {repo} />
                {/key}
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <div
    class:threshold-overlay={thresholdSettingsOverlayVisible}
    class:hidden={!thresholdSettingsOverlayVisible}
    role="button"
    tabindex="0"
  >
    <button
      id="close-overlay-button"
      class="button3 m-2 px-100"
      on:click={hideOverlay}>close <Fa icon={faClose} class="ml-5" /></button
    >
    <div class="overlay-container">
      {#if thresholdSettingsOverlayVisible}
        <ThresholdSettingsOverlay />
      {/if}
    </div>
  </div>
</Navigation>
