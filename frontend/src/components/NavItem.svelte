<script lang="ts">
  import type { IconDefinition } from "@fortawesome/free-solid-svg-icons";
  import { route, activeRootRoute } from "../ts/mainStore";
  import Fa from "svelte-fa";

  export let href: string;
  export let icon: IconDefinition;
  export let label: string;
  export let sublinks: { href: string; label: string }[] = [];

  let hasSubPages = sublinks.length > 0;

  $: isActiveRoot = $activeRootRoute === href;
  $: isRootRoute = href.split('/').length === 3;
  $: sublinksExpanded = isActiveRoot && hasSubPages;

  function handleMainClick(event: MouseEvent) {
    if (isRootRoute && hasSubPages) {
      event.preventDefault();
      activeRootRoute.set(href);
    } else {
      activeRootRoute.set(null);
    }
  }
</script>

<style>
  .selected {
    background-color: var(--primary-300);
    color: var(--primary-900);
  }
  ul.sublinks {
      list-style-type: none;
      margin: 0 0 0 1em;
      padding: 0;
  }
  ul.sublinks a {
      padding: 0.5em;
  }
  ul.sublinks span {
      font-size: 1.1rem;
  }
</style>

<a {href} on:click={handleMainClick}
   class={$route.pathname === href ? `selected` : ``}>
  <Fa {icon} size="sm" class="icon" />
  <span class="hide-sx">{label}</span>
</a>

{#if sublinksExpanded}
  <ul class="sublinks">
    {#each sublinks as { href, label }}
      <li>
        <a href={href} class="sublink {$route.pathname === href ? 'selected' : ''}">
          <Fa {icon} size="sm" class="icon" />
          <span class="hide-sx">{label}</span>
        </a>
      </li>
    {/each}
  </ul>
{/if}
