<script lang="ts">
  import type { Repos } from "../types/users";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import { error, isSubmitting, loadedSponsoredRepos, sponsoredRepos } from "../ts/store";

  import Navigation from "../components/Navigation.svelte";
  import RepoCard from "../components/RepoCard.svelte";
  import SearchResult from "../components/SearchResult.svelte";
  import Dots from "../components/Dots.svelte";

  let search = "";
  let searchRepos: Repos[] = [];
  let isSearchSubmitting = false;

  const handleSearch = async () => {
    try {
      isSearchSubmitting = true;
      searchRepos = await API.repos.search(search);
      console.log("searchRepos", searchRepos)
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
        console.log("$sponsoredRepos", $sponsoredRepos)
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
    .wrap {
        display: flex;
        flex-wrap: wrap;
    }
</style>

<Navigation>

  <div class="p-2">
    {#if $sponsoredRepos.length > 0}
      <div class="wrap">
        {#each $sponsoredRepos as repos, key (repos.uuid)}
          <RepoCard repos="{repos}" class = "child"/>
        {/each}
      </div>
    {/if}

    <h2 class="p-2 m-2">Find your favorite opes source projects</h2>
    {#if $sponsoredRepos.length === 0}
          <p class="p-2 m-2">Search for your repositories you want to tag. Currently only GitHub search is supported. You can tag as many
            repositories as you want.</p>
    {/if}
    <div class="p-2 m-2">
      <form class="flex" on:submit|preventDefault="{handleSearch}">
        <input type="text" bind:value="{search}" />
        <button class="button1" type="submit" disabled="{isSearchSubmitting}">Search{#if isSearchSubmitting}<Dots />{/if}</button>
      </form>
    </div>

    {#if searchRepos.length > 0}
      <h2>Results</h2>
      <div>
        {#each searchRepos as repos, key (repos.uuid)}
          <SearchResult repos="{repos}" />
        {/each}
      </div>
    {/if}

  </div>
</Navigation>
