<script lang="ts">
  import Header from "./Header.svelte";
  import "./styles.css";

  let error: string | null = "";
</script>

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  .close {
    cursor: pointer;
    text-align: right;
  }

  .err-container {
    display: flex;
    flex-direction: row;
  }
  footer {
    background-color: #000;
    color: white;
    flex: 0 0 auto;
    font-size: 1rem;
    padding: 0.5rem;
  }

  footer > :global(a) {
    color: white;
    font-size: 1rem;
  }

  @media (min-width: 480px) {
    footer {
      padding: 12px 0;
    }
  }
</style>

<div class="app">
  <Header />
  {#if error}
    <div class="bg-red p-2 parent err-container">
      <div class="w-100">{@html error}</div>
      <div>
        <button
          class="close accessible-btn"
          on:click|preventDefault={() => {
            error = null;
          }}
        >
          ✕
        </button>
      </div>
    </div>
  {/if}
  <main>
    <slot />
  </main>

  <footer class="text-center">
    We used the following <a href="/stats.html">dependencies</a>
  </footer>
</div>
