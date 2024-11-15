<script lang="ts">
    import { appState } from "ts/state.ts";
    import {hasToken, removeSession} from "ts/auth";
    import { onMount } from "svelte";
    import { API } from "./ts/api.ts";
    import '@fortawesome/fontawesome-free/css/all.min.css'
    import {route, goto} from "@mateothegreat/svelte5-router";

    function logout() {
        removeSession();
        goto("/login");
    }

    onMount(async () => {
        try {
            if(!appState.$state.user && hasToken()) {
                appState.$state.user = await API.user.get();
            }
        } catch (e) {
            appState.setError(e);
        }
    });
</script>

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

    header :global(a),
    header :global(a:visited),
    header :global(a:active) {
        text-decoration: none;
        color: #000;
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
    .imgNormalLogo {
        padding-right: 0.25em;
        width: 10rem;
    }
    nav p {
        margin: 0;
    }
</style>

<header>
    <a href="/" use:route>
        <img
                class="hide-sx imgNormalLogo"
                src="/images/ffs-logo.svg"
                alt="FlatFeeStack"
        />
    </a>
    <nav>
        {#if appState.$state.user.id}
            <a href="/user/search" use:route aria-label="Dashboard">
                <i class="fa-ome icon" ></i>
            </a>
            {#if appState.$state.user.image}
                <img class="image-org-sx" src={appState.$state.user.image} alt="user profile img" />
            {/if}
            <p class="hide-sx">{appState.$state.user.email}</p>
            <form on:submit|preventDefault={logout}>
                <button class="button3 center mx-2" type="submit">Sign out</button>
            </form>
        {:else}
            <form on:submit|preventDefault={() => goto("/login")}>
                <button class="button3 center mx-2" type="submit">Login</button>
            </form>
            <form on:submit|preventDefault={() => goto("/signup")}>
                <button class="button1 center mx-2" type="submit">Sign Up</button>
            </form>
        {/if}
        {#if appState.$state.config.env === "local" || appState.$state.config.env === "stage"}
            <button class="button4 center mx-2" on:click={() => goto("/dao/home")}
            >DAO</button
            >
        {/if}
    </nav>
</header>

{#if appState.$state.error}
    <div class="bg-red p-2 parent err-container">
        <div class="w-100">{@html appState.$state.error}</div>
        <div>
            <button class="close accessible-btn" on:click|preventDefault={() => {appState.$state.error = "";}}>
                ✕
            </button>
        </div>
    </div>
{/if}
