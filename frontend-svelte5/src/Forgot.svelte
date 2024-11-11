<script lang="ts">
  import {route} from "@mateothegreat/svelte5-router";
  import { API } from "@/api";
  import Dots from "./Dots.svelte";
  import { emailValidationPattern } from "@/utils";
  let email = "";
  let error = "";
  let isSubmitting = false;
  let info = "";

  async function handleSubmit() {
    try {
      error = "";
      isSubmitting = true;
      const res = await API.auth.reset(email);
      isSubmitting = false;
      email = "";
      info =
        "Your email is on the way. To enable your account, click on the link in the email.";
    } catch (e) {
      isSubmitting = false;
      error = "Something went wrong. Please try again.";
    }
  }
</script>

<style>
  button,
  input:focus {
    outline: none;
  }
  input:required {
    box-shadow: none;
  }

  label {
    color: var(--primary-900);
  }

  form {
    display: flex;
    flex-direction: column;
  }

  .max {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .flex {
    padding-top: 1em;
    display: flex;
    justify-content: space-between;
  }
</style>

<div class="max">
  <div class="box-container rounded p-5">
    <h2 class="py-5 text-center text-primary-900">Reset your password</h2>

    {#if info}
      <div class="bg-green rounded p-2">{info}</div>
    {:else}
      <form on:submit|preventDefault={handleSubmit}>
        <label for="email" class="py-1">Email address</label>
        <input
          required
          size="100"
          maxlength="100"
          type="email"
          pattern={emailValidationPattern}
          id="email"
          name="email"
          bind:value={email}
          class="rounded py-2 border-primary-900"
        />
        <button class="button1 my-4" disabled={isSubmitting} type="submit"
          >Reset password
          {#if isSubmitting}<Dots />{/if}
        </button>

        {#if error}
          <div class="bg-red rounded p-2">{error}</div>
        {/if}
      </form>
    {/if}

    <div class="divider" />
    <div class="flex">
      Already have an account?&nbsp;<a href="/login" use:route>Log in</a>
    </div>
  </div>
</div>
