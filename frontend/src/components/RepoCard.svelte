<script lang="ts">
  import type {Repos, Repo} from "../types/users";
  import { API } from "../ts/api";
  import { error, sponsoredRepos } from "../ts/store";
  import { getColor1, getColor2 } from "../ts/utils";

  export let repos: Repos;
  let star = true;
  let rootRepo;

  $: {
    rootRepo = repos.repos.find((r: Repo) => {
      return repos.uuid === r.uuid
    })
  }

  async function unTag () {
    star = false;
    try {
      await API.repos.untag(repos.uuid);
      $sponsoredRepos = $sponsoredRepos.filter((r: Repos) => {return r.uuid !== repos.uuid;});
    } catch (e) {
      $error = e;
    }
  }

</script>

<style>
.child {
    flex: 1 0;
    margin: 0.5em;
    max-width: 18em;
    min-width: 18em;
    box-shadow:  0.25em 0.25em 0.25em #e1e1e3;
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
}
.color {
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
    height: 3.5em;
    box-shadow: 0 3px 2px -2px black;
}
.center2{
    font-weight: bold;
    text-overflow: ellipsis;
}
.body {
    text-align: center;
    font-size: medium;
}
.url {
    text-align: center;
    font-size: small;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
}
svg {
    padding: 0.25em;
    height: 1em;
    width: 1em;
}

.color :global(a:hover) {
    filter: drop-shadow( 2px 2px 2px rgba(0, 0, 0, .7));
}

</style>

<div class="child rounded">
  <div class="color" style="background-image:radial-gradient(circle at top right,{getColor2(repos.uuid)},{getColor1(repos.uuid)});">

    <div>
      {#if star}
        <a href="#" on:click|preventDefault="{unTag}">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 573.655 550.909" color="gold" overflow="visible">
            <path
              d="M258.128 36.96l-65.3 132.4-146.1 21.3c-26.2 3.8-36.7 36.1-17.7 54.6l105.7 103-25 145.5c-4.5 26.3 23.2 46 46.4 33.7l130.7-68.7 130.7 68.7c23.2 12.2 50.9-7.4 46.4-33.7l-25-145.5 105.7-103c19-18.5 8.5-50.8-17.7-54.6l-146.1-21.3-65.3-132.4c-11.7-23.6-45.6-23.9-57.4 0z"
              fill="gold" stroke="gold" stroke-width="40" />
          </svg>
        </a>
      {:else}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 573.655 550.909" color="gold" overflow="visible">
          <path
            d="M258.128 36.96l-65.3 132.4-146.1 21.3c-26.2 3.8-36.7 36.1-17.7 54.6l105.7 103-25 145.5c-4.5 26.3 23.2 46 46.4 33.7l130.7-68.7 130.7 68.7c23.2 12.2 50.9-7.4 46.4-33.7l-25-145.5 105.7-103c19-18.5 8.5-50.8-17.7-54.6l-146.1-21.3-65.3-132.4c-11.7-23.6-45.6-23.9-57.4 0z"
            fill="none" stroke="gold" stroke-width="40" />
        </svg>
      {/if}
    </div>

  </div>
  {#if repos}
  <div class="center center2 py-2">{rootRepo.name}</div>
  <div class="body">{rootRepo.description}</div>
  <div>
    <a href="{rootRepo.url}" class="py-2 url" target="_blank" >
      {rootRepo.url}
    </a>
  </div>
    <div class="small center">
    Git URLs: ({#each repos.repos as repos2, i}
      {i==0? repos2.gitUrl: ", "+repos2.gitUrl}
    {/each})
  </div>
    {/if}
</div>
