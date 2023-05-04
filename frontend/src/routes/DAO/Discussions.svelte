<script lang="ts">
  import { onMount } from "svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import type { Post } from "../../types/forum";
  import { API } from "../../ts/api";
  import Spinner from "../../components/Spinner.svelte";
  import { navigate } from "svelte-routing";

  let isLoading = true;
  let posts: Post[] = [];

  onMount(async () => {
    posts = await API.forum.getAllPosts();
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
      FlateFeeStack DAO or the platform but still need refinement to create a
      proposal, feel free to create a discussion thread.
    </p>

    <hr />

    {#if posts.length > 0}
      Display the posts
    {:else}
      <p>It looks like nobody did start any discussion so far.</p>

      <button class="button1" on:click={navigateToCreateDiscussion}
        >Start a new discussion</button
      >
    {/if}
  {/if}
</Navigation>
