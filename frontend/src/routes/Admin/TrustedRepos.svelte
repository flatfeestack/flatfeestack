<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "../../ts/api";
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
  // import SearchResult from "../../components/SearchResult.svelte";
  // import { Link } from "svelte-routing";

  let search = "";
  let searchRepos: Repo[] = [];
  let isSearchSubmitting = false;

  $: isSearchDisabled = search.trim().length === 0 || isSearchSubmitting;

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
</script>

<style>
  .wrap {
    display: flex;
    flex-wrap: wrap;
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
    <div class="m-2">
      <form class="flex" on:submit|preventDefault={handleSearch}>
        <input type="text" bind:value={search} />
        <button class="button1 ml-5" type="submit" disabled={isSearchDisabled}
          >Search{#if isSearchSubmitting}<Dots />{/if}</button
        >
      </form>
    </div>

    {#if searchRepos?.length > 0}
      <h2 class="m-2">Results</h2>
      <div>
        {#each searchRepos as repo, key (repo.uuid)}
          <AdminSearchResult {repo} />
        {/each}
      </div>
    {/if}
  </div>
</Navigation>
