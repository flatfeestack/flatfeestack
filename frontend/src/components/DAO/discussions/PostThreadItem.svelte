<script lang="ts">
  import { navigate } from "svelte-routing";
  import { API } from "../../../ts/api";
  import type { Post } from "../../../types/forum";
  import DiscussionThreadItem from "./DiscussionThreadItem.svelte";
  import { error } from "../../../ts/mainStore";

  export let item: Post;

  function editItem() {
    navigate(`/dao/discussion/${item.id}/edit`);
  }

  async function deleteDiscussion() {
    try {
      await API.forum.deletePost(item.id);
      navigate("/dao/discussions");
    } catch (e) {
      $error = e.message;
    }
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
