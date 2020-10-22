<style type="text/scss">
.wrapper {
  background: transparent;
  transition: all ease 0.3s;
  height: var(--navbar-height);
}
.nav-wrapper {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-x3);
  max-width: 1440px;
  margin: auto;
}
.nav-link {
  font-family: "Open Sans", sans-serif;
  font-weight: 400;
  font-size: 18px;
  letter-spacing: 0.26px;
  color: var(--text);
  padding: 0 var(--spacing-x4);
  &:hover {
    color: var(--light);
    text-decoration: none;
  }
}
.link-primary {
  color: var(--primary);
  text-decoration: underline;
}
.sticky {
  position: fixed;
  top: 0;
  width: 100%;
  box-shadow: 0 5px 15px 2px rgba(0, 0, 0, 0.06);
  background: white;
}
</style>

<script lang="ts">
import { Link } from "svelte-routing";
import { user } from "src/store/auth";
let yScroll;
$: sticky = yScroll > 0;
</script>

<svelte:window bind:scrollY="{yScroll}" />
<div class="wrapper {sticky ? 'sticky' : ''}">
  <div class="nav-wrapper">
    <Link to="/"><img src="/assets/images/logo.svg" alt="Flatfeestack" /></Link>
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

      {#if $user}
        <Link to="/dashboard">
          <div class="nav-link link-primary">Dashboard</div>
        </Link>
      {:else}
        <Link to="/login">
          <div class="nav-link link-primary">Login</div>
        </Link>
      {/if}
    </div>
  </div>
</div>
