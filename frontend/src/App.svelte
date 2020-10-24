<script lang="ts">
import { Router, Link, Route } from "svelte-routing";
import About from "./routes/About.svelte";
import Landing from "./routes/Landing.svelte";
import Login from "./routes/Login.svelte";
import Integrate from "./routes/Integrate.svelte";
import Score from "./routes/Score.svelte";
import Signup from "./routes/Signup.svelte";
import Dashboard from "./routes/Dashboard/Dashboard.svelte";
import Navigation from "./components/Navigation.svelte";
import Footer from "./components/Footer.svelte";
import CatchAll from "./routes/CatchAllRoute.svelte";
import { user } from "./store/auth.ts";
import { onMount } from "svelte";
import { tryToAuthenticate } from "./store/authService";
import Modal from "svelte-simple-modal";
import { ROUTES } from "./types/routes";
import Sponsoring from "./routes/Dashboard/Sponsoring.svelte";
import Income from "./routes/Dashboard/Income.svelte";
import Settings from "./routes/Dashboard/Settings.svelte";
import Profile from "./routes/Dashboard/Profile.svelte";

onMount(() => tryToAuthenticate());

export let url = "";
</script>

<Router url="{url}">
  <Modal>
    <div>
      <Route path="{ROUTES.ABOUT}" component="{About}" />
      <Route path="{ROUTES.LOGIN}" component="{Login}" />
      <Route path="{ROUTES.INTEGRATE}" component="{Integrate}" />
      <Route path="{ROUTES.SCORE}" component="{Score}" />
      <Route path="{ROUTES.SIGNUP}" component="{Signup}" />
      <Route path="/">
        <Landing />
      </Route>
      {#if $user}
        <Route path="{ROUTES.DASHBOARD_OVERVIEW}" component="{Dashboard}" />
        <Route path="{ROUTES.DASHBOARD_SPONSORING}" component="{Sponsoring}" />
        <Route path="{ROUTES.DASHBOARD_INCOME}" component="{Income}" />
        <Route path="{ROUTES.DASHBOARD_SETTINGS}" component="{Settings}" />
        <Route path="{ROUTES.DASHBOARD_PROFILE}" component="{Profile}" />
      {/if}
      <Route path="*" component="{CatchAll}" />
    </div>
  </Modal>
</Router>

<svelte:head>
  <style src="styles.scss">
  </style>
  <link
    href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@300;400;600;700;800&family=Raleway:wght@100;200;300;400;500;600;700;800;900&display=swap"
    rel="stylesheet"
  />
</svelte:head>
