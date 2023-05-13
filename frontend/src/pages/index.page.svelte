<script lang="ts">
  import { Router, Route, navigate, link } from "svelte-routing";
  import { user, route, loginFailed, error, token } from "../ts/mainStore";
  import {hasAccessToken, removeSession} from "../ts/services";
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
  import DifferentChainId from "../routes/DifferentChainId.svelte";

  import DAOHome from "../routes/DAO/Home.svelte";
  import DAOVotes from "../routes/DAO/Votes.svelte";
  import DAOMembership from "../routes/DAO/Membership.svelte";
  import DAOMetamaskRequired from "../routes/DAO/MetaMaskRequired.svelte";
  import DAOCreateProposal from "../routes/DAO/CreateProposal.svelte";
  import DAOCastVotes from "../routes/DAO/CastVotes.svelte";
  import DAOExecuteProposals from "../routes/DAO/ExecuteProposals.svelte";
  import DAOCouncil from "../routes/DAO/Council.svelte";
  import DAOTreasury from "../routes/DAO/Treasury.svelte";
  import DAODiscussions from "../routes/DAO/Discussions.svelte";
  import DAOCreateDiscussion from "../routes/DAO/CreateDiscussion.svelte";
  import DAOShowDiscussion from "../routes/DAO/ShowDiscussion.svelte";
  import DAOEditDiscussion from "../routes/DAO/EditDiscussion.svelte";

  //https://github.com/EmilTholin/svelte-routing/issues/41
  import { globalHistory } from "svelte-routing/src/history";
  import Header from "../components/Header.svelte";
  import Footer from "../components/Footer.svelte";
  import EmptyUser from "../routes/EmptyUser.svelte";

  $route = globalHistory.location;
  globalHistory.listen((history) => {
    $route = history.location;
  });

  export let urlOriginal;
  export let showEmptyUser;

  let loading = true;
  let auth = false;

  function logout() {
    removeSession();
    navigate("/login");
  }

  $:{
    if($token) {
      auth = true;
    }
  }

  onMount(async () => {
    const authCookie = document.cookie.split('; ').find(row => row.startsWith('auth='));
    if(authCookie || $token || hasAccessToken()) {
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

<div class="all">

  <Header/>

  <main>
    <Modal>
      <Router url={urlOriginal}>
        <Route path="/confirm/reset/:email/:token" component={ConfirmForgot} />
        <Route path="/confirm/signup/:email/:token" component={ConfirmSignup} />
        <Route path="/confirm/git-email/:email/:token" component={ForwardGitEmail}/>
        <Route path="/confirm/invite/:email/:emailToken/:inviteByEmail" component={ConfirmInvite}/>

        <Route path="/user/search" component={auth ? Search:Landing} />
        <Route path="/user/payments" component={auth ? Payments:Landing} />
        <Route path="/user/settings" component={auth ? Settings:Landing} />
        <Route path="/user/income" component={auth ? Income:Landing} />
        <Route path="/user/badges" component={auth ? Badges:Landing} />
        <Route path="/user/admin" component={auth ? Admin:Landing} />
        <Route path="/user/invitations" component={auth ? Invitations:Landing} />

        <Route path="/dao/home" component={DAOHome} />
        <Route path="/dao/votes" component={DAOVotes} />
        <Route path="/dao/membership" component={DAOMembership} />
        <Route path="/dao/metamask" component={DAOMetamaskRequired} />
        <Route path="/dao/createProposal" component={DAOCreateProposal} />
        <Route path="/dao/castVotes/:blockNumber" component={DAOCastVotes} />
        <Route path="/dao/executeProposals/:blockNumber" component={DAOExecuteProposals}/>
        <Route path="/dao/council" component={DAOCouncil} />
        <Route path="/dao/treasury" component={DAOTreasury} />
        <Route path="/dao/discussions" component={DAODiscussions} />
        <Route path="/dao/createDiscussion" component={DAOCreateDiscussion} />
        <Route path="/dao/discussion/:postId" component={DAOShowDiscussion} />
        <Route path="/dao/discussion/:postId/edit" component={DAOEditDiscussion}/>

        <Route path="/differentChainId" component={DifferentChainId} />
        <Route path="/badges/:uuid" component={PublicBadges} />
        <Route path="/forgot" component={Forgot} />
        <Route path="/signup" component={Signup} />
        <Route path="/login" component={Login} />
        {#if showEmptyUser}
          <Route path="/" component={EmptyUser} />
        {:else}
          <Route path="/" component={Landing} />
        {/if}
        <Route path="*" component={CatchAll} />
      </Router>
    </Modal>
  </main>

<Footer/>
</div>
