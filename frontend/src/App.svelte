<script lang="ts">
  import { links, Route, Router } from "svelte-routing";
  import {globalHistory} from 'svelte-routing/src/history';
  import { user, route, loginFailed } from "./ts/store";
  import { removeSession } from "./ts/services";
  import { onMount } from "svelte";
  import { API } from "./ts/api";

  import Landing from "./routes/Landing.svelte";
  import Badges from "./routes/Badges.svelte";
  import Login from "./routes/Login.svelte";
  import Signup from "./routes/Signup.svelte";
  import Forgot from "./routes/Forgot.svelte";
  import ConfirmForgot from "./routes/ConfirmForgot.svelte";
  import ConfirmSignup from "./routes/ConfirmSignup.svelte";
  import Search from "./routes/Search.svelte";
  import CatchAll from "./routes/CatchAllRoute.svelte";
  import Income from "./routes/Income.svelte";
  import Profile from "./routes/Profile.svelte";
  import Admin from "./routes/Admin.svelte";
  import Spinner from "./components/Spinner.svelte";
  import ConfirmGitEmail from "./routes/ConfirmGitEmail.svelte";
  import ConfirmInvite from "./routes/ConfirmInvite.svelte";

  export let url;
  let loading = true;

  onMount(async () => {
    try {
      loading = true;
      $user = await API.user.get();
    } catch (e) {
      $loginFailed = true;
    } finally {
      loading = false;
    }
  });

  //https://github.com/EmilTholin/svelte-routing/issues/41
  //https://github.com/EmilTholin/svelte-routing/issues/62
  $route = globalHistory.location;
  globalHistory.listen(history => {
    $route = history.location;
  });

  let pathname="/";
  if (typeof window !== "undefined") {
    pathname = window.location.pathname;
  }

</script>

<style>
    .main {
        display: flex;
        flex-direction: column;
        min-height: 100%;
    }

    header {
        padding: 1em;
        background-color: #000;
        box-shadow: 0 0 5px rgba(0, 0, 0, 0.75);
        justify-content: space-between;
        flex: 0 1 auto;
    }

    header, nav {
        display: flex;
        align-items: center;
    }

    main {
        flex: 1 1 auto;
        display: flex;
    }

    footer {
        background-color: #000;
        color: white;
        height: 100%;
        flex: 0 1 auto;
    }

    header :global(a), .header :global(a:visited), .header :global(a:active) {
        text-decoration: none;
        color: #ffffff;
    }

    img {
        padding-right: 0.25em;
    }

    .main-nav :global(a), .main-nav :global(a:visited) {
        padding: 0.5em 1em 0.5em 1em;
    }

    .main-nav :global(a:hover) {
        color: var(--primary-500);
        transition: color .15s;
    }

    .signup :global(a), .signup :global(a:visited) {
        border: white 1px solid;
        border-radius: 3px;
    }

    .signup :global(a:hover) {
        border: var(--primary-500) 1px solid;
        transition: border .15s;
    }
</style>

<div class="main">
  <header use:links>
    <a href="/">
      <img src="/assets/images/new-logo-4.svg" alt="Flatfeestack" />
      <img class="hide-sx" src="/assets/images/logo-text-w.svg" alt="Flatfeestack" />
    </a>
    <nav>
      {#if $user.id}
        <span class="text-primary-500">{$user.email}</span>
        <div class="main-nav"><a href="/login" on:click={removeSession}>Sign out</a></div>
      {:else}
        <div class="main-nav"><a href="/login">Login</a></div>
        <div class="main-nav signup"><a href="/signup">Sign up</a></div>
      {/if}
    </nav>
  </header>

  <main>
    <Router url="{url}">

      {#if loading}
        <Spinner />
      {:else}
        <Route path="/" component="{Landing}" />

        <Route path="/login" component="{Login}" />
        <Route path="/signup" component="{Signup}" />
        <Route path="/forgot" component="{Forgot}" />

        <Route path="/confirm/reset/:email/:token" component="{ConfirmForgot}" />
        <Route path="/confirm/signup/:email/:token" component="{ConfirmSignup}" />
        <Route path="/confirm/git-email/:email/:token" component="{ConfirmGitEmail}" />
        <Route path="/confirm/invite/:email/:emailToken/:inviteEmail/:inviteDate/:inviteToken" component="{ConfirmInvite}" />

        {#if $user.id && ($route.pathname.startsWith("/dashboard") || pathname.startsWith("/dashboard"))}
          <Route path="/dashboard/search" component="{Search}" />
          <Route path="/dashboard/income" component="{Income}" />
          <Route path="/dashboard/profile" component="{Profile}" />
          <Route path="/dashboard/badges" component="{Badges}" />
          <Route path="/dashboard/admin" component="{Admin}" />
        {/if}
        <Route path="*" component="{CatchAll}" />
      {/if}

    </Router>
  </main>

  <footer>© flatfeestack.io</footer>
</div>
