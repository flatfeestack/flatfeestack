<script lang="ts">
    import Navigation from './Navigation.svelte';
    import Header from './Header.svelte';
    import type {Snippet} from "svelte";

    let {children}: {children: Snippet} = $props();
</script>

<style>
    .dashboard {
        display: flex;
        min-height: 100vh;
        background-color: white;
    }

    .main-content {
        flex: 1;
        min-width: 0; /* Prevents overflow */
    }

    .top-bar {
        display: flex;
        justify-content: flex-end;
    }
</style>

<div class="dashboard">
    <Navigation />
    <div class="main-content">
        <div class="top-bar">
            <Header />
        </div>
        <div class="px-100">
            {@render children?.()}
        </div>
    </div>
</div>