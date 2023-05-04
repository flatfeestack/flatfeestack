<style>
  .border-secondary-300 {
    border: 1px solid var(--secondary-300);
  }

  .shadow-md {
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  }
</style>

<div class="border-secondary-300 font-size-small mb-5 p-2 rounded shadow-md">
  <slot />
</div>
