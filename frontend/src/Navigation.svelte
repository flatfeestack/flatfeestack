<script lang="ts">
  import '@fortawesome/fontawesome-free/css/all.min.css';
  import {route, goto} from "@mateothegreat/svelte5-router";
  import {appState} from "ts/state.svelte.ts";

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

  const navItems: NavItem[] = [
    { icon: 'fa-search', label: 'Search', path: '/user/search' },
    { icon: 'fa-credit-card', label: 'Payments', path: '/user/payments' },
    { icon: 'fa-hand-holding-usd', label: 'Income', path: '/user/income' },
    { icon: 'fa-user-friends', label: 'Invitations', path: '/user/invitations' },
    { icon: 'fa-medal', label: 'Badges', path: '/user/badges' },
  ];

  const navItemsAdmin: NavItem[] = [
    { icon: 'fa-shield-alt', label: 'Admin', path: '/user/admin' },
    { icon: 'fa-shield-alt', label: 'Healthy Repos', path: '/user/healthy-repos' },
    { icon: 'fa-shield-alt', label: 'Repo Assessment', path: '/user/healthy-repo-assessment' },
  ];
</script>

<svelte:window bind:innerWidth={windowWidth} on:resize={handleResize}/>

<link rel="preload" href="/images/ffs-logo-min-white.svg" as="image">
<link rel="preload" href="/images/ffs-logo-white.svg" as="image">

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

<nav class="sidebar">
  <div class="sidebar-header center-flex">
    <a href="/" use:route>
      <img src={isSidebarCollapsed ? "/images/ffs-logo-min-white.svg" : "/images/ffs-logo-white.svg"} alt="FlatFeeStack"/>
    </a>
    <button class="toggle-button pr-025 pl-050" onclick={toggleSidebar} aria-label="Toggle sidebar">
      <i class="fas {isSidebarCollapsed ? 'fa-angle-right': 'fa-angle-left'}"></i>
    </button>
  </div>

  <div class="nav-items">
    {#each navItems as {icon, label, path}}
      <button class="nav-item" onclick={() => goto(path)} aria-label={label}>
        <i class="fas {icon}"></i>
        {isSidebarCollapsed ? " " : label}
      </button>
    {/each}
    {#if appState.user.role === "admin"}
      {#each navItemsAdmin as {icon, label, path}}
        <button class="nav-item" onclick={() => goto(path)} aria-label={label}>
          <i class="fas {icon}"></i>
          {isSidebarCollapsed ? " " : label}
        </button>
      {/each}
    {/if}
  </div>
</nav>
