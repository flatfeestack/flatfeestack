<script lang="ts">
  import { onMount } from "svelte";
  import CreateComment from "../../components/DAO/CreateComment.svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import CommentThreadItem from "../../components/DAO/discussions/CommentThreadItem.svelte";
  import PostThreadItem from "../../components/DAO/discussions/PostThreadItem.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import { API } from "../../ts/api";
  import type { Comment, Post } from "../../types/forum";

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

    <PostThreadItem item={post} />

    {#each comments as comment (comment.id)}
      <CommentThreadItem bind:item={comment} {postId} />
    {/each}

    <CreateComment bind:comments postId={post.id} />
  {/if}
</Navigation>
