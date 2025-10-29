<script lang="ts">
    import { appState } from "ts/state.svelte.ts";
    import { getRefreshToken } from "auth/auth.svelte.ts";
    import { onMount } from "svelte";
    import { API } from "./ts/api.ts";
    import { link, navigate } from "preveltekit";
    import { fade } from "svelte/transition";
    import Dots from "./Dots.svelte";
    import Error from "./Error.svelte";

    let loading = false;

    onMount(async () => {
        try {
            if (!appState.user.id && getRefreshToken() !== null) {
                loading = true;
                appState.user = await API.user.get();
            }
        } catch (e) {
            appState.setError(e);
        } finally {
            loading = false;
        }
    });
</script>

<header class="p-050">
    <a class="center-flex" href="/" use:link>
        <img class="logo" src="/images/ffs-logo.svg" alt="FlatFeeStack" />
    </a>
    <nav class="center-flex">
        <div class="auth-button-wrapper center-flex">
            {#if appState.user?.id}
                <div transition:fade class="button-overlay">
                    <button
                        class="button1 center p-050"
                        onclick={() => navigate("/user/search")}
                        >Dashboard</button
                    >
                </div>
            {:else}
                <div transition:fade class="button-overlay">
                    <button
                        class="button1 center p-050"
                        onclick={() => navigate("/login")}
                        >Login{#if loading}<Dots></Dots>{/if}</button
                    >
                </div>
            {/if}
        </div>
        {#if appState.config?.env === "local" || appState.config?.env === "stage"}
            <button
                class="button4 center mx-2"
                onclick={() => navigate("/dao/home")}>DAO</button
            >
        {/if}
    </nav>
</header>

<Error></Error>

<style>
    header {
        display: flex;
        justify-content: space-between;
        flex: 0 0 auto;
        background-color: #fff;
    }
    .logo {
        width: 15rem;
    }
    .auth-button-wrapper {
        position: relative;
        width: max-content;
        height: 2.75rem;
    }
    .button-overlay {
        position: absolute;
        top: 50%;
        right: 0;
        transform: translateY(-50%);
    }
</style>
