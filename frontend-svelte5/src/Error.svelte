<script lang="ts">
    import { appState } from "ts/state.svelte.ts";
    import '@fortawesome/fontawesome-free/css/all.min.css'
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
        margin-right: 2rem;
    }
</style>

{#if appState.error}
    <div class="bg-red p-2 err-container">
        <div class="w-100">{@html appState.error}</div>
        <div>
            <button class="close accessible-btn" onclick={() => {appState.error = "";}}>
                ✕
            </button>
        </div>
    </div>
{/if}
