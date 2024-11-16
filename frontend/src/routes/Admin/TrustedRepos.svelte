<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "../../ts/api";
  import {
    error,
    isSubmitting,
    loadedTrustedRepos,
    sponsoredRepos,
    trustedRepos,
  } from "../../ts/mainStore";
  import type { Repo } from "../../types/backend";

  import Dots from "../../components/Dots.svelte";
  import Navigation from "../../components/Navigation.svelte";
  import AdminSearchResult from "../../components/AdminSearchResult.svelte";
  import TrustedRepoCard from "../../components/TrustedRepoCard.svelte";
  // import { Link } from "svelte-routing";

  let search = "";
  let searchRepos: Repo[] = [];
  let isSearchSubmitting = false;
  let sortingFunction: (a: Repo, b: Repo) => number;

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
  .dropdown {
    margin: 0.5rem 0.5rem 0.5rem 0;
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
    height: 8px; /* Height of the scrollbar */
    background: #f1f1f1; /* Scrollbar track color */
  }

  .cards-overflow-x::-webkit-scrollbar-thumb {
    background: #888; /* Scrollbar thumb color */
    border-radius: 4px; /* Rounded corners for the thumb */
  }

  .cards-overflow-x::-webkit-scrollbar-thumb:hover {
    background: #555; /* Darker thumb on hover */
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
  <div class="p-2">
    <h2>Healthy Repositories</h2>
    <div class="container-small">
      <div class="container-small dropdown">
        <button class="button1 drop-button" id="drop-button">Sort</button>
        <div class="dropdown-content">
          <button on:click={() => (sortingFunction = sortByDateDesc)}
            >Sort by Date Desc</button
          >
          <button on:click={() => (sortingFunction = sortByDateAsc)}
            >Sort by Date Asc</button
          >
          <button on:click={() => (sortingFunction = sortByDate)}
            >Sort by Score</button
          >
        </div>
      </div>
      <h3 class="m-2">Recently Added:</h3>
    </div>
    {#if $trustedRepos.length > 0}
      <div class="m-2 cards-overflow-x">
        {#each sortedTrustedRepos as repo, key (repo.uuid)}
          <TrustedRepoCard {repo} />
        {/each}
      </div>
    {/if}

    <h2>Add new Healthy Repositories</h2>
    <div class="m-2">
      <form class="flex" on:submit|preventDefault={handleSearch}>
        <input type="text" bind:value={search} />
        <button class="button1 ml-5" type="submit" disabled={isSearchDisabled}
          >Search{#if isSearchSubmitting}<Dots />{/if}</button
        >
      </form>
    </div>

    {#if searchRepos?.length > 0}
      <h3 class="m-2">Results</h3>
      <div>
        {#each searchRepos as repo, key (repo.uuid)}
          <AdminSearchResult {repo} />
        {/each}
      </div>
    {/if}
  </div>
</Navigation>
