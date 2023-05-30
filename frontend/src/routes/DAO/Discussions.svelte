<script lang="ts">
  import { onMount } from "svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import type { Post } from "../../types/forum";
  import { API } from "../../ts/api";
  import Spinner from "../../components/Spinner.svelte";
  import { navigate } from "svelte-routing";
  import DiscussionListItem from "../../components/DAO/DiscussionListItem.svelte";
  import { error } from "../../ts/mainStore";

  let isLoading = true;
  let posts: Post[] = [];

  onMount(async () => {
    try {
      posts = await API.forum.getAllPosts();
    } catch (e) {
      $error = e.message;
    }
    isLoading = false;
  });

  const navigateToCreateDiscussion = () => {
    navigate("/dao/createDiscussion");
  };
</script>

<Navigation>
  {#if isLoading}
    <Spinner />
  {:else}
    <h1 class="text-secondary-900">Discussions</h1>

    <p>
      If you have ideas for improvements or new features regarding the
      FlatFeeStack DAO or the platform but still need refinement to create a
      proposal, feel free to create a discussion thread.
    </p>

    <div class="container-col2 items-start mt-2 mb-20">
      <button class="button1" on:click={navigateToCreateDiscussion}
        >Start a new discussion
      </button>
    </div>

    {#if posts.length > 0}
      {#each posts as post (post.id)}
        <DiscussionListItem {post} />
      {/each}
    {:else}
      <p>It looks like nobody did start any discussion so far.</p>
    {/if}
  {/if}
</Navigation>
