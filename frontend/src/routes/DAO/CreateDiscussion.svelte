<script lang="ts">
  import { onMount } from "svelte";
  import { navigate } from "svelte-routing";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import Spinner from "../../components/Spinner.svelte";
  import { API } from "../../ts/api";
  import { user } from "../../ts/mainStore";
  import { getAllFormErrors } from "../../utils/validationHelpers";
  import yup from "../../utils/yup";

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

  const schema = yup.object().shape({
    content: yup.string().min(1).max(1000).required(),
    title: yup.string().min(1).max(100).required(),
  });

  async function validateAt(property: string) {
    try {
      await schema.validateAt(property, formValues);
      formErrors[property] = null;
    } catch (error) {
      formErrors[property] = error.errors[0];
    }
  }

  async function handleSubmit() {
    isSubmitting = true;

    const validationErrors = await getAllFormErrors(formValues, schema);
    if (validationErrors.length > 0) {
      validationErrors.forEach((error) => {
        formErrors[error.path] = error.message;
      });
      isSubmitting = false;
      return;
    }

    const post = await API.forum.createPost(formValues);
    isSubmitting = false;
    navigate(`/dao/discussion/${post.id}`);
  }
</script>

<style>
  button {
    width: fit-content;
  }

  input {
    display: block;
  }
</style>

<Navigation>
  {#if isSubmitting}
    <Spinner />
  {:else}
    <h1 class="text-secondary-900">Create a new discussion</h1>

    <div class="container-col2 my-2">
      <label class="bold" for="title">Title</label>
    </div>

    <div class="container-col2 my-2">
      <input
        type="text"
        name="title"
        id="title"
        bind:value={formValues.title}
        on:blur={() => validateAt("title")}
      />

      {#if formErrors.title !== null}
        <p class="invalid" style="color:red">{formErrors.title}</p>
      {/if}
    </div>

    <div class="container-col2 my-2">
      <label class="bold" for="content">Content</label>
    </div>

    <div class="container-col2 my-2">
      <textarea
        class="box-sizing-border"
        name="content"
        id="content"
        bind:value={formValues.content}
        on:blur={() => validateAt("content")}
        rows="10"
        cols="50"
      />

      {#if formErrors.content !== null}
        <p class="invalid" style="color:red">{formErrors.content}</p>
      {/if}

      <div class="container-col2 items-end mt-2">
        <button class="button1" type="submit" on:click={() => handleSubmit()}
          >Create!</button
        >
      </div>
    </div>
  {/if}
</Navigation>
