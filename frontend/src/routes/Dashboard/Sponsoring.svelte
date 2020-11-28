<style type="text/scss">
</style>

<script lang="ts">
import DashboardLayout from "../../layout/DashboardLayout.svelte";
import RepoCard from "../../components/UI/SponsoredRepoCard.svelte";
import type { Repo } from "../../types/repo.type";
import { API } from "../../api/api";
import SearchRepoCard from "../../components/UI/SearchRepoCard.svelte";
import Spinner from "../../components/UI/Spinner.svelte";
import { onMount } from "svelte";
import { sponsoredRepos } from "../../store/repos";

let search = "";
let response: Repo[] = [];
let fetching = false;
const onSubmit = async (e) => {
  try {
    e.preventDefault();
    fetching = true;
    const res = await API.repos.search(search);
    response = res.data.data;
    console.log(response);
  } catch (e) {
    console.log("could not fetch");
  } finally {
    fetching = false;
  }
};

onMount(async () => {
  try {
    const res = await API.repos.getSponsored();
    sponsoredRepos.set(res.data.data);
  } catch (e) {
    console.log(e);
  }
});
</script>

<DashboardLayout>
  <h1>Sponsoring</h1>
  <h2>Sponsored Repos</h2>
  <div class="flex flex-wrap overflow-hidden p-3">
    {#if $sponsoredRepos && $sponsoredRepos.length > 0}
      {#each $sponsoredRepos as repo}
        <RepoCard repo="{repo}" />
      {/each}
    {/if}
  </div>
  <h2 class="mt-10">Sponsor new Repos</h2>
  <div class="w-1/3 mt-2">
    <form class="flex" on:submit="{onSubmit}">
      <input type="text" class="input" bind:value="{search}" />
      <button
        type="submit"
        class="button ml-5"
        disabled="{fetching}"
      >Suchen{fetching ? '...' : ''}</button>
    </form>
  </div>
  <div class="w-2/3">
    {#if fetching}
      <Spinner />
    {:else}
      {#each response as repo}
        <SearchRepoCard repo="{repo}" />
      {/each}
    {/if}
  </div>
</DashboardLayout>
