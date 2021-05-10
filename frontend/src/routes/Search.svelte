<script lang="ts">
  import type { Repo } from "../types/users";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import { errorSearch, sponsoredRepos } from "../ts/store";

  import Navigation from "../components/Navigation.svelte";
  import RepoCard from "../components/RepoCard.svelte";
  import SearchResult from "../components/SearchResult.svelte";
  import Spinner from "../components/Spinner.svelte";
  import Dots from "../components/Dots.svelte";

  let search = "";
  let repos: Repo[] = [];
  let isSubmitting = false;

  const handleSearch = async () => {
    try {
      isSubmitting = true;
      repos = await API.repos.search(search);
    } catch (e) {
      $errorSearch = e;
    } finally {
      isSubmitting = false;
    }
  };

  onMount(async () => {
    try {
      $sponsoredRepos = await API.user.getSponsored();
    } catch (e) {
      $errorSearch = e;
    }
  });
</script>

<style>
    .container {
        display: flex;
        flex-direction: row;
        margin: 1em;
    }
    .wrap {
        display: flex;
        flex-wrap: wrap;
    }
    .parent {
        position: relative;
        display: inline-block;
        padding-right: 1.5em;
    }
    .close:before {
        content: '✕';
    }
    .close {
        position: absolute;
        top: 0;
        right: 0;
        padding: 5px;
        cursor: pointer;
    }
</style>

<Navigation>
  {#if $errorSearch}<div class="bg-red rounded p-2 m-4 parent">{$errorSearch}<span class="close" on:click|preventDefault="{() => {$errorSearch=null}}"></span></div>{/if}
  {#if isSubmitting}<Spinner />{/if}

  <h1 class = "px-2">Search Repositories</h1>

  {#if $sponsoredRepos.length === 0}
    <div class="container bg-green rounded p-2 my-4">
      <div>
        <p>Search for your repositories you want to tag. Currently only GitHub search is supported. You can tag as many
          repositories as you want.</p>
        <p>Once you have tagged at least on repository, this message will disappear.</p>
      </div>
    </div>
  {/if}

  <div class="p-2">
    {#if $sponsoredRepos.length > 0}
      <div class="wrap">
        {#each $sponsoredRepos as repo}
          <RepoCard repo="{repo}" class = "child"/>
        {/each}
      </div>
    {/if}

    <h2>Find your favorite opes source projects</h2>

    <div class="py-3">
      <form class="flex" on:submit|preventDefault="{handleSearch}">
        <input type="text" class="rounded py-2 border-primary-900" bind:value="{search}" />
        <button type="submit" disabled="{isSubmitting}">Search{#if isSubmitting}<Dots />{/if}</button>
      </form>
    </div>

    {#if repos.length > 0}
      <h2>Results</h2>
      <div>
        {#each repos as repo}
          <SearchResult repo="{repo}" />
        {/each}
      </div>
    {/if}

  </div>
</Navigation>
