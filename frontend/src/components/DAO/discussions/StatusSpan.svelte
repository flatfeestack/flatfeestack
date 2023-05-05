<script lang="ts">
  import type { Post } from "../../../types/forum";

  export let post: Post;
</script>

<style>
  span {
    border-radius: 4px;
    display: inline-block;
    font-size: 0.8rem;
    margin-left: 0.5rem;
    padding: 0.15rem;
  }

  span.open {
    border: var(--primary-700) 1px solid;
    background-color: var(--primary-300);
    color: var(--primary-700);
  }

  span.closed {
    border: var(--secondary-700) 1px solid;
    background-color: var(--secondary-300);
    color: var(--secondary-700);
  }
</style>

{#if post.open}<span class="bold open">Open</span>{:else}<span
    class="bold closed">Closed</span
  >{/if}
