<script lang="ts">
  import {appState} from "./ts/state.svelte.ts";

  export let href: string;
  export let icon: string;
  export let label: string;
  import '@fortawesome/fontawesome-free/css/all.min.css'
</script>

<style>
</style>

<a {href} class={appState.route === href ? `selected` : ``}>
  <i class="{icon}"></i>
  <span class="hide-sx">{label}</span>
</a>
