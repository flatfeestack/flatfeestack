<script type="ts">
import {faMedal, faSearch, faUser, faHandHoldingUsd, faShieldAlt, } from "@fortawesome/free-solid-svg-icons";
import { links } from "svelte-routing";
import Fa from "svelte-fa";
import { user } from "../ts/store";
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
    nav :global(a){
        display: block;
        color: var(--secondary-900);
        padding: 1em;
        text-decoration: none;
    }
    nav :global(a:hover) {
        background-color: var(--primary-500);
        color: var(--secondary-100);
    }
    nav .ac {
        display: block;
        color: lightgray;
        padding: 1em;
        text-decoration: none;
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
  <nav class="sideBar" use:links>
    <a href="/dashboard/search">
      <Fa icon="{faSearch}" size="sm" class="icon" />
      <span class="hide-sx">Search</span>
    </a>
    <a href="/dashboard/profile">
      <Fa icon="{faUser}" size="sm" class="icon" />
      <span class="hide-sx">Profile</span>
    </a>
    {#if $user.role != "ORG" }
      <a href="/dashboard/income">
        <Fa icon="{faHandHoldingUsd}" size="sm" class="icon" />
        <span class="hide-sx">Income</span>
      </a>
    {:else}
      <div class="ac">
        <Fa icon="{faHandHoldingUsd}" size="sm" class="icon" />
        <span class="hide-sx">Income</span>
      </div>
    {/if}
    <a href="/dashboard/badges">
      <Fa icon="{faMedal}" size="sm" class="icon" />
      <span class="hide-sx">Badges</span>
    </a>
    {#if $user.email.endsWith("@flatfeestack.io") || $user.email.endsWith("@bocek.ch") || $user.email.endsWith("@machados.org") }
      <a href="/dashboard/admin">
        <Fa icon="{faShieldAlt}" size="sm" class="icon" />
        <span class="hide-sx">Admin</span>
      </a>
    {/if}
  </nav>
  <div>
    <slot />
  </div>
</div>
