<style type="text/scss">
.nav-item {
  @apply block mt-4 text-primary-700;
}
.nav-item:hover {
  @apply text-black;
}
@screen lg {
  .nav-item {
    @apply inline-block mt-0 ml-4;
  }
}
</style>

<script lang="ts">
import { Link } from "svelte-routing";
import { user } from "src/store/auth";
let sticky = false;
let yScroll;
$: sticky = yScroll > 0;
let showMenu = false;
$: console.log(sticky);
</script>

<svelte:window bind:scrollY="{yScroll}" />
<div
  class="py-6 px-1 bg-white {sticky ? 'shadow-2xl bg-white' : 'lg:bg-transparent'} fixed w-full top-0 transition duration-300"
>
  <nav
    class="container mx-auto flex items-end justify-between flex-wrap sticky px-3 lg:px-0"
  >
    <div class="flex flex-shrink-0 text-white mr-6">
      <Link to="/">
        <img src="assets/images/logo.svg" alt="Flatfeestack" />
      </Link>
    </div>
    <div class="lg:hidden">
      <button
        on:click="{() => (showMenu = !showMenu)}"
        class="flex items-center px-3 py-2 border rounded text-primary-700 border-primary-400 hover:text-black hover:border-white"
      >
        <svg
          class="fill-current h-3 w-3"
          viewBox="0 0 20 20"
          xmlns="http://www.w3.org/2000/svg"
        ><title>Menu</title>
          <path d="M0 3h20v2H0V3zm0 6h20v2H0V9zm0 6h20v2H0v-2z"></path></svg>
      </button>
    </div>
    <div
      class="{!showMenu ? 'hidden' : ''} lg:show block w-full lg:w-0 lg:flex lg:items-end lg:w-auto"
    >
      <div class="lg:flex-grow">
        <Link to="/">
          <div class="nav-item">Home</div>
        </Link>
        <Link to="/about">
          <div class="lg:px-3 nav-item">About</div>
        </Link>
        <Link to="/integrate">
          <div class="lg:px-3 nav-item">Integrate</div>
        </Link>
        <Link to="/score">
          <div class="lg:px-3 nav-item">Score</div>
        </Link>

        {#if user}
          <Link to="/dashboard">
            <div class="nav-item">Dashboard</div>
          </Link>
        {:else}
          <Link to="/login">
            <div class="nav-item">Login</div>
          </Link>
        {/if}
      </div>
    </div>
  </nav>
</div>

<!--
<svelte:window bind:scrollY="{yScroll}" />
<div class="wrapper {sticky ? 'sticky' : ''} bg-blue-700">
  <div class="nav-wrapper">
    <Link to="/"><img src="assets/images/logo.svg" alt="Flatfeestack" /></Link>
    <div class="flex">
      <Link to="/">
        <div class="nav-link">Home</div>
      </Link>
      <Link to="/about">
        <div class="nav-link">About</div>
      </Link>
      <Link to="/integrate">
        <div class="nav-link">Integrate</div>
      </Link>
      <Link to="/score">
        <div class="nav-link">Score</div>
      </Link>

      {#if loggedIn}
        <Link to="/dashboard">
          <div class="nav-link &lt;!&ndash;link-primary&ndash;&gt;">Dashboard</div>
        </Link>
      {:else}
        <Link to="/login">
          <div class="nav-link &lt;!&ndash;link-primary&ndash;&gt;">Login</div>
        </Link>
      {/if}
    </div>
  </div>
</div>
-->
