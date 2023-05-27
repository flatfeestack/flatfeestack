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
  import { links } from "svelte-routing";
  import { isSubmitting, user } from "../ts/mainStore";
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
  <nav use:links>
    <NavItem href="/user/settings" icon={faUserCog} label="Settings" />
    <NavItem href="/user/search" icon={faSearch} label="Search" />
    <NavItem href="/user/payments" icon={faCreditCard} label="Payments" />
    <NavItem href="/user/income" icon={faHandHoldingUsd} label="Income" />
    <NavItem
      href="/user/invitations"
      icon={faUserFriends}
      label="Invitations"
    />
    <NavItem href="/user/badges" icon={faMedal} label="Badges" />

    {#if $user.role === "admin"}
      <NavItem href="/user/admin" icon={faShieldAlt} label="Admin" />
    {/if}
  </nav>
  <div>
    {#if $isSubmitting}<Spinner />{/if}
    <slot />
  </div>
</div>
