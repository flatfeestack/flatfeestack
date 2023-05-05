<script lang="ts">
  import { API } from "../../../ts/api";
  import { commentSchema } from "../../../ts/validationSchemas";
  import type { Comment } from "../../../types/forum";
  import { getAllFormErrors } from "../../../utils/validationHelpers";
  import DiscussionThreadItem from "./DiscussionThreadItem.svelte";

  export let discussionOpen: boolean;
  export let item: Comment;
  export let postId: string;

  let editMode = false;
  let formValues = {
    content: item.content,
  };

  let formErrors = {
    content: null,
  };

  let isSubmitting = false;

  async function validateAt(property: string) {
    try {
      await commentSchema.validateAt(property, formValues);
      formErrors[property] = null;
    } catch (error) {
      formErrors[property] = error.errors[0];
    }
  }

  function cancelEdit() {
    editMode = false;
  }

  function editItem() {
    editMode = true;
  }

  async function handleSubmit() {
    isSubmitting = true;

    const validationErrors = await getAllFormErrors(formValues, commentSchema);
    if (validationErrors.length > 0) {
      validationErrors.forEach((error) => {
        formErrors[error.path] = error.message;
      });

      isSubmitting = false;
      return;
    }

    item = await API.forum.updateComment(postId, item.id, formValues);

    editMode = false;
    isSubmitting = false;
  }
</script>

<style>
  p {
    white-space: pre-line;
  }
</style>

<DiscussionThreadItem {item} {editItem} {discussionOpen}>
  {#if editMode}
    {#if isSubmitting}
      Updating comment, one moment please ...
    {:else}
      <div class="container-col2 my-2">
        <label class="bold" for="content">Edit comment</label>
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

        <div class="flex gap-3 justify-end mt-2">
          <button class="button3" type="submit" on:click={() => cancelEdit()}
            >Cancel</button
          >

          <button class="button1" type="submit" on:click={() => handleSubmit()}
            >Update!</button
          >
        </div>
      </div>
    {/if}
  {:else}
    <p class="mb-2 mt-2">{item.content}</p>
  {/if}
</DiscussionThreadItem>
