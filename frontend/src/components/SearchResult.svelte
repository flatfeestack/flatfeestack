<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import {
    user,
    error,
    sponsoredRepos,
    trustedRepos,
    multiplierSponsoredRepos,
    multiplierCountByRepo,
  } from "../ts/mainStore";
  import { getColor1 } from "../ts/utils";
  import type { Repo } from "../types/backend";
  import { faClose } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import { get } from "svelte/store";

  export let repo: Repo;
  let repoIsHealthy = false;
  let star = false;
  let multiplier = false;

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

  const onMultiplier = async () => {
    try {
      if (!star) {
        const resSponsor = await API.repos.tag(repo.uuid);
        $sponsoredRepos = [...$sponsoredRepos, resSponsor];
        star = true;
      }
      const resMultiplier = await API.repos.setMultiplier(repo.uuid);
      $multiplierSponsoredRepos = [...$multiplierSponsoredRepos, resMultiplier];
      multiplier = true;

      multiplierCountByRepo.update((counts) => {
        counts[repo.uuid] = (counts[repo.uuid] || 0) + 1;
        return counts;
      });
    } catch (e) {
      $error = e;
      star = false;
      multiplier = false;
    }
  };

  $: {
    const tmpStar = $sponsoredRepos.find((r: Repo) => r.uuid === repo.uuid);
    star = tmpStar !== undefined;

    const tmpHealth = $trustedRepos.find((r: Repo) => r.uuid === repo.uuid);
    repoIsHealthy = tmpHealth !== undefined;

    const tmpMultiplier = $multiplierSponsoredRepos.find(
      (r: Repo) => r.uuid === repo.uuid
    );
    multiplier = tmpMultiplier !== undefined;
  }

  onMount(async () => {
    try {
      const initialCount = await API.repos.getMultiplierCount(repo.uuid);

      multiplierCountByRepo.update((counts) => {
        return {
          ...counts,
          [repo.uuid]: initialCount,
        };
      });
    } catch (e) {
      console.error("Failed to fetch initial multiplier count:", e);
    }
  });
</script>

<style>
  .container {
    margin-bottom: 2em;
    max-width: 40rem;
  }

  #sponsor-buttons-div {
    margin-left: auto;
  }

  a.icons-a {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    margin: 0.25rem;
  }

  svg,
  #multiplier-sponsor-amount-div {
    padding: 0.25rem;
    height: 1.5rem;
    width: 1.5rem;
  }
  svg#multiplier-amount-coin {
    padding: 0;
    height: 1rem;
    width: 1rem;
  }
  .url {
    font-size: small;
  }
  .title {
    font-weight: bold;
  }
  div.icons-div {
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 0.25rem;
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
  <div class="container-col2 mx-2">
    {#if !star}
      <a class="icons-a" href={"#"} on:click|preventDefault={onSponsor}>
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
      <div class="icons-div">
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
      </div>
    {/if}

    {#if $user.multiplier}
      {#if !multiplier}
        <a class="icons-a" href={"#"} on:click|preventDefault={onMultiplier}>
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
                <stop offset="0%" style="stop-color:#D3D3D3; stop-opacity:1" />
                <stop offset="50%" style="stop-color:#A9A9A9; stop-opacity:1" />
                <stop
                  offset="100%"
                  style="stop-color:#696969; stop-opacity:1"
                />
              </radialGradient>
            </defs>
            <circle cx="10" cy="10" r="10" fill="url(#greyGradient)" />
            <path
              fill="#404040"
              d="M9 13v-2a3 3 0 1 1 0-6V4a1 1 0 1 1 2 0v1h.022A2.978 2.978 0 0 1 14 7.978a1 1 0 0 1-2 0A.978.978 0 0 0 11.022 7H11v2a3 3 0 0 1 0 6v1a1 1 0 0 1-2 0v-1h-.051A2.949 2.949 0 0 1 6 12.051a1 1 0 1 1 2 0 .95.95 0 0 0 .949.949H9zm2 0a1 1 0 0 0 0-2v2zM9 7a1 1 0 1 0 0 2V7zm1 13C4.477 20 0 15.523 0 10S4.477 0 10 0s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 1 0 0-16 8 8 0 0 0 0 16z"
            />
          </svg>
        </a>
      {:else}
        <div class="icons-div">
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
                <stop offset="0%" style="stop-color:#98FB98; stop-opacity:1" />
                <stop offset="50%" style="stop-color:#32CD32; stop-opacity:1" />
                <stop
                  offset="100%"
                  style="stop-color:#006400; stop-opacity:1"
                />
              </radialGradient>
            </defs>
            <circle cx="10" cy="10" r="10" fill="url(#greenGradient)" />
            <path
              fill="#004d00"
              d="M9 13v-2a3 3 0 1 1 0-6V4a1 1 0 1 1 2 0v1h.022A2.978 2.978 0 0 1 14 7.978a1 1 0 0 1-2 0A.978.978 0 0 0 11.022 7H11v2a3 3 0 0 1 0 6v1a1 1 0 0 1-2 0v-1h-.051A2.949 2.949 0 0 1 6 12.051a1 1 0 1 1 2 0 .95.95 0 0 0 .949.949H9zm2 0a1 1 0 0 0 0-2v2zM9 7a1 1 0 1 0 0 2V7zm1 13C4.477 20 0 15.523 0 10S4.477 0 10 0s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 1 0 0-16 8 8 0 0 0 0 16z"
            />
          </svg>
        </div>
      {/if}
    {/if}
  </div>

  <div>
    <div class="title">{repo.name}</div>
    <div class="desc">{repo.description}</div>
    <div class="url"><a href={repo.url}>{repo.url}</a></div>
  </div>

  <div class="container-col2 mx-2" id="sponsor-buttons-div">
    <div class="icons-div">
      {#if !repoIsHealthy}
        <svg
          viewBox="2 2 20 20"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          overflow="visible"
        >
          <path
            d="M8.38 12L10.79 14.42L15.62 9.57996"
            stroke="#169df0"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <path
            d="M10.75 2.44995C11.44 1.85995 12.57 1.85995 13.27 2.44995L14.85 3.80995C15.15 4.06995 15.71 4.27995 16.11 4.27995H17.81C18.87 4.27995 19.74 5.14995 19.74 6.20995V7.90995C19.74 8.29995 19.95 8.86995 20.21 9.16995L21.57 10.7499C22.16 11.4399 22.16 12.5699 21.57 13.2699L20.21 14.8499C19.95 15.1499 19.74 15.7099 19.74 16.1099V17.8099C19.74 18.8699 18.87 19.7399 17.81 19.7399H16.11C15.72 19.7399 15.15 19.9499 14.85 20.2099L13.27 21.5699C12.58 22.1599 11.45 22.1599 10.75 21.5699L9.17 20.2099C8.87 19.9499 8.31 19.7399 7.91 19.7399H6.18C5.12 19.7399 4.25 18.8699 4.25 17.8099V16.0999C4.25 15.7099 4.04 15.1499 3.79 14.8499L2.44 13.2599C1.86 12.5699 1.86 11.4499 2.44 10.7599L3.79 9.16995C4.04 8.86995 4.25 8.30995 4.25 7.91995V6.19995C4.25 5.13995 5.12 4.26995 6.18 4.26995H7.91C8.3 4.26995 8.87 4.05995 9.17 3.79995L10.75 2.44995Z"
            stroke="#169df0"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      {:else}
        <svg
          viewBox="2 2 20 20"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          overflow="visible"
        >
          <path
            d="M21.5609 10.7386L20.2009 9.15859C19.9409 8.85859 19.7309 8.29859 19.7309 7.89859V6.19859C19.7309 5.13859 18.8609 4.26859 17.8009 4.26859H16.1009C15.7109 4.26859 15.1409 4.05859 14.8409 3.79859L13.2609 2.43859C12.5709 1.84859 11.4409 1.84859 10.7409 2.43859L9.17086 3.80859C8.87086 4.05859 8.30086 4.26859 7.91086 4.26859H6.18086C5.12086 4.26859 4.25086 5.13859 4.25086 6.19859V7.90859C4.25086 8.29859 4.04086 8.85859 3.79086 9.15859L2.44086 10.7486C1.86086 11.4386 1.86086 12.5586 2.44086 13.2486L3.79086 14.8386C4.04086 15.1386 4.25086 15.6986 4.25086 16.0886V17.7986C4.25086 18.8586 5.12086 19.7286 6.18086 19.7286H7.91086C8.30086 19.7286 8.87086 19.9386 9.17086 20.1986L10.7509 21.5586C11.4409 22.1486 12.5709 22.1486 13.2709 21.5586L14.8509 20.1986C15.1509 19.9386 15.7109 19.7286 16.1109 19.7286H17.8109C18.8709 19.7286 19.7409 18.8586 19.7409 17.7986V16.0986C19.7409 15.7086 19.9509 15.1386 20.2109 14.8386L21.5709 13.2586C22.1509 12.5686 22.1509 11.4286 21.5609 10.7386ZM16.1609 10.1086L11.3309 14.9386C11.1909 15.0786 11.0009 15.1586 10.8009 15.1586C10.6009 15.1586 10.4109 15.0786 10.2709 14.9386L7.85086 12.5186C7.56086 12.2286 7.56086 11.7486 7.85086 11.4586C8.14086 11.1686 8.62086 11.1686 8.91086 11.4586L10.8009 13.3486L15.1009 9.04859C15.3909 8.75859 15.8709 8.75859 16.1609 9.04859C16.4509 9.33859 16.4509 9.81859 16.1609 10.1086Z"
            fill="#169df0"
          />
        </svg>
      {/if}
    </div>

    <div class="container-small icons-div" id="multiplier-sponsor-amount-div">
      {$multiplierCountByRepo[repo.uuid]}
      <Fa icon={faClose} />
    </div>
  </div>
</div>
