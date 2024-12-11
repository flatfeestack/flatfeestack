<script lang="ts">
  import { API } from "../ts/api";
  import {
    user,
    error,
    sponsoredRepos,
    multiplierSponsoredRepos,
    multiplierCountByRepo,
  } from "../ts/mainStore";
  import { getColor1, getColor2 } from "../ts/utils";
  import type { Repo } from "../types/backend";

  export let repo: Repo;
  let star = true;
  let multiplier: boolean;

  $: {
    const tmpMultiplier = $multiplierSponsoredRepos.find(
      (r: Repo) => r.uuid === repo.uuid
    );
    multiplier = tmpMultiplier !== undefined;
  }

  async function unsetMultiplierHelper() {
    await API.repos.unsetMultiplier(repo.uuid);
    $multiplierSponsoredRepos = $multiplierSponsoredRepos.filter((r: Repo) => {
      return r.uuid !== repo.uuid;
    });

    multiplierCountByRepo.update((counts) => {
      if (counts[repo.uuid] > 1) {
        counts[repo.uuid] -= 1;
      } else {
        delete counts[repo.uuid];
      }
      return counts;
    });
  }

  async function unTag() {
    try {
      await API.repos.untag(repo.uuid);
      $sponsoredRepos = $sponsoredRepos.filter((r: Repo) => {
        return r.uuid !== repo.uuid;
      });
      if (multiplier) {
        await unsetMultiplierHelper();
        multiplier = false;
      }
      star = false;
    } catch (e) {
      $error = e;
    }
  }

  async function unsetMultiplier() {
    multiplier = false;
    try {
      await unsetMultiplierHelper();
    } catch (e) {
      $error = e;
    }
  }

  async function setMultiplier() {
    try {
      const resMultiplier = await API.repos.setMultiplier(repo.uuid);
      $multiplierSponsoredRepos = [...$multiplierSponsoredRepos, resMultiplier];
      multiplier = true;

      multiplierCountByRepo.update((counts) => {
        counts[repo.uuid] = (counts[repo.uuid] || 0) + 1;
        return counts;
      });
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
    box-shadow: 0.25em 0.25em 0.25em #e1e1e3;
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
  }
  .color {
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
    height: 3.5em;
    box-shadow: 0 3px 2px -2px black;
  }
  .center2 {
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
    filter: drop-shadow(2px 2px 2px rgba(0, 0, 0, 0.7));
  }

  @media screen and (max-width: 600px) {
    .child {
      max-width: unset;
      min-width: unset;
      width: 100%;
      margin: 0.5em 0;
    }
  }
</style>

<div class="child rounded">
  <div
    class="color"
    style="background-image:radial-gradient(circle at top right,{getColor2(
      repo.uuid
    )},{getColor1(repo.uuid)});"
  >
    <div>
      {#if star}
        <a href={"#"} on:click|preventDefault={unTag}>
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
            fill="none"
            stroke="gold"
            stroke-width="40"
          />
        </svg>
      {/if}
      {#if $user.multiplier}
        {#if multiplier}
          <a href={"#"} on:click|preventDefault={unsetMultiplier}>
            <svg
              viewBox="0 0 20 20"
              xmlns="http://www.w3.org/2000/svg"
              preserveAspectRatio="xMinYMin"
              overflow="visible"
              class="jam jam-coin"
            >
              <defs>
                <radialGradient
                  id="greenGradient"
                  cx="50%"
                  cy="50%"
                  r="50%"
                  fx="50%"
                  fy="50%"
                >
                  <stop
                    offset="0%"
                    style="stop-color:#98FB98; stop-opacity:1"
                  />
                  <!-- Light green -->
                  <stop
                    offset="50%"
                    style="stop-color:#32CD32; stop-opacity:1"
                  />
                  <!-- Medium green -->
                  <stop
                    offset="100%"
                    style="stop-color:#006400; stop-opacity:1"
                  />
                  <!-- Dark green -->
                </radialGradient>
              </defs>
              <circle cx="10" cy="10" r="10" fill="url(#greenGradient)" />
              <path
                fill="#004d00"
                d="M9 13v-2a3 3 0 1 1 0-6V4a1 1 0 1 1 2 0v1h.022A2.978 2.978 0 0 1 14 7.978a1 1 0 0 1-2 0A.978.978 0 0 0 11.022 7H11v2a3 3 0 0 1 0 6v1a1 1 0 0 1-2 0v-1h-.051A2.949 2.949 0 0 1 6 12.051a1 1 0 1 1 2 0 .95.95 0 0 0 .949.949H9zm2 0a1 1 0 0 0 0-2v2zM9 7a1 1 0 1 0 0 2V7zm1 13C4.477 20 0 15.523 0 10S4.477 0 10 0s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 1 0 0-16 8 8 0 0 0 0 16z"
              />
            </svg>
          </a>
        {:else}
          <a href={"#"} on:click|preventDefault={setMultiplier}>
            <svg
              viewBox="0 0 20 20"
              xmlns="http://www.w3.org/2000/svg"
              preserveAspectRatio="xMinYMin"
              overflow="visible"
              class="jam jam-coin"
            >
              <defs>
                <radialGradient
                  id="greyGradient"
                  cx="50%"
                  cy="50%"
                  r="50%"
                  fx="50%"
                  fy="50%"
                >
                  <stop
                    offset="0%"
                    style="stop-color:#D3D3D3; stop-opacity:1"
                  />
                  <!-- Light grey -->
                  <stop
                    offset="50%"
                    style="stop-color:#A9A9A9; stop-opacity:1"
                  />
                  <!-- Medium grey -->
                  <stop
                    offset="100%"
                    style="stop-color:#696969; stop-opacity:1"
                  />
                  <!-- Dark grey -->
                </radialGradient>
              </defs>
              <circle cx="10" cy="10" r="10" fill="url(#greyGradient)" />
              <path
                fill="#404040"
                d="M9 13v-2a3 3 0 1 1 0-6V4a1 1 0 1 1 2 0v1h.022A2.978 2.978 0 0 1 14 7.978a1 1 0 0 1-2 0A.978.978 0 0 0 11.022 7H11v2a3 3 0 0 1 0 6v1a1 1 0 0 1-2 0v-1h-.051A2.949 2.949 0 0 1 6 12.051a1 1 0 1 1 2 0 .95.95 0 0 0 .949.949H9zm2 0a1 1 0 0 0 0-2v2zM9 7a1 1 0 1 0 0 2V7zm1 13C4.477 20 0 15.523 0 10S4.477 0 10 0s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 1 0 0-16 8 8 0 0 0 0 16z"
              />
            </svg>
          </a>
        {/if}
      {/if}
    </div>
  </div>
  {#if repo}
    <div class="center center2 py-2">{repo.name}</div>
    <div class="body">{repo.description}</div>
    <div>
      <a href={repo.url} class="py-2 url" target="_blank" rel="noreferrer">
        {repo.url}
      </a>
    </div>
  {/if}
</div>
