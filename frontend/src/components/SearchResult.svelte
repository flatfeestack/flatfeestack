<script lang="ts">
  import { API } from "../ts/api";
  import { error, sponsoredRepos } from "../ts/mainStore";
  import { getColor1 } from "../ts/utils";
  import type { Repo } from "../types/backend";

  export let repo: Repo;
  let star = false;

  const onSponsor = async () => {
    try {
      const res = await API.repos.tag(repo.uuid);
      $sponsoredRepos = [...$sponsoredRepos, res];
      star = true;
    } catch (e) {
      $error = e;
      star = false;
    }
  };

  $: {
    const tmp = $sponsoredRepos.find((r: Repo) => {
      return r.uuid === repo.uuid;
    });
    star = tmp !== undefined;
  }
</script>

<style>
  .container {
    display: flex;
    flex-direction: row;
    margin-bottom: 2em;
    max-width: 40rem;
  }

  svg {
    padding: 0.25em;
    height: 1em;
    width: 1em;
  }
  .url {
    font-size: small;
  }
  .title {
    font-weight: bold;
  }
  @media screen and (max-width: 600px) {
    .container {
      margin: 1rem 0.5rem;
    }
    .title,
    .desc,
    .url {
      word-break: break-word;
    }
  }
</style>

<div
  class="container rounded p-2 m-2"
  style="border-left: solid 6px {getColor1(repo.uuid)}"
>
  <div>
    {#if !star}
      <a href={"#"} on:click|preventDefault={onSponsor}>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 573.655 550.909"
          color="gold"
          overflow="visible"
        >
          <path
            d="M258.128 36.96l-65.3 132.4-146.1 21.3c-26.2 3.8-36.7 36.1-17.7 54.6l105.7 103-25 145.5c-4.5 26.3 23.2 46 46.4 33.7l130.7-68.7 130.7 68.7c23.2 12.2 50.9-7.4 46.4-33.7l-25-145.5 105.7-103c19-18.5 8.5-50.8-17.7-54.6l-146.1-21.3-65.3-132.4c-11.7-23.6-45.6-23.9-57.4 0z"
            fill="none"
            stroke="gold"
            stroke-width="40"
          />
        </svg>
      </a>
    {:else}
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 573.655 550.909"
        color="gold"
        overflow="visible"
      >
        <path
          d="M258.128 36.96l-65.3 132.4-146.1 21.3c-26.2 3.8-36.7 36.1-17.7 54.6l105.7 103-25 145.5c-4.5 26.3 23.2 46 46.4 33.7l130.7-68.7 130.7 68.7c23.2 12.2 50.9-7.4 46.4-33.7l-25-145.5 105.7-103c19-18.5 8.5-50.8-17.7-54.6l-146.1-21.3-65.3-132.4c-11.7-23.6-45.6-23.9-57.4 0z"
          fill="gold"
          stroke="gold"
          stroke-width="40"
        />
      </svg>
    {/if}
  </div>
  <div>
    <div class="title">{repo.name}</div>
    <div class="desc">{repo.description}</div>
    <div class="url"><a href={repo.url}>{repo.url}</a></div>
  </div>
</div>
