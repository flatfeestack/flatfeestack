<script lang="ts">
  import DashboardLayout from "./DashboardLayout.svelte";
  import RepoCard from "../../components/UI/RepoCard.svelte";
  import type { Repo } from "../../types/repo.type";
  import { API } from "ts/api";
  import SearchRepoCard from "../../components/UI/SearchRepoCard.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import { onMount } from "svelte";
  import { sponsoredRepos } from "ts/repos";
  import { user } from "ts/auth";
  import Dots from "../../components/Dots.svelte";
  import { links } from "svelte-routing";

  let checked = $user.mode != "ORGANIZATION";
  $: {
    if (checked == false) {
      $user.mode = "ORGANIZATION";
    } else {
      $user.mode = "CONTRIBUTOR";
    }
  }

  let search = "";
  let response: Repo[] = [];
  let isSubmitting = false;
  const onSubmit = async () => {
    try {
      isSubmitting = true;
      const res = await API.repos.search(search);
      response = res.data;
      console.log(response);
    } catch (e) {
      console.log("could not fetch");
    } finally {
      isSubmitting = false;
    }
  };

  onMount(async () => {
    try {
      const res = await API.user.getSponsored();
      if (res.data) {
        sponsoredRepos.set(res.data);
      }
    } catch (e) {
      console.log(e);
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
</style>

<DashboardLayout>
  <h1 class = "px-2">Find Repesitories</h1>
    <div class="container bg-green rounded p-2 my-4" use:links>
      <div>
        <p>If you are an organization, switch to organization mode, and in the <a href="/dashboard/profile">Profile</a> section you can
          invite others that can support awesome open source projects. In the organization mode, you can provide examples of cool projects,
          but you cannot sponsor as an organization directly.</p>

        <p>You can also switch now the modes (contributor/organization):</p>
        <div class="onoffswitch">
          <input type="checkbox" bind:checked={checked} name="onoffswitch" class="onoffswitch-checkbox"
                 id="myonoffswitch" tabindex="0">
          <label class="onoffswitch-label" for="myonoffswitch">
            <span class="onoffswitch-inner"></span>
            <span class="onoffswitch-switch"></span>
          </label>
        </div>
      </div>
      <div class="xbut"></div>
    </div>

  <div class="p-2">


    {#if $sponsoredRepos.length > 0}
    <div class="wrap">
      {#each $sponsoredRepos as repo}
        <RepoCard repo="{repo}" class = "child"/>
      {/each}
    </div>
    {/if}

    {#if !$user.mode || $user.mode == "" || $user.mode == "CONTRIBUTOR"}
      <h2>Find your favorite opes source projects</h2>
    {:else}
      <h2>Examples of cool opes source projects</h2>
      These examples can be sent as an example to your invites
    {/if}

    <div class = "py-3">
      <form class="flex" on:submit|preventDefault="{onSubmit}">
        <input type="text" class="rounded py-2 border-primary-700" bind:value="{search}" />
        <button type="submit" disabled="{isSubmitting}">Search
          {#if isSubmitting}
            <Dots />
          {/if}
        </button>
      </form>
    </div>

    {#if response.length > 0}
      <h2 class="mt-10">Result</h2>
    {/if}

    <div class = "flex testme">
      {#if isSubmitting}
        <Spinner />
      {:else}
        {#each response as repo}
          <SearchRepoCard repo="{repo}" />
        {/each}
      {/if}
    </div>

  </div>
</DashboardLayout>
