<script lang="ts">
  import type { IconDefinition } from "@fortawesome/free-solid-svg-icons";
  import { route } from "../ts/mainStore";
  import Fa from "svelte-fa";

  export let href: string;
  export let icon: IconDefinition;
  export let label: string;
</script>

<style>
  .selected {
    background-color: var(--primary-300);
    color: var(--primary-900);
  }
</style>

<a {href} class={$route.pathname === href ? `selected` : ``}>
  <Fa {icon} size="sm" class="icon" />
  <span class="hide-sx">{label}</span>
</a>
