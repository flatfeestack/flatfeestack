<script lang="ts">
  import { Router, Route } from "svelte-routing";
  import { route } from "../ts/mainStore";
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
  import Test from "../routes/Admin/Test.svelte";
  import TrustedRepos from "../routes/Admin/TrustedRepos.svelte";
  import ForwardGitEmail from "../routes/ForwardGitEmail.svelte";
  import Settings from "../routes/Settings.svelte";
  import ConfirmInvite from "../routes/ConfirmInvite.svelte";
  import Invitations from "../routes/Invitations.svelte";
  import DifferentChainId from "../routes/DifferentChainId.svelte";
  import PrivateRoute from "../routes/PrivateRoute.svelte";

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
  import DAOShowProposal from "../routes/DAO/ShowProposal.svelte";

  //https://github.com/EmilTholin/svelte-routing/issues/41
  import { globalHistory } from "svelte-routing/src/history";
  import Header from "../components/Header.svelte";
  import Footer from "../components/Footer.svelte";

  $route = globalHistory.location;
  globalHistory.listen((history) => {
    $route = history.location;
  });

  export let urlOriginal: string;
</script>

<style>
  .all {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  main {
    flex: 1 0 auto;
    display: flex;
    height: 100%;
  }
</style>

<div class="all">
  <Header />

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
        <PrivateRoute path="/user/search">
          <Search />
        </PrivateRoute>
        <PrivateRoute path="/user/payments">
          <Payments />
        </PrivateRoute>
        <PrivateRoute path="/user/settings">
          <Settings />
        </PrivateRoute>
        <PrivateRoute path="/user/income">
          <Income />
        </PrivateRoute>
        <PrivateRoute path="/user/badges">
          <Badges />
        </PrivateRoute>
        <PrivateRoute path="/user/invitations">
          <Invitations />
        </PrivateRoute>
        <PrivateRoute path="/user/admin/test">
          <Test />
        </PrivateRoute>
        <PrivateRoute path="/user/admin/trusted-repos">
          <TrustedRepos />
        </PrivateRoute>

        <Route path="/dao/home" component={DAOHome} />
        <Route path="/dao/votes" component={DAOVotes} />
        <Route path="/dao/membership" component={DAOMembership} />
        <Route path="/dao/metamask" component={DAOMetamaskRequired} />
        <Route path="/dao/createProposal" component={DAOCreateProposal} />
        <Route path="/dao/proposals/:proposalId" component={DAOShowProposal} />
        <Route path="/dao/castVotes/:blockNumber" component={DAOCastVotes} />
        <Route
          path="/dao/executeProposals/:blockNumber"
          component={DAOExecuteProposals}
        />
        <Route path="/dao/council" component={DAOCouncil} />
        <Route path="/dao/treasury" component={DAOTreasury} />
        <Route path="/dao/discussions" component={DAODiscussions} />
        <Route path="/dao/createDiscussion" component={DAOCreateDiscussion} />
        <Route path="/dao/discussion/:postId" component={DAOShowDiscussion} />
        <Route
          path="/dao/discussion/:postId/edit"
          component={DAOEditDiscussion}
        />

        <Route path="/differentChainId" component={DifferentChainId} />
        <Route path="/badges/:uuid" component={PublicBadges} />
        <Route path="/forgot" component={Forgot} />
        <Route path="/signup" component={Signup} />
        <Route path="/login" component={Login} />
        <Route path="/" component={Landing} />
        <Route path="*" component={CatchAll} />
      </Router>
    </Modal>
  </main>

  <Footer />
</div>
