<script lang="ts">
import { Router, Link, Route } from "svelte-routing";
import About from "./routes/About.svelte";
import Landing from "./routes/Landing.svelte";
import Login from "./routes/Login.svelte";
import Integrate from "./routes/Integrate.svelte";
import Score from "./routes/Score.svelte";
import Signup from "./routes/Signup.svelte";
import Dashboard from "./routes/Dashboard.svelte";
import Navigation from "./components/Navigation.svelte";
import Footer from "./components/Footer.svelte";
import CatchAll from "./routes/CatchAllRoute.svelte";
import { user } from "./store/auth.ts";
import { onMount } from "svelte";
import { tryToAuthenticate } from "./store/authService";
import Modal from "svelte-simple-modal";

onMount(() => tryToAuthenticate());

export let url = "";
</script>

<Router url="{url}">
  <Modal>
    <Navigation />
    <div>
      <Route path="about" component="{About}" />
      <Route path="login" component="{Login}" />
      <Route path="integrate" component="{Integrate}" />
      <Route path="score" component="{Score}" />
      <Route path="signup" component="{Signup}" />
      <Route path="/">
        <Landing />
      </Route>
      {#if $user}
        <Route path="dashboard" component="{Dashboard}" />
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
