<script lang="ts">
  import { timeSince } from "../../ts/services";
  import type { Comment, Post } from "../../types/forum";

  export let item: Post | Comment;
</script>

<style>
  .border-secondary-300 {
    border: 1px solid var(--secondary-300);
  }

  .border-bottom {
    border-bottom: 1px solid var(--secondary-300);
  }

  .mr-4 {
    margin-right: 1rem;
  }

  .shadow-md {
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  }
</style>

<div class="border-secondary-300 font-size-small p-2 rounded shadow-md">
  <div class="border-bottom flex justify-between">
    <p class="bold">{item.author}</p>
    <p class="color-secondary-500 mr-4">
      {timeSince(new Date(item.created_at), new Date())} ago
    </p>
  </div>

  <p class="mb-2 mt-2">{item.content}</p>
</div>
