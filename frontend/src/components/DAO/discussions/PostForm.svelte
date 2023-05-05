<script lang="ts">
  import { postSchema } from "../../../ts/validationSchemas";

  interface FormValues {
    content: string;
    title: string;
  }

  interface FormErrors {
    content: string | null;
    title: string | null;
  }

  export let handleSubmit: () => void;
  export let formValues: FormValues;
  export let formErrors: FormErrors;
  export let submitButtonLabel: string;

  async function validateAt(property: string) {
    try {
      await postSchema.validateAt(property, formValues);
      formErrors[property] = null;
    } catch (error) {
      formErrors[property] = error.errors[0];
    }
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
      >{submitButtonLabel}</button
    >
  </div>
</div>
