<script lang="ts">
  import { navigate } from "svelte-routing";
  import { API } from "../../../ts/api";
  import type { Post } from "../../../types/forum";
  import DiscussionThreadItem from "./DiscussionThreadItem.svelte";

  export let item: Post;

  function editItem() {
    navigate(`/dao/discussion/${item.id}/edit`);
  }

  async function deleteDiscussion() {
    await API.forum.deletePost(item.id);
    navigate("/dao/discussions");
  }
</script>

<style>
  p {
    white-space: pre-line;
  }
</style>

<DiscussionThreadItem
  {item}
  {editItem}
  discussionOpen={item.open}
  deleteItem={deleteDiscussion}
>
  <p class="mb-2 mt-2">{item.content}</p>
</DiscussionThreadItem>
