<script lang="ts">
  import { API } from "../ts/api";
  import {
    error,
    trustedRepos,
    reposToUnTrustAfterTimeout,
    reposInSearchResult,
  } from "../ts/mainStore";
  import { getColor1, getColor2 } from "../ts/utils";
  import type { Repo } from "../types/backend";
  import { onDestroy } from "svelte";
  import RepoAssessmentOverlay from "./RepoAssessmentOverlay.svelte";

  export let repo: Repo;
  let verifiedStar = true;
  let unTrustProgress = 100;
  let unTrustOnDestroy = false;
  let showUnTrustProgressBar = false;
  let intervalId: NodeJS.Timer | null = null;
  let repoHealthValue: Number | "NA";
  const undoDuration: number = 5000;
  let assessmentOverlayVisible: boolean = false;

  $: {
    const tmp = $trustedRepos.find((r: Repo) => {
      return r.uuid === repo.uuid;
    });
    repoHealthValue = tmp ? tmp.healthValue : "NA";
    const repoIsTrusted = tmp !== undefined;

    showUnTrustProgressBar = $reposToUnTrustAfterTimeout.some(
      (r) => r.uuid === repo.uuid
    );

    verifiedStar = repoIsTrusted && !showUnTrustProgressBar;

    if (showUnTrustProgressBar) {
      startProgressBar();
    }
  }

  async function unTrust() {
    unTrustProgress = 100;
    try {
      await API.repos.untrust(repo.uuid);
      $trustedRepos = $trustedRepos.filter((r: Repo) => {
        return r.uuid !== repo.uuid;
      });
    } catch (e) {
      $error = e;
    } finally {
      reposToUnTrustAfterTimeout.update((list) =>
        list.filter((r) => r !== repo)
      );
    }
  }

  function delayUnTrust() {
    reposToUnTrustAfterTimeout.update((list) => [...list, repo]);

    setTimeout(async () => {
      if (
        $reposToUnTrustAfterTimeout.some((r) => r.uuid === repo.uuid) &&
        !unTrustOnDestroy &&
        unTrustProgress <= 2
      ) {
        await unTrust();
      }
    }, undoDuration);
  }

  function trustRepo() {
    if (!verifiedStar) {
      reposToUnTrustAfterTimeout.update((list) =>
        list.filter((r) => r.uuid !== repo.uuid)
      );
    }
  }

  function startProgressBar(): void {
    if (intervalId) return;

    const interval = 100;
    const decrement = (100 / undoDuration) * interval;

    intervalId = setInterval(() => {
      if (!$reposToUnTrustAfterTimeout.some((r) => r.uuid === repo.uuid)) {
        clearProgressBar();
        return;
      }
      unTrustProgress -= decrement;
      if (unTrustProgress <= 0) {
        clearProgressBar();
        unTrustProgress = 0;
      }
    }, interval);
  }

  function clearProgressBar(): void {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }
    unTrustProgress = 100;
  }

  function showOverlay() {
    assessmentOverlayVisible = true;
  }

  function hideOverlay() {
    assessmentOverlayVisible = false;
  }

  onDestroy(async () => {
    // setTimeout(async () => {
    //
    // }, 100)
    if (
      $reposToUnTrustAfterTimeout.some((r) => r.uuid === repo.uuid) &&
      !$reposInSearchResult.some((r) => r.uuid === repo.uuid)
    ) {
      await unTrust();
    }
    unTrustOnDestroy = true;
  });
</script>

<style>
  .child {
    flex: 1 0;
    margin: 0.5em;
    max-width: 7em;
    min-width: 7em;
    box-shadow: 0.25em 0.25em 0.25em #e1e1e3;
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
    position: relative;
    overflow: hidden;
  }
  .color {
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
    height: 2.5em;
    box-shadow: 0 3px 2px -2px black;
  }
  .repo-title {
    font-weight: bold;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: center;
    display: block;
  }
  .progress-bar {
    position: absolute;
    top: 0;
    left: 0;
    height: 0.25rem;
    background-color: #169df0;
    width: 100%;
    transition: width 0.1s linear;
    z-index: 1;
  }
  .body {
    text-align: center;
    font-size: medium;
  }
  .url {
    text-decoration: none;
    color: black;
  }
  svg {
    padding: 0.25em;
    height: 2em;
    width: 2em;
  }

  button.square-1 {
    margin: 0.25em;
    height: 2.5em;
    min-width: 2.5em;
    width: auto;
    color: black;
  }

  #trust-value-button {
    display: flex;
    justify-content: center;
    align-items: center;
    text-align: center;
    border: 0.15em solid #494949;
    border-radius: 0.5em;
    padding: 0.2em;
    font-size: 0.9rem;
    font-weight: bold;
    transition: background-color 0.1s linear, filter 0.1s linear;
  }

  #trust-value-button:hover {
    background-color: var(--primary-100);
    filter: drop-shadow(2px 2px 2px rgba(0, 0, 0, 0.7));
  }

  .color :global(a:hover) {
    filter: drop-shadow(2px 2px 2px rgba(0, 0, 0, 0.7));
  }

  .assessment-overlay {
    position: fixed;
    display: block;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: rgba(
      221,
      221,
      221,
      0.3
    ); /* secondary-200 with 30% opacity */
    z-index: 2;
    cursor: pointer;
  }

  .overlay-container {
    position: absolute;
    width: 60vw;
    height: 90vh;
    background-color: white;
    color: black;
    overflow-y: auto;
    margin: 5vh 20vw;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    border-radius: 10px;
    cursor: pointer;
  }

  @media screen and (min-width: 2000px) {
    .overlay-container {
      width: 1185px;
      margin: 5vh calc((100vw - 1185px) / 2);
    }
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
    class="progress-bar"
    style="width: {unTrustProgress}%; visibility: {showUnTrustProgressBar
      ? 'visible'
      : 'hidden'}"
  />
  <div
    class="color container-small justify-between"
    style="background-image:radial-gradient(circle at top right,{getColor2(
      repo.uuid
    )},{getColor1(repo.uuid)});"
  >
    <div class="container-small">
      {#if verifiedStar}
        <a
          class="container-small"
          href={"#"}
          on:click|preventDefault={delayUnTrust}
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M21.5609 10.7386L20.2009 9.15859C19.9409 8.85859 19.7309 8.29859 19.7309 7.89859V6.19859C19.7309 5.13859 18.8609 4.26859 17.8009 4.26859H16.1009C15.7109 4.26859 15.1409 4.05859 14.8409 3.79859L13.2609 2.43859C12.5709 1.84859 11.4409 1.84859 10.7409 2.43859L9.17086 3.80859C8.87086 4.05859 8.30086 4.26859 7.91086 4.26859H6.18086C5.12086 4.26859 4.25086 5.13859 4.25086 6.19859V7.90859C4.25086 8.29859 4.04086 8.85859 3.79086 9.15859L2.44086 10.7486C1.86086 11.4386 1.86086 12.5586 2.44086 13.2486L3.79086 14.8386C4.04086 15.1386 4.25086 15.6986 4.25086 16.0886V17.7986C4.25086 18.8586 5.12086 19.7286 6.18086 19.7286H7.91086C8.30086 19.7286 8.87086 19.9386 9.17086 20.1986L10.7509 21.5586C11.4409 22.1486 12.5709 22.1486 13.2709 21.5586L14.8509 20.1986C15.1509 19.9386 15.7109 19.7286 16.1109 19.7286H17.8109C18.8709 19.7286 19.7409 18.8586 19.7409 17.7986V16.0986C19.7409 15.7086 19.9509 15.1386 20.2109 14.8386L21.5709 13.2586C22.1509 12.5686 22.1509 11.4286 21.5609 10.7386ZM16.1609 10.1086L11.3309 14.9386C11.1909 15.0786 11.0009 15.1586 10.8009 15.1586C10.6009 15.1586 10.4109 15.0786 10.2709 14.9386L7.85086 12.5186C7.56086 12.2286 7.56086 11.7486 7.85086 11.4586C8.14086 11.1686 8.62086 11.1686 8.91086 11.4586L10.8009 13.3486L15.1009 9.04859C15.3909 8.75859 15.8709 8.75859 16.1609 9.04859C16.4509 9.33859 16.4509 9.81859 16.1609 10.1086Z"
              fill="#169df0"
            />
          </svg>
        </a>
      {:else}
        <a
          class="container-small"
          href={"#"}
          on:click|preventDefault={trustRepo}
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
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
        </a>
      {/if}
    </div>
    <button
      id="trust-value-button"
      class="square-1 tooltip"
      on:click={showOverlay}
    >
      {repoHealthValue}
      <span class="tooltiptext">Show Repo Assessment</span>
    </button>
  </div>
  {#if repo}
    <div title={repo.name}>
      <a
        href={repo.url}
        class="repo-title py-2 url"
        target="_blank"
        rel="noreferrer"
      >
        {repo.name}
      </a>
    </div>
  {/if}
</div>

<div
  class:assessment-overlay={assessmentOverlayVisible}
  class:hidden={!assessmentOverlayVisible}
  role="button"
  tabindex="0"
  on:click={hideOverlay}
  on:keydown={(e) => (e.key === "Enter" || e.key === " ") && hideOverlay()}
>
  <div class="overlay-container">
    {#if assessmentOverlayVisible}
      <RepoAssessmentOverlay {repo} />
    {/if}
  </div>
</div>
