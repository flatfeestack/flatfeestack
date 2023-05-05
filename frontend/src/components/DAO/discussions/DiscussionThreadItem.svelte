<script lang="ts">
  import Fa from "svelte-fa";
  import ThreadItemBox from "../ThreadItemBox.svelte";
  import type { Comment, Post } from "../../../types/forum";
  import { faPencil } from "@fortawesome/free-solid-svg-icons";
  import { user } from "../../../ts/mainStore";
  import { timeSince } from "../../../ts/services";

  export let discussionOpen: boolean;
  export let editItem: () => void;
  export let item: Post | Comment;
</script>

<style>
  .mr-4 {
    margin-right: 0.5rem;
  }

  p {
    margin: 0.1rem;
  }
</style>

<ThreadItemBox>
  <div class="border-bottom flex justify-between">
    <p class="bold">{item.author}</p>
    <div class="color-secondary-500 flex gap-3 items-center mr-4">
      <div>
        <p>
          {timeSince(new Date(item.created_at), new Date())} ago
        </p>
        {#if item.updated_at}
          <p>(edited)</p>
        {/if}
      </div>
      {#if item.author === $user.id && discussionOpen}
        <button class="accessible-btn" on:click={() => editItem()}>
          <Fa class="ml-4" icon={faPencil} />
        </button>
      {/if}
    </div>
  </div>

  <slot />
</ThreadItemBox>
