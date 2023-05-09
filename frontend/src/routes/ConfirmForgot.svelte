<script lang="ts">
  import Dots from "../components/Dots.svelte";
  import { confirmReset } from "../ts/services";
  import { navigate, link } from "svelte-routing";

  export let email: string;
  export let token: string;

  let password = "";
  let error = "";
  let isSubmitting = false;

  async function handleSubmit() {
    try {
      error = "";
      isSubmitting = true;
      await confirmReset(email, password, token);
      navigate("/user/search");
      isSubmitting = false;
      email = "";
      password = "";
    } catch (e) {
      isSubmitting = false;
      error = e;
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
    <h2 class="py-5 text-center text-primary-900">Set your new password</h2>

    <form on:submit|preventDefault={handleSubmit}>
      <label for="email" class="py-1">Email address</label>
      <input
        required
        size="100"
        maxlength="100"
        type="email"
        id="email"
        name="email"
        bind:value={email}
        class="rounded py-2 border-primary-900"
      />
      <label for="password" class="flex py-1">Set new password</label>
      <input
        required
        size="100"
        maxlength="100"
        type="password"
        id="password"
        minlength="8"
        bind:value={password}
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

    <div class="divider" />
    <div class="flex">
      Already have an account?&nbsp;<a href="/login" use:link>Log in</a>
    </div>
  </div>
</div>
