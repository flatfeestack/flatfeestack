<script lang="ts">
  import { navigate } from "svelte-routing";
  import type { Post } from "../../types/forum";
  import formatDateTime from "../../utils/formatDateTime";

  export let post: Post;
</script>

<style>
  .align-middle {
    align-items: center;
  }

  .border-bottom {
    border-bottom: 1px solid var(--secondary-300);
  }

  .pb-4 {
    padding-bottom: 1rem;
  }

  button {
    height: fit-content;
    width: fit-content;
  }

  p {
    margin: 0.25rem;
  }

  span {
    border-radius: 4px;
    display: inline-block;
    font-size: 0.8rem;
    margin-left: 0.25rem;
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

<div
  class="align-middle border-bottom flex font-size-small justify-between pb-4"
>
  <div>
    <p class="bold">
      {post.title}
      {#if post.open}<span class="open">Open</span>{:else}<span class="closed"
          >Closed</span
        >{/if}
    </p>
    <p>
      Created on {formatDateTime(new Date(post.created_at))} by {post.author}
    </p>
  </div>

  <button
    class="button3"
    on:click={() => navigate(`/dao/discussion/${post.id}`)}
    >View discussion</button
  >
</div>
