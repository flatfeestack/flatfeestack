<script lang="ts">
  import {appState} from "./ts/state.ts";
  import Spinner from "./Spinner.svelte";
  import NavItem from "./NavItem.svelte";
</script>

<style>
  .page {
    flex: 1 1 auto;
    display: flex;
  }
  nav {
    padding-top: 2rem;
    display: flex;
    flex-flow: column;
    min-width: 12rem;
    background-color: var(--secondary-100);
    border-right: solid 1px var(--secondary-300);
    white-space: nowrap;
  }
  nav :global(a) {
    display: block;
    color: var(--secondary-700);
    padding: 1em;
    text-decoration: none;
    transition: color 0.3s linear, background-color 0.3s linear;
  }

  nav :global(a:hover) {
    background-color: var(--primary-300);
    color: var(--primary-900);
  }

  @media (max-width: 36rem) {
    .page {
      flex-direction: column;
      display: flex;
    }
    nav {
      display: flex;
      flex-direction: row;
      justify-content: space-around;
      width: 100%;
      border-bottom: solid 1px var(--primary-500);
      padding: 0;
    }
    nav :global(a) {
      text-align: center;
      width: 100%;
      padding: 0.5em;
    }
  }
</style>

<div class="page">
  <nav>
    <NavItem href="/user/settings" icon="fa-user-cog" label="Settings" />
    <NavItem href="/user/search" icon="fa-search" label="Search" />
    <NavItem href="/user/payments" icon="fa-credit-card" label="Payments" />
    <NavItem href="/user/income" icon="fa-hand-holding-usd" label="Income" />
    <NavItem href="/user/invitations" icon="fa-user-friends" label="Invitations"/>
    <NavItem href="/user/badges" icon="fa-medal" label="Badges" />
    {#if appState.$state.user.role === "admin"}
      <NavItem href="/user/admin" icon="fa-shield-alt" label="Admin" />
    {/if}
  </nav>
  <div>
    {#if appState.$state.isSubmitting}<Spinner />{/if}
    <slot />
  </div>
</div>
