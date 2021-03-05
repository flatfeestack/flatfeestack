<script lang="ts">
  import { links, Route, Router } from "svelte-routing";
  import { user, loading } from "ts/auth.ts";
  import { refreshSession, removeSession } from "ts/authService";
  import { onMount } from "svelte";

  import Landing from "./routes/Landing.svelte";
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
  onMount(() => refreshSession());

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
        height: 2em;
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
        <div class="main-nav"><a href="/signin" on:click={removeSession}>Sign out</a></div>
      {:else}
        <div class="main-nav"><a href="/signin">Sign in</a></div>
        <div class="main-nav signup"><a href="/signup">Sign up</a></div>
      {/if}
    </nav>
  </header>

  <main>
    <Router url="{url}">

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
        <Route path="/confirm/invite/:email/:token/:invite_email" let:params>
          <ConfirmInvite email="{params.email}" token="{params.token}" inviteEmail="{params.invite_email}"/>
        </Route>
        {#if $user.id}
          <Route path="/dashboard" component="{FindRepos}" />
          <Route path="/dashboard/sponsoring" component="{Sponsoring}" />
          <Route path="/dashboard/income" component="{Income}" />
          <Route path="/dashboard/profile" component="{Profile}" />
          <Route path="/dashboard/admin" component="{Admin}" />
        {/if}
        <Route path="*" component="{CatchAll}" />
      {/if}

    </Router>
  </main>

  <footer>© flatfeestack.io</footer>
</div>
