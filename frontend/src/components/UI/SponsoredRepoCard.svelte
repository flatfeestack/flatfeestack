<style type="text/scss">
.repocard {
  @apply shadow-lg p-5 overflow-hidden mx-2 my-2;
  min-width: 200px;
  max-width: 400px;
}
.body {
  @apply text-gray-400;
}
</style>

<script lang="ts">
import type { Repo } from "../../types/repo.type";
import Fa from "svelte-fa";
import { faGitAlt } from "@fortawesome/free-brands-svg-icons";
import { API } from "../../api/api";
import { sponsoredRepos } from "../../store/repos";
export let repo: Repo;

const unsponsor = async () => {
  try {
    const res = await API.repos.unsponsor(repo.id);
    const repoId = res.data.data.repo_id;
    if (!repoId) {
      throw Error("No repo id");
    }
    sponsoredRepos.set(
      $sponsoredRepos.filter((r) => {
        return r.id !== parseInt(repoId);
      })
    );
    console.log(res.data.data.repo_id);
  } catch (e) {
    console.log(e);
  }
};
</script>

<div class="repocard flex flex-col">
  <h3 class="font-semibold text-xl text-primary-500">{repo.full_name}</h3>
  <div class="body">{repo.description}</div>
  <div>
    <a
      href="{repo.html_url}"
      target="_blank"
      class="text-primary-500 inline-block"
    >
      <Fa icon="{faGitAlt}" size="lg" />
    </a>
  </div>
  <div>
    <button
      class="border-red-500 border-2 inline-block py-1 px-2 text-red-500 rounded hover:bg-red-500 hover:text-white text-sm"
      on:click="{unsponsor}"
    >Unsponsor</button>
  </div>
</div>
