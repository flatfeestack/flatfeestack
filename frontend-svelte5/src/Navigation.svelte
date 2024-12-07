<script lang="ts">
  import '@fortawesome/fontawesome-free/css/all.min.css';
  import {route, goto} from "@mateothegreat/svelte5-router";
  import {appState} from "ts/state.svelte.ts";

  let windowWidth = $state(0);
  let isSidebarCollapsed = $state(false);

  function handleResize() {
    isSidebarCollapsed = windowWidth <= 768;
  }

  const navItems = [
    { icon: 'fa-user-cog', label: 'Settings', path: '/user/settings' },
    { icon: 'fa-search', label: 'Search', path: '/user/search' },
    { icon: 'fa-credit-card', label: 'Payments', path: '/user/payments' },
    { icon: 'fa-hand-holding-usd', label: 'Income', path: '/user/income' },
    { icon: 'fa-user-friends', label: 'Invitations', path: '/user/invitations' },
    { icon: 'fa-medal', label: 'Badges', path: '/user/badges' },
  ];

  if(appState.user.role === "admin") {
    navItems.push({ icon: 'fa-shield-alt', label: 'Admin', path: '/user/badadminges' });
  }

</script>

<svelte:window
        bind:innerWidth={windowWidth}
        on:resize={handleResize}
/>

<style>
  .sidebar {
    background-color: black;
    height: 100vh;
    transition: width 0.3s ease;
    position: sticky;
    left: 0;
    top: 0;
    overflow-y: auto;
  }

  .sidebar-header {
    padding: 1rem;
    color: white;
  }

  .nav-item {
    color: white;
    padding: 0.75rem 1rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    cursor: pointer;
    text-decoration: none;
    border: none;
    background: none;
    width: 100%;
    text-align: left;
  }

  .nav-item:hover {
    background-color: #333;
  }

  @media (max-width: 768px) {
    .sidebar {
      width: 4rem;
    }
  }
</style>

<nav class="sidebar {isSidebarCollapsed ? 'sidebar-collapsed' : 'sidebar-expanded'}">
  <div class="sidebar-header center-flex">
    <a class="center-flex" href="/" use:route>
      <img src="/images/ffs-logo-white.svg" alt="FlatFeeStack"/>
    </a>
  </div>

  <div class="nav-items">
    {#each navItems as {icon, label, path}}
      <button class="nav-item" onclick={() => goto(path)} aria-label={label}>
        <i class="fas {icon}"></i>
        {#if !isSidebarCollapsed}
          <span>{label}</span>
        {/if}
      </button>
    {/each}
  </div>
</nav>
