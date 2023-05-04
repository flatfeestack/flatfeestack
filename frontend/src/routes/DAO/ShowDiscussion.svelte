<script lang="ts">
  import { onMount } from "svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import type { Comment, Post } from "../../types/forum";
  import { API } from "../../ts/api";
  import DiscussionThreadItem from "../../components/DAO/DiscussionThreadItem.svelte";
  import CreateComment from "../../components/DAO/CreateComment.svelte";

  export let postId: string;

  let isLoading = true;
  let post: Post;
  let comments: Comment[];

  onMount(async () => {
    [post, comments] = await Promise.all([
      API.forum.getPost(postId),
      API.forum.getAllComments(postId),
    ]);

    isLoading = false;
  });
</script>

<Navigation>
  {#if isLoading}
    <Spinner />
  {:else}
    <h1 class="text-secondary-900">{post.title}</h1>

    <DiscussionThreadItem item={post} />

    {#each comments as comment (comment.id)}
      <DiscussionThreadItem item={comment} />
    {/each}

    <CreateComment bind:comments postId={post.id} />
  {/if}
</Navigation>
