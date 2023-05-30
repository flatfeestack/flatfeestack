<script lang="ts">
  import { Link } from "svelte-routing";
  import { API } from "../../ts/api";
  import { error, user } from "../../ts/mainStore";
  import type { Comment } from "../../types/forum";
  import { getAllFormErrors } from "../../utils/validationHelpers";
  import yup from "../../utils/yup";
  import ThreadItemBox from "./ThreadItemBox.svelte";

  export let postId: string;
  export let comments: Comment[];

  let formValues = {
    content: "",
  };

  let formErrors = {
    content: null,
  };

  let isSubmitting = false;
  let isLoggedIn = false;

  const schema = yup.object().shape({
    content: yup.string().min(1).max(500).required(),
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

    try {
      const comment = await API.forum.createComment(postId, formValues);
      comments = [...comments, comment];
      formValues = {
        content: "",
      };
    } catch (e) {
      $error = e.message;
    }
    isSubmitting = false;
  }

  $: {
    isLoggedIn = Object.keys($user).length !== 0;
  }
</script>

<ThreadItemBox>
  {#if isLoggedIn}
    {#if isSubmitting}
      Creating comment ...
    {:else}
      <div class="container-col2 my-2">
        <label class="bold" for="content">Add a new comment</label>
      </div>

      <div class="container-col2 my-2">
        <textarea
          class="box-sizing-border"
          name="content"
          id="content"
          bind:value={formValues.content}
          on:blur={() => validateAt("content")}
          rows="5"
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
  {:else}
    You need to <Link href="/login">sign in</Link> to post a comment.
  {/if}
</ThreadItemBox>
