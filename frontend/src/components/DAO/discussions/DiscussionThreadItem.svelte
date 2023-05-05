<script lang="ts">
  import { faPencil, faTrash } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import { user } from "../../../ts/mainStore";
  import { timeSince } from "../../../ts/services";
  import type { Comment, Post } from "../../../types/forum";
  import ThreadItemBox from "../ThreadItemBox.svelte";
  import { users } from "../../../ts/userStore";

  export let deleteItem: () => void;
  export let discussionOpen: boolean;
  export let editItem: () => void;
  export let item: Post | Comment;
</script>

<style>
  .mr-4 {
    margin-right: 0.5rem;
  }

  .image {
    border-radius: 9999px;
    height: 2rem;
    width: 2rem;
  }

  p {
    margin: 0.1rem;
  }
</style>

<ThreadItemBox>
  <div class="border-bottom flex justify-between">
    <div class="flex gap-3 items-center">
      {#await users.get(item.author)}
        ...
      {:then user}
        {#if user.image}
          <img
            class="image"
            src={user.image}
            alt={`Profile picture of ${user.name || "[unknown]"}`}
          />
        {:else}
          <div class="bg-green image" />
        {/if}
        <p class="bold">{user.name || "[unknown]"}</p>
      {/await}
    </div>
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
        <button
          class="accessible-btn"
          on:click={() => editItem()}
          title="Edit comment"
        >
          <Fa class="ml-4" icon={faPencil} />
        </button>
      {/if}

      {#if $user.role === "admin"}
        <button
          class="accessible-btn"
          on:click={() => deleteItem()}
          title="Delete item"
        >
          <Fa class="ml-4" icon={faTrash} />
        </button>
      {/if}
    </div>
  </div>

  <slot />
</ThreadItemBox>
