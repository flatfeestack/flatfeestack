<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "@/api";
  import { error } from "@/mainStore";
  import type { ContributionSummary, User } from "@/types/backend";

  export let uuid: string;
  let contributionSummaries: ContributionSummary[] = [];
  let user: User;

  onMount(async () => {
    try {
      const pr1 = API.user.contributionsSummary2(uuid);
      const pr2 = API.user.summary(uuid);

      const res1 = await pr1;
      const res2 = await pr2;

      contributionSummaries = res1 || contributionSummaries;
      user = res2;
    } catch (e) {
      $error = e;
    }
  });
</script>

<style>
</style>

<div class="container-col">
  {#if contributionSummaries && contributionSummaries.length > 0}
    <h2 class="px-2">
      Supported Repositories for {user.name || user.email}
    </h2>
    {#if user.image}
      <img class="image-org" src={user.image} alt="supported user repository" />
    {/if}
    <div class="container">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Repos</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          {#each contributionSummaries as cs}
            <tr>
              <td><a href={cs.repo.url}>{cs.repo.name}</a></td>
              <td>
                <a href={cs.repo.gitUrl}>{cs.repo.gitUrl}</a>
              </td>
              <td>{cs.repo.description}</td>
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
