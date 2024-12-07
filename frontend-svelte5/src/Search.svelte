<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "./ts/api.ts";
  import {appState} from "./ts/state.svelte.ts";
  import type { Repo } from "./types/backend";

  import Dots from "./Dots.svelte";
  import RepoCard from "./RepoCard.svelte";
  import SearchResult from "./SearchResult.svelte";
  import {route} from "@mateothegreat/svelte5-router";
  import Main from "./Main.svelte";

  let search = $state("");
  let searchRepos = $state<Repo[]>([]);
  let isSearchSubmitting = $state(false);
  let isSearchDisabled = $state(false);

  $effect(() => {
    isSearchDisabled = search.trim().length === 0 || isSearchSubmitting;
  });

  const handleSearch = async () => {
    try {
      isSearchSubmitting = true;
      searchRepos = await API.repos.search(search);
    } catch (e) {
      appState.setError(e as Error);
    } finally {
      isSearchSubmitting = false;
    }
  };

  onMount(async () => {
    if (!appState.loadedSponsoredRepos) {
      try {
        appState.isSubmitting = true;
        appState.sponsoredRepos = await API.user.getSponsored();
        appState.loadedSponsoredRepos = true;
      } catch (e) {
        appState.setError(e as Error);
      } finally {
        appState.isSubmitting = false;
      }
    }
  });
</script>

<style>
  .wrap {
    display: flex;
    flex-wrap: wrap;
  }
  @media screen and (max-width: 600px) {
    h2,
    p {
      word-break: break-word;
    }
  }
</style>

<Main>
  <div class="p-2">
    {#if appState.sponsoredRepos.length > 0}
      <div class="m-2 wrap">
        {#each appState.sponsoredRepos as repo (repo.uuid)}
          <RepoCard {repo} />
        {/each}
      </div>
    {/if}

    <h2 class="m-2">Find your favorite open-source projects</h2>
    <p class="m-2">
      Search for repositories you want to support. Only the GitHub search is
      supported now, but you can enter a full URL (like
      https://gitlab.com/fdroid/fdroiddata) into the field to find a repository
      hosted elsewhere.
    </p>

    <p class="m-2">You can tag as many repositories as you want.</p>

    <p class="m-2">
      Please note that you need <a use:route href="/user/payments">credit</a> on your account
      to support projects. You can still tag them even without any balance, but the
      project will not receive any contributions.
    </p>

    <div class="flex">
      <input type="text" bind:value={search}/>
      <button class="button1 ml-5" onclick={handleSearch} disabled={isSearchDisabled} aria-label="Search">
        Search{#if isSearchSubmitting}<Dots />{/if}
      </button>
    </div>

    {#if searchRepos?.length > 0}
      <h2 class="m-2">Results</h2>
      <div>
        {#each searchRepos as repo (repo.uuid)}
          <SearchResult {repo} />
        {/each}
      </div>
    {/if}
  </div>
</Main>
