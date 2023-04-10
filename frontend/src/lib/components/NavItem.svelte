<script lang="ts">
  import { page } from "$app/stores";
  import Fa from "svelte-fa";

  export let href: string;
  export let icon: string;
  export let label: string;
</script>

<a {href} class={$page.url.pathname === href ? `selected` : ``}>
  <Fa {icon} size="sm" class="icon" />
  <span class="hide-sx">{label}</span>
</a>
