<script type="ts">
import {faMedal, faSearch, faUser, faCodeBranch, faHandHoldingUsd, faCogs, faShieldAlt, } from "@fortawesome/free-solid-svg-icons";
import { Link, links } from "svelte-routing";
import Fa from "svelte-fa";
import { user } from "ts/auth";

</script>

<style>
    .page {
        flex: 1 1 auto;
        display: flex;
        color: var(--primary-700);
    }
    nav {
        padding-top: 2em;
        display: flex;
        flex-flow: column;
        min-width: 11rem;
        background-color: var(--primary-100);
        border-right: solid 1px var(--primary-500);
        white-space: nowrap;
    }
    nav :global(a){
        display: block;
        color: var(--primary-700);
        padding: 0.75em;
        text-decoration: none;
    }
    nav :global(a:hover) {
        background-color: var(--primary-500);
        color: var(--primary-100);
    }
    nav .ac {
        display: block;
        color: lightgray;
        padding: 0.75em;
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
            width: 100%;
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
  <a href="/dashboard/sponsoring" disabled="true"><Fa icon="{faSearch}" size="sm" class="icon"/>
    <span class="hide-sx">Find Repos</span>
  </a>
  <a href="/dashboard/profile"><Fa icon="{faUser}" size="sm"  class="icon"/>
    <span class="hide-sx">Profile</span>
  </a>
  {#if $user.mode != "ORG" }
  <a href="/dashboard/income"><Fa icon="{faHandHoldingUsd}" size="sm"  class="icon"/>
    <span class="hide-sx">Income</span>
  </a>
  {:else}
    <div class="ac"><Fa icon="{faHandHoldingUsd}" size="sm" class="icon"/>
      <span class="hide-sx">Income</span>
    </div>
  {/if}
  <a href="#"><Fa icon="{faMedal}" size="sm"  class="icon"/>
    <span class="hide-sx">Badges</span>
  </a>
  <a href="/dashboard/admin"><Fa icon="{faShieldAlt}" size="sm"  class="icon"/>
    <span class="hide-sx">Admin</span>
  </a>
</nav>
  <div>
    <slot />
  </div>
</div>
