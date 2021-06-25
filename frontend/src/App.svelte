<script lang="ts">
  import { links, Route, Router } from "svelte-routing";
  import {globalHistory} from 'svelte-routing/src/history';
  import { user, route, loginFailed, error } from "./ts/store";
  import { removeSession } from "./ts/services";
  import { onMount } from "svelte";
  import { API } from "./ts/api";
  import { faHome } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  import Landing from "./routes/Landing.svelte";
  import Badges from "./routes/Badges.svelte";
  import PublicBadges from "./routes/PublicBadges.svelte";
  import Login from "./routes/Login.svelte";
  import Signup from "./routes/Signup.svelte";
  import Forgot from "./routes/Forgot.svelte";
  import ConfirmForgot from "./routes/ConfirmForgot.svelte";
  import ConfirmSignup from "./routes/ConfirmSignup.svelte";
  import Search from "./routes/Search.svelte";
  import CatchAll from "./routes/CatchAllRoute.svelte";
  import Income from "./routes/Income.svelte";
  import Payments from "./routes/Payments.svelte";
  import Admin from "./routes/Admin.svelte";
  import Spinner from "./components/Spinner.svelte";
  import ForwardGitEmail from "./routes/ForwardGitEmail.svelte";
  import ForwardInvite from "./routes/ForwardInvite.svelte";
  import Settings from "./routes/Settings.svelte";
  import ConfirmInviteNew from "./routes/ConfirmInviteNew.svelte";
  import Invitations from "./routes/Invitations.svelte";

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

    .close {
        cursor: pointer;
        text-align: right;
    }

    .err-container {
        display: flex;
        flex-direction: row;
    }

    footer > :global(a) {
        color: white;
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
        <div class="main-nav signup"><a href="/user/search"><Fa icon="{faHome}" size="sm" class="icon" /></a></div>
        {#if $user.role === "USR" && $user.image}
          <img class="image-usr-sx" src="{$user.image}" />
        {:else if $user.image}
          <img class="image-org-sx" src="{$user.image}" />
        {/if}
        <div class="main-nav"><a href="/user/settings">{$user.email}</a></div>
        <div class="main-nav"><a href="/login" on:click={removeSession}>Sign out</a></div>
      {:else}
        <div class="main-nav"><a href="/login">Login</a></div>
        <div class="main-nav signup"><a href="/signup">Sign up</a></div>
      {/if}
    </nav>
  </header>

  {#if $error}<div class="bg-red p-2 parent err-container"><div class="w-100">{@html $error}</div><div class="close" on:click|preventDefault="{() => {$error=null}}">✕</div></div>{/if}

  <main>
    <Router url="{url}">

      {#if loading}
        <Spinner />
      {:else}
        <Route path="/" component="{Landing}" />

        <Route path="/login" component="{Login}" />
        <Route path="/signup" component="{Signup}" />
        <Route path="/forgot" component="{Forgot}" />
        <Route path="/badges/:uuid" component="{PublicBadges}" />

        <Route path="/confirm/reset/:email/:token" component="{ConfirmForgot}" />
        <Route path="/confirm/signup/:email/:token" component="{ConfirmSignup}" />
        <Route path="/confirm/git-email/:email/:token" component="{ForwardGitEmail}" />
        <Route path="/confirm/invite-new/:email/:emailToken/:inviteEmail/:expireAt/:inviteToken/:inviteMeta" component="{ConfirmInviteNew}" />
        <Route path="/confirm/invite/:email/:inviteEmail/:expireAt/:inviteToken/:inviteMeta" component="{ForwardInvite}" />

        {#if $user.id && ($route.pathname.startsWith("/user") || pathname.startsWith("/user"))}
          <Route path="/user/search" component="{Search}" />
          <Route path="/user/payments" component="{Payments}" />
          <Route path="/user/settings" component="{Settings}" />
          <Route path="/user/income" component="{Income}" />
          <Route path="/user/badges" component="{Badges}" />
          <Route path="/user/admin" component="{Admin}" />
          <Route path="/user/invitations" component="{Invitations}" />
        {/if}
        <Route path="*" component="{CatchAll}" />
      {/if}

    </Router>
  </main>

  <footer>© flatfeestack.io. We used the following <a href="dependencies.txt">dependencies</a></footer>
</div>
