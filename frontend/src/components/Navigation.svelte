<script type="ts">
  import { faUserFriends, faSearch, faMedal, faUserCog, faCreditCard, faHandHoldingUsd, faShieldAlt } from "@fortawesome/free-solid-svg-icons";
  import { links } from "svelte-routing";
  import Fa from "svelte-fa";
  import { isSubmitting, user, route } from "../ts/store";
  import Spinner from "./Spinner.svelte";

  let pathname="/";
  if (typeof window !== "undefined") {
    pathname = window.location.pathname;
  }

</script>

<style>
    .page {
        flex: 1 1 auto;
        display: flex;
    }
    nav {
        padding-top: 3em;
        display: flex;
        flex-flow: column;
        min-width: 12rem;
        background-color: var(--secondary-100);
        border-right: solid 1px var(--secondary-300);
        white-space: nowrap;

    }
    nav :global(a), nav .inactive{
        display: block;
        color: var(--secondary-900);
        padding: 1em;
        text-decoration: none;
    }

    nav .inactive{
        color: var(--secondary-300);
    }

    nav :global(a:hover), nav .selected  {
        background-color: var(--primary-500);
        color: var(--secondary-100);
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
    <a href="/user/settings" class="{$route.pathname === `/user/settings` ? `selected`:``}">
      <Fa icon="{faUserCog}" size="sm" class="icon" />
      <span class="hide-sx">Settings</span>
    </a>
    {#if $user.role != "ORG" }
    <a href="/user/search" class="{$route.pathname === `/user/search` ? `selected`:``}">
      <Fa icon="{faSearch}" size="sm" class="icon" />
      <span class="hide-sx">Search</span>
    </a>
    {:else}
      <div class="inactive">
        <Fa icon="{faSearch}" size="sm" class="icon" />
        <span class="hide-sx">Search</span>
      </div>
    {/if}
    <a href="/user/payments" class="{$route.pathname === `/user/payments` ? `selected`:``}">
      <Fa icon="{faCreditCard}" size="sm" class="icon" />
      <span class="hide-sx">Payments</span>
    </a>
    {#if $user.role != "ORG" }
      <a href="/user/income" class="{$route.pathname === `/user/income` ? `selected`:``}">
        <Fa icon="{faHandHoldingUsd}" size="sm" class="icon" />
        <span class="hide-sx">Income</span>
      </a>
    {:else}
      <a href="/user/invitations" class="{$route.pathname === `/user/invitations` ? `selected`:``}">
        <Fa icon="{faUserFriends}" size="sm" class="icon" />
        <span class="hide-sx">Invitations</span>
      </a>
    {/if}
    <a href="/user/badges" class="{$route.pathname === `/user/badges` ? `selected`:``}">
      <Fa icon="{faMedal}" size="sm" class="icon" />
      <span class="hide-sx">Badges</span>
    </a>
    {#if $user.email.endsWith("@flatfeestack.io") || $user.email.endsWith("@bocek.ch") || $user.email.endsWith("@machados.org") }
      <a href="/user/admin" class="{$route.pathname === `/user/admin` ? `selected`:``}">
        <Fa icon="{faShieldAlt}" size="sm" class="icon" />
        <span class="hide-sx">Admin</span>
      </a>
    {/if}
  </nav>
  <div>
    {#if $isSubmitting}<Spinner />{/if}
    <slot />
  </div>
</div>
