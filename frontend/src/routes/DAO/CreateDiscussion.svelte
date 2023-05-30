<script lang="ts">
  import { onMount } from "svelte";
  import { navigate } from "svelte-routing";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import PostForm from "../../components/DAO/discussions/PostForm.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import { API } from "../../ts/api";
  import { error, user } from "../../ts/mainStore";
  import { postSchema } from "../../ts/validationSchemas";
  import { getAllFormErrors } from "../../utils/validationHelpers";

  onMount(() => {
    if (Object.keys($user).length === 0) {
      navigate("/login");
    }
  });

  let formValues = {
    content: "",
    title: "",
  };

  let formErrors = {
    content: null,
    title: null,
  };

  let isSubmitting = false;

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

    try {
      const post = await API.forum.createPost(formValues);
      navigate(`/dao/discussion/${post.id}`);
    } catch (e) {
      $error = e.message;
    }
    isSubmitting = false;
  }
</script>

<Navigation>
  {#if isSubmitting}
    <Spinner />
  {:else}
    <h1 class="text-secondary-900">Create a new discussion</h1>

    <PostForm
      bind:formValues
      bind:formErrors
      {handleSubmit}
      submitButtonLabel="Create!"
    />
  {/if}
</Navigation>
