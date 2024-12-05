<script lang="ts">
    import { appState } from "ts/state.svelte.ts";
    import {getRefreshToken, removeSession} from "auth/auth.svelte.ts";
    import { onMount } from "svelte";
    import { API } from "./ts/api.ts";
    import '@fortawesome/fontawesome-free/css/all.min.css'
    import {route, goto} from "@mateothegreat/svelte5-router";

    function logout(event: SubmitEvent) {
        event.preventDefault();
        removeSession();
        goto("/login");
    }

    onMount(async () => {
        try {
            if(!appState.user.id && getRefreshToken() !== null) {
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
</script>

<svelte:window on:click={handleClickOutside}/>

<style>
    header {
        padding: 0;
        background-color: #fff;
        border-bottom: 1px #000 solid;
        justify-content: space-between;
        flex: 0 0 auto;
    }

    header,
    nav {
        display: flex;
        align-items: center;
        font-size: 1rem;
    }

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
        margin-right: 30px;
    }
    nav p {
        margin: 0;
    }

    .user-menu-container {
        position: relative;
        display: inline-block;
    }

    .menu-dropdown {
        position: absolute;
        right: 0;
        top: 100%;
        background-color: #fff;
        border: 1px solid #ddd;
        border-radius: 4px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        min-width: 200px;
        z-index: 50;
    }

    .menu-header {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid #ddd;
    }

    .user-email {
        margin: 0;
        font-size: 0.875rem;
        color: #666;
    }

    .menu-item {
        width: 100%;
        text-align: left;
        padding: 0;
        background: none;
        border: none;
        color: #333;
        cursor: pointer;
    }

    .menu-item-content {
        display: flex;
        align-items: center;
        justify-content: flex-start;  /* Aligns items to start */
        padding: 0.75rem 1rem;
        width: 100%;
        font-size: 0.875rem;
    }

    /* Create a wrapper for the icon and text that takes full width */
    .menu-item-content span {
        display: flex;
        align-items: center;
        width: 100%;      /* Takes full width */
        gap: 0.75rem;     /* Space between icon and text */
    }

    /* Style the icon container specifically */
    .menu-item-content i {
        display: flex;          /* Ensures proper icon alignment */
        align-items: center;
        min-width: 1rem;       /* Ensures consistent icon width */
    }

    .menu-item-content:hover {
        background-color: #f5f5f5;
    }

    .menu-item:hover {
        background-color: #f5f5f5;
    }

    .mr-2 {
        margin-right: 0.5rem;
    }

    .user-menu-container button {
       width: 2.5rem;
       height: 2.5rem;
    }

    .user-menu-container button.fas::before {
        font-size: 1.5rem;
    }

    .imgNormalLogo {
        width: 15rem;
    }

</style>

<header>
    <a href="/" use:route>
        <img class="imgNormalLogo p-050" src="/images/ffs-logo.svg" alt="FlatFeeStack"/>
    </a>
    <nav>
        {#if appState.user.id}
            <a href="/user/search" use:route aria-label="Dashboard">

            </a>
            <div class="user-menu-container mx-100">
            {#if appState.user.image}
                <button onclick={toggleMenu} class="button2 user-menu-button" aria-label="User menu">
                    <img class="image-org-sx" src={appState.user.image} alt="user profile img" />
                </button>
            {:else}
                <button class="round fas fa-user" onclick={toggleMenu}></button>
            {/if}


                {#if isMenuOpen}
                    <div class="menu-dropdown">
                        <div class="menu-header">
                            <p class="user-email">{appState.user.email}</p>
                        </div>
                        <button class="menu-item" onclick={logout}>
                            <div class="menu-item-content">
                                <span>
                                <i class="fas fa-sign-out-alt mr-2"></i>
                                Sign out
                                </span>
                            </div>
                        </button>
                    </div>
                {/if}
            </div>
        {:else}
            <button class="button1 center mx-2 p-125" type="submit" onsubmit={() => goto("/login")}>Login</button>
        {/if}
        {#if appState.config.env === "local" || appState.config.env === "stage"}
            <button class="button4 center mx-2" onclick={() => goto("/dao/home")}
            >DAO</button
            >
        {/if}
    </nav>
</header>

{#if appState.error}
    <div class="bg-red p-2 parent err-container">
        <div class="w-100">{@html appState.error}</div>
        <div>
            <button class="close accessible-btn" onclick={() => {appState.error = "";}}>
                ✕
            </button>
        </div>
    </div>
{/if}
