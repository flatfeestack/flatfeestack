<style type="text/scss">
.repocard {
  @apply shadow-md p-5 overflow-hidden my-4;
  &:hover {
    @apply bg-primary-100;
  }
  min-width: 200px;
}
.body {
  @apply text-gray-400 w-full;
}
</style>

<script lang="ts">
import type { Repo } from "../../types/repo.type";
import { API } from "../../api/api";
import { sponsoredRepos } from "../../store/repos";

export let repo: Repo;

const onSponsor = async () => {
  try {
    const res = await API.repos.sponsor(repo.id);
    sponsoredRepos.set([...$sponsoredRepos, res.data]);
  } catch (e) {
    console.log(e);
  }
};
</script>

<div class="repocard flex">
  <div class="flex-1">
    <h2>{repo.full_name}</h2>
    <div class="body">{repo.description}</div>
  </div>
  <div class="flex items-center">
    <button class="button" on:click="{onSponsor}">Sponsor</button>
  </div>
</div>
