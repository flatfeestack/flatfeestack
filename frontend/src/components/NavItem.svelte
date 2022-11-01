<script lang="ts">
    import {route} from "../ts/store";
    import Fa from "svelte-fa";

    export let href;
    export let icon;
    export let label;
</script>

<a href="{href}" class="{$route.pathname === href ? `selected`:``}">
    <Fa icon="{icon}" size="sm" class="icon" />
    <span class="hide-sx">{label}</span>
</a>