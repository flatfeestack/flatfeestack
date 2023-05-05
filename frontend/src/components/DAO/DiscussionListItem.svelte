<script lang="ts">
  import { navigate } from "svelte-routing";
  import type { Post } from "../../types/forum";
  import formatDateTime from "../../utils/formatDateTime";
  import StatusSpan from "./discussions/StatusSpan.svelte";

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
</style>

<div
  class="align-middle border-bottom flex font-size-small justify-between pb-4"
>
  <div>
    <p class="bold">
      {post.title}
      <StatusSpan {post} />
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
