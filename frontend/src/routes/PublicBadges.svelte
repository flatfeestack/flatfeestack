<script type="ts">
  import Navigation from "../components/Navigation.svelte";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { error, firstTime } from "../ts/store";
  import type { Contributions, Repo, Users } from "../types/users";
  import { formatDay, formatMUSD } from "../ts/services";
  import confetti from "canvas-confetti";

  export let uuid: string;
  let repos: Repo[] = [];
  let user: Users;

  onMount(async () => {
    try {
      const pr1 = API.user.contributionsSummary2(uuid);
      const pr2 = API.user.summary(uuid);

      const res1 = await pr1;
      const res2 = await pr2;

      repos = res1 ? res1 : repos;
      user = res2;
    } catch (e) {
      $error = e;
    }
  });

</script>

<style>
</style>

<div class="container-col">
<h1 class="px-2">Badges</h1>

{#if repos && repos.length > 0}
  <h2 class="px-2">Supported Repositories for {user.name}</h2>
  {#if user.role != "ORG"}
    <img class="image-usr" src="{user.image}" />
  {:else}
    <img class="image-org" src="{user.image}" />
  {/if}
  <div class="container">
    <table>
      <thead>
      <tr>
        <th>Name</th>
        <th>URL</th>
        <th>Description</th>
      </tr>
      </thead>
      <tbody>
      {#each repos as repo}
        <tr>
          <td>{repo.full_name}</td>
          <td><a href="{repo.html_url}">{repo.html_url}</a></td>
          <td>{repo.description}</td>
        </tr>
      {:else}
        <tr>
          <td colspan="3">No Data</td>
        </tr>
      {/each}
      </tbody>
    </table>
  </div>
{/if}
</div>
