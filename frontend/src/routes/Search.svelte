<script lang="ts">
  import type { Repo } from "../types/users";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import { error, isSubmitting, loadedSponsoredRepos, sponsoredRepos } from "../ts/store";

  import Navigation from "../components/Navigation.svelte";
  import RepoCard from "../components/RepoCard.svelte";
  import SearchResult from "../components/SearchResult.svelte";
  import Dots from "../components/Dots.svelte";

  let search = "";
  let repos: Repo[] = [];
  let isSearchSubmitting = false;

  const handleSearch = async () => {
    try {
      isSearchSubmitting = true;
      repos = await API.repos.search(search);
    } catch (e) {
      $error = e;
    } finally {
      isSearchSubmitting = false;
    }
  };

  onMount(async () => {
    if(!$loadedSponsoredRepos) {
      try {
        $isSubmitting = true;
        $sponsoredRepos = await API.user.getSponsored();
        $loadedSponsoredRepos = true;
      } catch (e) {
        $error = e;
      } finally {
        $isSubmitting = false;
      }
    }
  });
</script>

<style>
    .container {
        display: flex;
        flex-direction: row;
    }
    .wrap {
        display: flex;
        flex-wrap: wrap;
    }
</style>

<Navigation>
  <h1 class="px-2">Search Repositories</h1>

  {#if $sponsoredRepos.length === 0}
    <div class="container bg-green rounded p-2 m-2">
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
        {#each $sponsoredRepos as repo, key (repo.uuid)}
          <RepoCard repo="{repo}" class = "child"/>
        {/each}
      </div>
    {/if}

    <h2>Find your favorite opes source projects</h2>

    <div class="py-3">
      <form class="flex" on:submit|preventDefault="{handleSearch}">
        <input type="text" class="rounded py-2 border-primary-900" bind:value="{search}" />
        <button type="submit" disabled="{isSearchSubmitting}">Search{#if isSearchSubmitting}<Dots />{/if}</button>
      </form>
    </div>

    {#if repos.length > 0}
      <h2>Results</h2>
      <div>
        {#each repos as repo, key (repo.id)}
          <SearchResult repo="{repo}" />
        {/each}
      </div>
    {/if}

  </div>
</Navigation>
