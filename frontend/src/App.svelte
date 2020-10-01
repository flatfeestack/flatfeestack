<script lang="ts">
  import {Router, Link, Route} from "svelte-routing";
  import About from "./routes/About.svelte";
  import Landing from "./routes/Landing.svelte";
  import Login from "./routes/Login.svelte";
  import Integrate from "./routes/Integrate.svelte";
  import Score from "./routes/Score.svelte";
  import Navigation from "./components/Navigation.svelte";
  import Footer from "./components/Footer.svelte";

  export let url = "";

</script>

<style>
  main {
    text-align: center;
    padding: 1em;
    max-width: 240px;
    margin: 0 auto;
  }

  @media (min-width: 640px) {
    main {
      max-width: none;
    }
  }
</style>

<Router url="{url}">
  <Navigation/>
  <div>
    <Route path="about" component="{About}" />
    <Route path="login" component="{Login}" />
    <Route path="integrate" component="{Integrate}" />
    <Route path="score" component="{Score}" />
    <Route path="/"><Landing /></Route>
  </div>
  <Footer/>
</Router>

<svelte:head>
  <style src="styles.scss">
  </style>
  <link href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@300;400;600;700;800&family=Raleway:wght@100;200;300;400;500;600;700;800;900&display=swap" rel="stylesheet">
</svelte:head>
