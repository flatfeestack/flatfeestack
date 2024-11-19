<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "../../ts/api";
  import {
    faCaretUp,
    type IconDefinition,
  } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  import {
    error,
    isSubmitting,
    loadedTrustedRepos,
    trustedRepos,
  } from "../../ts/mainStore";
  import type { Repo } from "../../types/backend";
  import Dots from "../../components/Dots.svelte";
  import Navigation from "../../components/Navigation.svelte";
  import AdminSearchResult from "../../components/AdminSearchResult.svelte";
  import TrustedRepoCard from "../../components/HealthyRepoCard.svelte";
  // import { Link } from "svelte-routing";

  let icon: IconDefinition;
  let search = "";
  let searchRepos: Repo[] = [];
  let isSearchSubmitting = false;
  let sortingFunction: (a: Repo, b: Repo) => number;
  let sortingTitle: string;

  $: isSearchDisabled = search.trim().length === 0 || isSearchSubmitting;

  $: sortedTrustedRepos = $trustedRepos
    .slice()
    .sort(sortingFunction)
    .slice(0, 50);

  const handleSearch = async () => {
    try {
      isSearchSubmitting = true;
      searchRepos = await API.repos.search(search);
    } catch (e) {
      $error = e;
    } finally {
      isSearchSubmitting = false;
    }
    // console.log('sortedtrustedRepos', sortedTrustedRepos);
  };

  function sortByName(a: Repo, b: Repo) {
    return a.name?.localeCompare(b.name || "") || 0;
  }

  function sortByDate(a: Repo, b: Repo, ascending: boolean = true): number {
    const dateA = new Date(a.trustAt).getTime();
    // console.log('dateA:', dateA)
    const dateB = new Date(b.trustAt).getTime();
    // console.log('dateB:', dateB)
    return ascending ? dateA - dateB : dateB - dateA;
  }
  function sortByDateAsc(a: Repo, b: Repo) {
    return sortByDate(a, b, true); // Ascending
  }
  function sortByDateDesc(a: Repo, b: Repo) {
    return sortByDate(a, b, false); // Descending
  }

  // function sortByScore(a: Repo, b: Repo, ascending: boolean = true): number {
  //   const scoreA =
  //   const scoreB =
  //   return ascending ? scoreA - scoreB : scoreB - scoreA;
  // }
  // function sortByScoreAsc(a: Repo, b: Repo) {
  //   return sortByScore(a, b, true); // Ascending
  // }
  // function sortByScoreDesc(a: Repo, b: Repo) {
  //   return sortByScore(a, b, false); // Descending
  // }

  function sortByScore(a: Repo, b: Repo) {
    return (b.score || 0) - (a.score || 0);
  }

  // Set default sorting function
  sortingFunction = sortByDateDesc;
  sortingTitle = "Recently Added:";

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
    // console.log('trustedRepos', $trustedRepos);
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

  .search-container {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
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
      <h2 class="p-2 m-2">Healthy Repositories</h2>

      <div class="container-col2 m-4">
        <div class="container-small">
          <div class="container-small m-2 dropdown">
            <button class="button1 drop-button" id="drop-button"
              ><Fa icon={faCaretUp} /> Sort</button
            >
            <div class="dropdown-content">
              <button
                on:click={() => {
                  sortingFunction = sortByDateDesc;
                  sortingTitle = "Recently Added:";
                }}>Sort by Date - Recently Added</button
              >
              <button
                on:click={() => {
                  sortingFunction = sortByDateAsc;
                  sortingTitle = "First Added:";
                }}>Sort by Date - First Added</button
              >
              <button
                on:click={() => {
                  sortingFunction = sortByDateDesc;
                  sortingTitle = "Score - high to low:";
                }}>Sort by Score - high to low</button
              >
              <button
                on:click={() => {
                  sortingFunction = sortByDateAsc;
                  sortingTitle = "Score - low to high:";
                }}>Sort by Score - low to high</button
              >
            </div>
          </div>
          <h3 class="m-2">{sortingTitle}</h3>
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
                <AdminSearchResult {repo} />
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</Navigation>
