<script lang="ts">
    import { appState } from "ts/state.svelte.ts";
    import { library, icon } from '@fortawesome/fontawesome-svg-core';
    import { faClose } from '@fortawesome/free-solid-svg-icons';
        
    library.add(faClose);
    const closeIcon = icon({ prefix: 'fas', iconName: 'close' }).html[0];
</script>

<style>
    .close {
        cursor: pointer;
        text-align: right;
    }
    .err-container {
        position: fixed;
        width: 100%;
        display: flex;
        flex-direction: row;
    }
    .err-container button {
        margin-right: 1rem;
    }
</style>

{#if appState.error}
    <div class="bg-red p-2 err-container">
        <div class="w-100">{@html appState.error}</div>
        <div>
            <button class="close accessible-btn" onclick={() => {appState.error = "";}}>
                {@html closeIcon}
            </button>
        </div>
    </div>
{/if}
