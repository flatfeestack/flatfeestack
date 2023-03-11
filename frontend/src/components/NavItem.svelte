<script lang="ts">
  import { route } from "../ts/mainStore";
  import Fa from "svelte-fa";

  export let href;
  export let icon;
  export let label;
</script>

<a {href} class={$route.pathname === href ? `selected` : ``}>
  <Fa {icon} size="sm" class="icon" />
  <span class="hide-sx">{label}</span>
</a>
