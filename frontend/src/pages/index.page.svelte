<script lang="ts">
  import { Router, Route, navigate, link } from "svelte-routing";
  import { user, route, loginFailed, error } from "../ts/mainStore";
  import { removeSession } from "../ts/services";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { faHome } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import Modal from "svelte-simple-modal";

  import Landing from "../routes/Landing.svelte";
  import Badges from "../routes/Badges.svelte";
  import PublicBadges from "../routes/PublicBadges.svelte";
  import Login from "../routes/Login.svelte";
  import Signup from "../routes/Signup.svelte";
  import Forgot from "../routes/Forgot.svelte";
  import ConfirmForgot from "../routes/ConfirmForgot.svelte";
  import ConfirmSignup from "../routes/ConfirmSignup.svelte";
  import Search from "../routes/Search.svelte";
  import CatchAll from "../routes/CatchAllRoute.svelte";
  import Income from "../routes/Income.svelte";
  import Payments from "../routes/Payments.svelte";
  import Admin from "../routes/Admin.svelte";
  import ForwardGitEmail from "../routes/ForwardGitEmail.svelte";
  import Settings from "../routes/Settings.svelte";
  import ConfirmInvite from "../routes/ConfirmInvite.svelte";
  import Invitations from "../routes/Invitations.svelte";

  import DaaVotes from "../components/DAA/Votes.svelte";
  import DaaMembershipRequests from "../components/DAA/MembershipRequests.svelte";
  import DaaCreateProposal from "../components/DAA/CreateProposal.svelte";
  import DaaCastVotes from "../components/DAA/CastVotes.svelte";
  import DaaExecuteProposals from "../components/DAA/ExecuteProposals.svelte";
  import DaaCouncil from "../components/DAA/Council.svelte";

  //https://github.com/EmilTholin/svelte-routing/issues/41
  import { globalHistory } from "svelte-routing/src/history";
  $route = globalHistory.location;
  globalHistory.listen((history) => {
    $route = history.location;
  });

  export let urlOriginal;
  let loading = true;

  function logout() {
    removeSession();
    navigate("/login");
  }

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
</script>

<style>
  .all {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  header {
    padding: 1em;
    background-color: #fff;
    border-bottom: 1px #000 solid;
    justify-content: space-between;
    flex: 0 0 auto;
  }

  main {
    flex: 1 0 auto;
    display: flex;
    height: 100%;
  }

  footer {
    background-color: #000;
    color: white;
    flex: 0 0 auto;
    font-size: 1rem;
    padding: 0.5rem;
  }

  header,
  nav {
    display: flex;
    align-items: center;
    font-size: 1.1rem;
  }

  footer > :global(a) {
    color: white;
    font-size: 1rem;
  }

  header :global(a),
  .header :global(a:visited),
  .header :global(a:active) {
    text-decoration: none;
    color: #000;
  }

  .close {
    cursor: pointer;
    text-align: right;
  }

  .err-container {
    display: flex;
    flex-direction: row;
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

<div class="all">
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
        <a href="/user/search" use:link
          ><Fa icon={faHome} size="sm" class="icon" /></a
        >
        {#if $user.image}
          <img class="image-org-sx" src={$user.image} />
        {/if}
        &nbsp;
        {$user.email}
        <form on:submit|preventDefault={logout}>
          <button class="button3 center mx-2" type="submit">Sign out</button>
        </form>
      {:else}
        <form on:submit|preventDefault={() => navigate("/login")}>
          <button class="button3 center mx-2" type="submit">Login</button>
        </form>
        <form on:submit|preventDefault={() => navigate("/signup")}>
          <button class="button1 center" type="submit">Sign Up</button>
        </form>
      {/if}
    </nav>
  </header>

  {#if $error}<div class="bg-red p-2 parent err-container">
      <div class="w-100">{@html $error}</div>
      <div
        class="close"
        on:click|preventDefault={() => {
          $error = null;
        }}
      >
        ✕
      </div>
    </div>{/if}

  <main>
    <Modal>
      <Router url={urlOriginal}>
        <Route path="/confirm/reset/:email/:token" component={ConfirmForgot} />
        <Route path="/confirm/signup/:email/:token" component={ConfirmSignup} />
        <Route
          path="/confirm/git-email/:email/:token"
          component={ForwardGitEmail}
        />
        <Route
          path="/confirm/invite/:email/:emailToken/:inviteByEmail"
          component={ConfirmInvite}
        />

        <Route path="/user/search" component={Search} />
        <Route path="/user/payments" component={Payments} />
        <Route path="/user/settings" component={Settings} />
        <Route path="/user/income" component={Income} />
        <Route path="/user/badges" component={Badges} />
        <Route path="/user/admin" component={Admin} />
        <Route path="/user/invitations" component={Invitations} />

        <Route path="/daa/votes" component={DaaVotes} />
        <Route
          path="/daa/membershipRequests"
          component={DaaMembershipRequests}
        />
        <Route path="/daa/createProposal" component={DaaCreateProposal} />
        <Route path="/daa/castVotes/:blockNumber" component={DaaCastVotes} />
        <Route
          path="/daa/executeProposals/:blockNumber"
          component={DaaExecuteProposals}
        />
        <Route path="/daa/council" component={DaaCouncil} />

        <Route path="/badges/:uuid" component={PublicBadges} />
        <Route path="/forgot" component={Forgot} />
        <Route path="/signup" component={Signup} />
        <Route path="/login" component={Login} />
        <Route path="/" component={Landing} />
        <Route path="*" component={CatchAll} />
      </Router>
    </Modal>
  </main>

  <footer class="text-center">
    We used the following <a href="stats.html">dependencies</a>
  </footer>
</div>
