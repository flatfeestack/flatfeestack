<script lang="ts">
  import type { Post } from "../../../types/forum";
  import DiscussionThreadItem from "./DiscussionThreadItem.svelte";

  export let item: Post;

  function editItem() {
    // todo redirect to "edit post" page
  }
</script>

<DiscussionThreadItem {item} {editItem}>
  <p class="mb-2 mt-2">{item.content}</p>
</DiscussionThreadItem>
