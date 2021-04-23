<script lang="ts">
  import { links, Route, Router } from "svelte-routing";
  import {globalHistory} from 'svelte-routing/src/history';
  import { user, loading, route, loginFailed } from "ts/auth.ts";
  import { removeSession, updateUser } from "ts/authService";
  import { onMount } from "svelte";

  import Landing from "./routes/Landing.svelte";
  import Badges from "./routes/Dashboard/Badges.svelte";
  import Signin from "./routes/Signin.svelte";
  import Signup from "./routes/Signup.svelte";
  import Forgot from "./routes/Forgot.svelte";
  import ConfirmForgot from "./routes/ConfirmForgot.svelte";
  import ConfirmSignup from "./routes/ConfirmSignup.svelte";
  import FindRepos from "./routes/Dashboard/FindRepos.svelte";
  import CatchAll from "./routes/CatchAllRoute.svelte";
  import Sponsoring from "./routes/Dashboard/FindRepos.svelte";
  import Income from "./routes/Dashboard/Income.svelte";
  import Profile from "./routes/Dashboard/Profile.svelte";
  import Admin from "./routes/Dashboard/Admin.svelte";
  import Spinner from "./components/Spinner.svelte";
  import ConfirmGitEmail from "./routes/ConfirmGitEmail.svelte";
  import ConfirmInvite from "./routes/ConfirmInvite.svelte";

  export let url;

  let pathname = window.location.pathname;
  onMount(() => {
    try {
      updateUser()
    } catch (e) {
      $loginFailed = true;
    }
    }
  );

  //https://github.com/EmilTholin/svelte-routing/issues/41
  //https://github.com/EmilTholin/svelte-routing/issues/62
  $route = globalHistory.location;
  globalHistory.listen(history => {
    $route = history.location;
  });

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
        <span>{$user.email}</span>
        <div class="main-nav"><a href="/signin" on:click={removeSession}>Sign out</a></div>
      {:else}
        <div class="main-nav"><a href="/signin">Sign in</a></div>
        <div class="main-nav signup"><a href="/signup">Sign up</a></div>
      {/if}
    </nav>
  </header>

  <main>
    <Router url="{url}" let:basepath>

      {#if $loading}
        <Spinner />
      {:else}

        <Route path="/" component="{Landing}" />

        <Route path="/signin" component="{Signin}" />
        <Route path="/signup" component="{Signup}" />
        <Route path="/forgot" component="{Forgot}" />

        <Route path="/confirm/reset/:email/:token" component="{ConfirmForgot}" let:params>
          <ConfirmForgot email="{params.email}" token="{params.token}" />
        </Route>
        <Route path="/confirm/signup/:email/:token" let:params>
          <ConfirmSignup email="{params.email}" token="{params.token}" />
        </Route>
        <Route path="/confirm/git-email/:email/:token" let:params>
          <ConfirmGitEmail email="{params.email}" token="{params.token}" />
        </Route>
        <Route path="/confirm/invite/:email/:emailToken/:inviteEmail/:inviteDate/:inviteToken" let:params>
          <ConfirmInvite email="{params.email}"
                         emailToken="{params.emailToken}"
                         inviteEmail="{params.inviteEmail}"
                         inviteDate="{params.inviteDate}"
                         inviteToken="{params.inviteToken}"/>
        </Route>
        {#if $route.pathname.startsWith("/dashboard") || pathname.startsWith("/dashboard")}
          {#if $user.id}
            <Route path="/dashboard" component="{FindRepos}" />
            <Route path="/dashboard/sponsoring" component="{Sponsoring}" />
            <Route path="/dashboard/income" component="{Income}" />
            <Route path="/dashboard/profile" component="{Profile}" />
            <Route path="/dashboard/badges" component="{Badges}" />
            <Route path="/dashboard/admin" component="{Admin}" />
          {:else}
            {#if $loginFailed}
              <Route path="*" component="{Signin}" />
            {:else}
              <Spinner/>
            {/if}
          {/if}
        {:else}
          <Route path="*" component="{CatchAll}" />
        {/if}

      {/if}

    </Router>
  </main>

  <footer>© flatfeestack.io</footer>
</div>
