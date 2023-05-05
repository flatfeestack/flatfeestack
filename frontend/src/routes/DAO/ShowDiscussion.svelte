<script lang="ts">
  import { onMount } from "svelte";
  import CreateComment from "../../components/DAO/CreateComment.svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import CommentThreadItem from "../../components/DAO/discussions/CommentThreadItem.svelte";
  import PostThreadItem from "../../components/DAO/discussions/PostThreadItem.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import { API } from "../../ts/api";
  import type { Comment, Post } from "../../types/forum";
  import { user } from "../../ts/mainStore";
  import StatusSpan from "../../components/DAO/discussions/StatusSpan.svelte";

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

  async function closeDiscussion() {
    await API.forum.closePost(postId);
    post.open = false;
  }
</script>

<style>
  button {
    height: fit-content;
    width: fit-content;
  }
</style>

<Navigation>
  {#if isLoading}
    <Spinner />
  {:else}
    <div class="flex items-center justify-between">
      <div class="flex items-center">
        <h1 class="text-secondary-900">
          {post.title}
        </h1>
        <StatusSpan {post} />
      </div>
      {#if $user.id === post.author && post.open}
        <button class="button3" on:click={() => closeDiscussion()}
          >Close discussion</button
        >
      {/if}
    </div>

    <PostThreadItem item={post} />

    {#each comments as comment (comment.id)}
      <CommentThreadItem
        bind:item={comment}
        {postId}
        discussionOpen={post.open}
      />
    {/each}

    {#if post.open}
      <CreateComment bind:comments postId={post.id} />
    {/if}
  {/if}
</Navigation>
