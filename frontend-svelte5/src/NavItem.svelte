<script lang="ts">
  import { route } from "@/mainStore";

  export let href: string;
  export let icon: string;
  export let label: string;
  import '@fortawesome/fontawesome-free/css/all.min.css'
</script>

<style>
  .selected {
    background-color: var(--primary-300);
    color: var(--primary-900);
  }
</style>

<a {href} class={$route.pathname === href ? `selected` : ``}>
  <i class="{icon}"></i>
  <span class="hide-sx">{label}</span>
</a>
