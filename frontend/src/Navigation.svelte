<script lang="ts">
    import { link, navigate } from "preveltekit";
    import { appState } from "ts/state.svelte.ts";
    import { library, icon } from "@fortawesome/fontawesome-svg-core";
    import {
        faSearch,
        faCreditCard,
        faHandHoldingUsd,
        faUserFriends,
        faMedal,
        faShieldAlt,
        faAngleRight,
        faAngleLeft,
    } from "@fortawesome/free-solid-svg-icons";

    library.add(
        faSearch,
        faCreditCard,
        faHandHoldingUsd,
        faUserFriends,
        faMedal,
        faShieldAlt,
        faAngleRight,
        faAngleLeft,
    );

    // Types
    interface NavItem {
        icon: string;
        label: string;
        path: string;
    }

    let windowWidth = $state(0);
    let isSidebarCollapsed = $state(false);
    let manualOverride = $state(false);

    function handleResize() {
        const shouldBeCollapsed = windowWidth <= 768;
        if (manualOverride && isSidebarCollapsed === shouldBeCollapsed) {
            manualOverride = false;
        }
        if (!manualOverride) {
            isSidebarCollapsed = shouldBeCollapsed;
        }
    }

    function toggleSidebar() {
        manualOverride = true;
        isSidebarCollapsed = !isSidebarCollapsed;
    }

    function getIcon(iconName: string) {
        const iconMap: Record<string, string> = {
            "fa-search": "search",
            "fa-credit-card": "credit-card",
            "fa-hand-holding-usd": "hand-holding-usd",
            "fa-user-friends": "user-friends",
            "fa-medal": "medal",
            "fa-shield-alt": "shield-alt",
        };
        return icon({ prefix: "fas", iconName: iconMap[iconName] }).html[0];
    }

    const navItems: NavItem[] = [
        { icon: "fa-search", label: "Search", path: "/user/search" },
        { icon: "fa-credit-card", label: "Payments", path: "/user/payments" },
        { icon: "fa-hand-holding-usd", label: "Income", path: "/user/income" },
        {
            icon: "fa-user-friends",
            label: "Invitations",
            path: "/user/invitations",
        },
        { icon: "fa-medal", label: "Badges", path: "/user/badges" },
    ];

    const navItemsAdmin: NavItem[] = [
        { icon: "fa-shield-alt", label: "Admin", path: "/user/admin" },
        {
            icon: "fa-shield-alt",
            label: "Healthy Repos",
            path: "/user/healthy-repos",
        },
        {
            icon: "fa-shield-alt",
            label: "Repo Assessment",
            path: "/user/healthy-repo-assessment",
        },
    ];

    const angleRightIcon = icon({ prefix: "fas", iconName: "angle-right" })
        .html[0];
    const angleLeftIcon = icon({ prefix: "fas", iconName: "angle-left" })
        .html[0];
</script>

<svelte:window bind:innerWidth={windowWidth} on:resize={handleResize} />

<link rel="preload" href="/images/ffs-logo-min-white.svg" as="image" />
<link rel="preload" href="/images/ffs-logo-white.svg" as="image" />

<nav class="sidebar">
    <div class="sidebar-header center-flex">
        <a href="/" use:link>
            <img
                src={isSidebarCollapsed
                    ? "/images/ffs-logo-min-white.svg"
                    : "/images/ffs-logo-white.svg"}
                alt="FlatFeeStack"
            />
        </a>
        <button
            class="toggle-button pr-025 pl-050"
            onclick={toggleSidebar}
            aria-label="Toggle sidebar"
        >
            {@html isSidebarCollapsed ? angleRightIcon : angleLeftIcon}
        </button>
    </div>
    <div class="nav-items">
        {#each navItems as { icon, label, path }}
            <button
                class="nav-item"
                onclick={() => navigate(path)}
                aria-label={label}
            >
                {@html getIcon(icon)}
                {isSidebarCollapsed ? " " : label}
            </button>
        {/each}
        {#if appState.user.role === "admin"}
            {#each navItemsAdmin as { icon, label, path }}
                <button
                    class="nav-item"
                    onclick={() => navigate(path)}
                    aria-label={label}
                >
                    {@html getIcon(icon)}
                    {isSidebarCollapsed ? " " : label}
                </button>
            {/each}
        {/if}
    </div>
</nav>

<style>
    .sidebar {
        color: white;
        background-color: black;
        height: 100vh;
        transition: width 0.3s ease;
        position: sticky;
        left: 0;
        top: 0;
        overflow-y: auto;
    }
    .sidebar-header {
        padding: 1rem 0 1rem 1rem;
        display: flex;
        justify-content: space-between;
    }
    .nav-item {
        padding: 0.75rem 1rem 1rem 1rem;
        display: flex;
        gap: 0.75rem;
        cursor: pointer;
        text-decoration: none;
        border: none;
        background: none;
        width: 100%;
        text-align: left;
    }
    button:hover {
        background-color: #333;
    }
    .toggle-button {
        cursor: pointer;
        background: none;
    }
</style>
