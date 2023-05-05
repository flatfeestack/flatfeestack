<script lang="ts">
  import { onMount } from "svelte";
  import { user } from "../../ts/mainStore";
  import { navigate } from "svelte-routing";
  import { API } from "../../ts/api";
  import { getAllFormErrors } from "../../utils/validationHelpers";
  import { postSchema } from "../../ts/validationSchemas";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import PostForm from "../../components/DAO/discussions/PostForm.svelte";

  export let postId: string;

  let formValues = {
    content: "",
    title: "",
  };

  let formErrors = {
    content: null,
    title: null,
  };

  let isSubmitting = true;

  onMount(async () => {
    if (Object.keys($user).length === 0) {
      navigate("/login");
    }

    const post = await API.forum.getPost(postId);
    formValues.content = post.content;
    formValues.title = post.title;
    isSubmitting = false;
  });

  async function handleSubmit() {
    isSubmitting = true;

    const validationErrors = await getAllFormErrors(formValues, postSchema);
    if (validationErrors.length > 0) {
      validationErrors.forEach((error) => {
        formErrors[error.path] = error.message;
      });
      isSubmitting = false;
      return;
    }

    const post = await API.forum.updatePost(postId, formValues);
    isSubmitting = false;
    navigate(`/dao/discussion/${post.id}`);
  }
</script>

<Navigation>
  {#if isSubmitting}
    <Spinner />
  {:else}
    <h1 class="text-secondary-900">Edit discussion {formValues.title}</h1>

    <PostForm
      bind:formValues
      bind:formErrors
      {handleSubmit}
      submitButtonLabel="Update!"
    />
  {/if}
</Navigation>
