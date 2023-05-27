<script lang="ts">
  import { navigate, link } from "svelte-routing";
  import { user, loginFailed, error, token } from "../ts/mainStore";
  import { hasAccessToken, removeSession } from "../ts/services";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { faHome } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  let loading = true;
  let auth = false;

  function logout() {
    removeSession();
    navigate("/login");
  }

  $: {
    if ($token) {
      auth = true;
    }
  }

  onMount(async () => {
    const authCookie = document.cookie
      .split("; ")
      .find((row) => row.startsWith("auth="));
    if (authCookie || $token || hasAccessToken()) {
      auth = true;
    }
    try {
      loading = true;
      $user = await API.user.get();
    } catch (e) {
      $loginFailed = true;
    } finally {
      loading = false;
    }
  });
</script>

<style>
  header {
    padding: 1em;
    background-color: #fff;
    border-bottom: 1px #000 solid;
    justify-content: space-between;
    flex: 0 0 auto;
  }

  header,
  nav {
    display: flex;
    align-items: center;
    font-size: 1.1rem;
  }

  header :global(a),
  header :global(a:visited),
  header :global(a:active) {
    text-decoration: none;
    color: #000;
  }

  .close {
    cursor: pointer;
    text-align: right;
  }

  .err-container {
    position: fixed;
    width: 100%;
    display: flex;
    flex-direction: row;
  }
  .err-container button {
    margin-right: 30px;
  }

  .imgSmallLogo {
    padding-right: 0.25em;
    width: 3rem;
  }
  .imgNormalLogo {
    padding-right: 0.25em;
    width: 10rem;
  }
</style>

<header>
  <a href="/" use:link>
    <img
      class="hide-mda imgSmallLogo"
      src="/images/favicon.svg"
      alt="FlatFeeStack"
    />
    <img
      class="hide-sx imgNormalLogo"
      src="/images/ffs-logo.svg"
      alt="FlatFeeStack"
    />
  </a>
  <nav>
    {#if $user.id}
      <a href="/user/search" use:link>
        <Fa icon={faHome} size="sm" class="icon" />
      </a>
      {#if $user.image}
        <img class="image-org-sx" src={$user.image} alt="user profile img" />
      {/if}

      {$user.email}
      <form on:submit|preventDefault={logout}>
        <button class="button3 center mx-2" type="submit">Sign out</button>
      </form>
    {:else}
      <form on:submit|preventDefault={() => navigate("/login")}>
        <button class="button3 center mx-2" type="submit">Login</button>
      </form>
      <form on:submit|preventDefault={() => navigate("/signup")}>
        <button class="button1 center mx-2" type="submit">Sign Up</button>
      </form>
    {/if}
    <button class="button4 center mx-2" on:click={() => navigate("/dao/home")}
      >DAO</button
    >
  </nav>
</header>

{#if $error}
  <div class="bg-red p-2 parent err-container">
    <div class="w-100">{@html $error}</div>
    <div>
      <button
        class="close accessible-btn"
        on:click|preventDefault={() => {
          $error = null;
        }}
      >
        ✕
      </button>
    </div>
  </div>
{/if}
