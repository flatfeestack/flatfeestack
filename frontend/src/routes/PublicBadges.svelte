<script type="ts">
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { error } from "../ts/mainStore";
  import type { Repos, Users } from "../types/users";

  export let uuid: string;
  let repos: Repos[] = [];
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
  {#if repos && repos.length > 0}
    <h2 class="px-2">
      Supported Repositories for {user.name ? user.name : user.id}
    </h2>
    {#if user.image}
      <img class="image-org" src={user.image} alt="supported user repository" />
    {/if}
    <div class="container">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>URL</th>
            <th>Repos</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          {#each repos as repo}
            <tr>
              <td>{repo.repos[0].name}</td>
              <td><a href={repo.repos[0].url}>{repo.repos[0].url}</a></td>
              <td>
                {#each repo.repos as r2}
                  <a href={r2.gitUrl}>{r2.gitUrl}</a>
                {/each}
              </td>
              <td>{repo.repos[0].description}</td>
            </tr>
          {:else}
            <tr>
              <td colspan="4">No Data</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
