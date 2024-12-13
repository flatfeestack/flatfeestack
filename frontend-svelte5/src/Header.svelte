<script lang="ts">
    import {appState} from "ts/state.svelte.ts";
    import {getRefreshToken, removeSession} from "auth/auth.svelte.ts";
    import {onMount} from "svelte";
    import {API} from "./ts/api.ts";
    import '@fortawesome/fontawesome-free/css/all.min.css'
    import {goto} from "@mateothegreat/svelte5-router";
    import Error from './Error.svelte';

    function logout(event: MouseEvent) {
        event.preventDefault();
        removeSession();
        goto("/login");
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            removeSession();
            goto("/login");
        }
    }

    onMount(async () => {
        try {
            if (!appState.user.id && getRefreshToken() !== null) {
                appState.user = await API.user.get();
                appState.user = await API.user.get();
            }
        } catch (e) {
            appState.setError(e);
        }
    });

    let isMenuOpen = $state(false);

    function toggleMenu() {
        isMenuOpen = !isMenuOpen;
    }

    function handleClickOutside(event: Event) {
        if (isMenuOpen) {
            const menu = document.querySelector('.user-menu-container');
            if (menu && !menu.contains(event.target as Node)) {
                isMenuOpen = false;
            }
        }
    }

    function gotoSettings() {
        goto('/user/settings');
    }
</script>

<svelte:window on:click={handleClickOutside}/>

<style>
    header {
        display: flex;
        background-color: #fff;
        justify-content: space-between;
        flex: 0 0 auto;
    }

    .user-menu-button:hover {
        border: 1px solid var(--primary-900);
    }

    .user-menu-container {
        position: relative;
        display: inline-block;
    }

    .user-menu-container button {
        width: 2.5rem;
        height: 2.5rem;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
    }

    .user-menu-container button.fas::before {
        font-size: 1.5rem;
    }

    .menu-dropdown {
        position: absolute;
        right: 0;
        top: 100%;
        background-color: #fff;
        border: 1px solid var(--primary-100);
        border-radius: 4px;
        box-shadow: 0 2px 4px var(--primary-900);
        z-index: 50;
        min-width: 200px;
    }

    .menu {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid var(--primary-100);
        font-size: 0.875rem;
    }

    .menu-item {
        padding: 0.75rem 1rem;
        color: #374151;
        display: flex;
        align-items: center;
        gap: 0.75rem;
        cursor: pointer;
    }

    .menu-item:hover {
        background-color: var(--primary-100);
    }

    .image-org-sx {
        display: flex;
        max-height: 2.4rem;
        max-width: 2.4rem;
        width: auto;
        margin: 0 0.5rem;
    }
</style>

<header class="p-050">
    <nav class="center-flex">
        <div class="user-menu-container mx-100">
            {#if appState.user.image}
                <button onclick={toggleMenu} class="round user-menu-button" aria-label="User menu">
                    <img class="round image-org-sx" src={appState.user.image} alt="user profile img"/>
                </button>
            {:else}
                <button class="round fas fa-user" onclick={toggleMenu} aria-label="User menu"></button>
            {/if}

            {#if isMenuOpen}
                <div class="menu-dropdown" role="menu">
                    <div class="menu">
                        <p class="small">Email:</p>
                        <p>{appState.user.email}</p>
                    </div>
                    <div
                            class="menu menu-item"
                            onclick={gotoSettings}
                            role="menuitem"
                            aria-label="Go to settings"
                            onkeydown={handleKeydown}
                            tabindex="0"
                    >
                        <i class="fas fa-user-cog" aria-hidden="true"></i>
                        <span>Settings</span>
                    </div>
                    <div
                            class="menu menu-item"
                            onclick={logout}
                            role="menuitem"
                            aria-label="Sign out of your account"
                            onkeydown={handleKeydown}
                            tabindex="0"
                    >
                        <i class="fas fa-sign-out-alt" aria-hidden="true"></i>
                        <span>Sign out</span>
                    </div>
                </div>
            {/if}
        </div>
        {#if appState.config.env === "local" || appState.config.env === "stage"}
            <button class="button4 center mx-2" onclick={() => goto("/dao/home")}>DAO</button>
        {/if}
    </nav>
</header>

<Error />