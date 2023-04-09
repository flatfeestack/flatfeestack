<script lang="ts">
  import {
    faUserFriends,
    faSearch,
    faMedal,
    faUserCog,
    faCreditCard,
    faHandHoldingUsd,
    faShieldAlt,
  } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import { isSubmitting, user, route } from "../../ts/mainStore";
  import Spinner from "../../components/Spinner.svelte";
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
  nav :global(a),
  nav .inactive {
    display: block;
    color: var(--secondary-700);
    padding: 1em;
    text-decoration: none;
    transition: color 0.3s linear, background-color 0.3s linear;
  }

  nav .inactive {
    color: var(--secondary-300);
  }

  nav :global(a:hover),
  nav .selected {
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
      justify-content: space-between;
      width: 99.9%;
      border-bottom: solid 1px var(--primary-500);
      padding: 0;
    }
    nav :global(a) {
      text-align: center;
      width: 100%;
      float: left;
    }
  }
</style>

<div class="page">
  <nav>
    <a
      href="/user/settings"
      class={$route.pathname === `/user/settings` ? `selected` : ``}
    >
      <Fa icon={faUserCog} size="sm" class="icon" />
      <span class="hide-sx">Settings</span>
    </a>
    <a
      href="/user/search"
      class={$route.pathname === `/user/search` ? `selected` : ``}
    >
      <Fa icon={faSearch} size="sm" class="icon" />
      <span class="hide-sx">Search</span>
    </a>
    <a
      href="/user/payments"
      class={$route.pathname === `/user/payments` ? `selected` : ``}
    >
      <Fa icon={faCreditCard} size="sm" class="icon" />
      <span class="hide-sx">Payments</span>
    </a>
    <a
      href="/user/income"
      class={$route.pathname === `/user/income` ? `selected` : ``}
    >
      <Fa icon={faHandHoldingUsd} size="sm" class="icon" />
      <span class="hide-sx">Income</span>
    </a>
    <a
      href="/user/invitations"
      class={$route.pathname === `/user/invitations` ? `selected` : ``}
    >
      <Fa icon={faUserFriends} size="sm" class="icon" />
      <span class="hide-sx">Invitations</span>
    </a>
    <a
      href="/user/badges"
      class={$route.pathname === `/user/badges` ? `selected` : ``}
    >
      <Fa icon={faMedal} size="sm" class="icon" />
      <span class="hide-sx">Badges</span>
    </a>
    {#if $user.role === "admin"}
      <a
        href="/user/admin"
        class={$route.pathname === `/user/admin` ? `selected` : ``}
      >
        <Fa icon={faShieldAlt} size="sm" class="icon" />
        <span class="hide-sx">Admin</span>
      </a>
    {/if}
  </nav>
  <div>
    {#if $isSubmitting}<Spinner />{/if}
    <slot />
  </div>
</div>
