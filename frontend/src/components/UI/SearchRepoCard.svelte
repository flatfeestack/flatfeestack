<style>
    .container {
        display: flex;
        flex-direction: row;
        margin-bottom: 2em;

    }

    svg {
        padding: 0.25em;
        height: 1em;
        width: 1em;
    }
    .url {
        font-size: 0.8em;
    }

    .container :global(a:hover) {
        filter: drop-shadow( 2px 2px 2px rgba(0, 0, 0, .7));
    }

</style>

<script lang="ts">
  import type { Repo } from "../../types/repo.type";
  import { API } from "../../ts/api";
  import { sponsoredRepos } from "../../ts/repos";

  export let repo: Repo;
  let star = false;

  const onSponsor = async () => {
    try {
      star = true;
      const res = await API.repos.tag(repo);
      $sponsoredRepos = [...$sponsoredRepos, res.data];
    } catch (e) {
      star = false;
      console.log(e);
    }
  };

  sponsoredRepos.subscribe(() => {
    const tmp = $sponsoredRepos.find((r: Repo) => {
      return r.clone_url === repo.clone_url;
    });

    star = tmp !== undefined;
  })

  const getColor = function(input: string) {
    return "hsl(" + 360 * cyrb53(input+"a") + ',' +
      (25 + 60 * cyrb53(input+"b")) + '%,' +
      (60 + 20 * cyrb53(input+"c")) + '%)'
  }

  //https://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript?rq=1
  const cyrb53 = function(str, seed = 0) {
    let h1 = 0xdeadbeef ^ seed, h2 = 0x41c6ce57 ^ seed;
    for (let i = 0, ch; i < str.length; i++) {
      ch = str.charCodeAt(i);
      h1 = Math.imul(h1 ^ ch, 2654435761);
      h2 = Math.imul(h2 ^ ch, 1597334677);
    }
    h1 = Math.imul(h1 ^ (h1>>>16), 2246822507) ^ Math.imul(h2 ^ (h2>>>13), 3266489909);
    h2 = Math.imul(h2 ^ (h2>>>16), 2246822507) ^ Math.imul(h1 ^ (h1>>>13), 3266489909);
    let hash = 4294967296 * (2097151 & h2) + (h1>>>0);
    return hash / Number.MAX_SAFE_INTEGER;
  };

</script>

<div class="container rounded p-2 m-2" style="border-left: solid 4px {getColor(repo.html_url)}">
  <div>
    {#if !star}
      <a href="#" on:click|preventDefault="{onSponsor}">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 573.655 550.909" color="gold" overflow="visible">
          <path
            d="M258.128 36.96l-65.3 132.4-146.1 21.3c-26.2 3.8-36.7 36.1-17.7 54.6l105.7 103-25 145.5c-4.5 26.3 23.2 46 46.4 33.7l130.7-68.7 130.7 68.7c23.2 12.2 50.9-7.4 46.4-33.7l-25-145.5 105.7-103c19-18.5 8.5-50.8-17.7-54.6l-146.1-21.3-65.3-132.4c-11.7-23.6-45.6-23.9-57.4 0z"
            fill="none" stroke="gold" stroke-width="40" />
        </svg>
      </a>
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 573.655 550.909" color="gold" overflow="visible">
        <path
          d="M258.128 36.96l-65.3 132.4-146.1 21.3c-26.2 3.8-36.7 36.1-17.7 54.6l105.7 103-25 145.5c-4.5 26.3 23.2 46 46.4 33.7l130.7-68.7 130.7 68.7c23.2 12.2 50.9-7.4 46.4-33.7l-25-145.5 105.7-103c19-18.5 8.5-50.8-17.7-54.6l-146.1-21.3-65.3-132.4c-11.7-23.6-45.6-23.9-57.4 0z"
          fill="gold" stroke="gold" stroke-width="40" />
      </svg>
    {/if}
  </div>
  <div class="">
    <div><b>{repo.full_name}</b></div>
    <div>{repo.description}</div>
    <div class="url"><a href="{repo.html_url}">{repo.html_url}</a></div>
  </div>


</div>
