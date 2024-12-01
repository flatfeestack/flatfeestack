<script lang="ts">
    import {route} from "@mateothegreat/svelte5-router";
    import type { Snippet } from 'svelte';

    let { title, bottomText = "", bottomLinkText = "Log in", bottomLinkHref = "/login", children}:
        {title: string, bottomText:string, bottomLinkText:string, bottomLinkHref:string, children: Snippet} = $props();

</script>

<div class="max">
    <div class="box-container rounded p-5">
        <h2 class="py-5 text-center text-primary-900">{title}</h2>
        {@render children?.()}
        <div class="divider"></div>
        <div class="flex">
            {bottomText}&nbsp;<a href={bottomLinkHref} use:route>{bottomLinkText}</a>
        </div>
    </div>
</div>

<style>
    .max {
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    .flex {
        padding-top: 1em;
        display: flex;
        justify-content: space-between;
    }
</style>



