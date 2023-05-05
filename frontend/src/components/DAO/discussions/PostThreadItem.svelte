<script lang="ts">
  import { navigate } from "svelte-routing";
  import type { Post } from "../../../types/forum";
  import DiscussionThreadItem from "./DiscussionThreadItem.svelte";

  export let item: Post;

  function editItem() {
    navigate(`/dao/discussion/${item.id}/edit`);
  }
</script>

<style>
  p {
    white-space: pre-line;
  }
</style>

<DiscussionThreadItem {item} {editItem} discussionOpen={item.open}>
  <p class="mb-2 mt-2">{item.content}</p>
</DiscussionThreadItem>
