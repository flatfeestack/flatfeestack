<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import {
    error,
    isSubmitting,
    loadedSponsoredRepos,
    loadedTrustedRepos,
    loadedMultiplierRepos,
    sponsoredRepos,
    trustedRepos,
    multiplierSponsoredRepos,
  } from "../ts/mainStore";
  import type { Repo } from "../types/backend";

  import Dots from "../components/Dots.svelte";
  import Navigation from "../components/Navigation.svelte";
  import RepoCard from "../components/RepoCard.svelte";
  import SearchResult from "../components/SearchResult.svelte";
  import { Link } from "svelte-routing";

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
    if (!$loadedSponsoredRepos) {
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
    if (!$loadedMultiplierRepos) {
      try {
        $isSubmitting = true;
        $multiplierSponsoredRepos = await API.user.getMultiplier();
        $loadedMultiplierRepos = true;
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
    {#if $sponsoredRepos.length > 0}
      <div class="m-2 wrap">
        {#each $sponsoredRepos as repo, key (repo.uuid)}
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
      Please note that you need <Link to="/user/payments">credit</Link> on your account
      to support projects. You can still tag them even without any balance, but the
      project will not receive any contributions.
    </p>

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
          <SearchResult {repo} />
        {/each}
      </div>
    {/if}
  </div>
</Navigation>
