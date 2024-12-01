<script lang="ts">
    import type {Snippet} from "svelte";
    import {route} from "@mateothegreat/svelte5-router";
    let {children}: {children: Snippet} = $props();
</script>

<style>
    .max {
        width: 100vw;
        height: 75vh;
        display: flex;
        justify-content: center;
        align-items: center;
    }
    .box-container {
        width: 25rem;
        background-color: var(--primary-100);
    }
</style>

<div class="max">
    <div class="box-container rounded p-125">
        <a use:route href="/">
            <img class="hide-sx imgNormalLogo" src="/images/ffs-logo.svg" alt="FlatFeeStack"/>
        </a>
        {@render children?.()}
    </div>
</div>