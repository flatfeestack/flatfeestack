<script lang="ts">
  import { timeSince } from "../../ts/services";
  import type { Comment, Post } from "../../types/forum";
  import ThreadItemBox from "./ThreadItemBox.svelte";

  export let item: Post | Comment;
</script>

<style>
  .border-bottom {
    border-bottom: 1px solid var(--secondary-300);
  }

  .mr-4 {
    margin-right: 1rem;
  }
</style>

<ThreadItemBox>
  <div class="border-bottom flex justify-between">
    <p class="bold">{item.author}</p>
    <p class="color-secondary-500 mr-4">
      {timeSince(new Date(item.created_at), new Date())} ago
    </p>
  </div>

  <p class="mb-2 mt-2">{item.content}</p>
</ThreadItemBox>
