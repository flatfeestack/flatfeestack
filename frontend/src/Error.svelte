<script lang="ts">
    import { appState } from "ts/state.svelte.ts";
    let formattedError = $derived(appState.error?.replace(/\n/g, '<br>'));
</script>

{#if appState.error}
    <div class="bg-red p-2 err-container">
        <div>{@html formattedError}</div>
        <button class="close accessible-btn" onclick={() => {appState.error = "";}}>✕</button>
    </div>
{/if}

<style>
    .err-container {
        position: relative;
        padding-right: 2.5rem;
    }
    
    .close {
        position: absolute;
        top: 0;
        right: 0;
        cursor: pointer;
        padding: 0 0.3rem;
    }
</style>