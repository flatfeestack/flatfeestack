<script lang="ts">
  import { Link } from "svelte-routing";
  import { navigate } from "svelte-routing";
  import { login } from "./../ts/authService";
  import Dots from "../components/Dots.svelte";

  let email = "";
  let password = "";
  let error = "";
  let isSubmitting = false;

  async function handleSubmit() {
    try {
      error = "";
      isSubmitting = true;
      await login(email, password);
      isSubmitting = false;
      navigate("/dashboard");
      email = "";
      password = "";
    } catch (e) {
      isSubmitting = false;
      password = "";
      if (e?.response?.status === 400) {
        error = "No match found for username / password combination";
      } else {
        error = "Something went wrong. Please try again.";
      }
    }
  }
</script>

<style>
    .container {
        margin-top: 2em;
        max-width: 20rem;
        background-color: var(--primary-100);
    }

    button, input:focus{
        outline: none;
    }
    input:required {
        box-shadow: none;
    }

    label {
        color: var(--primary-700);
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
  <div class="container rounded p-5">
    <h2 class="py-5 text-center text-primary-700">Sign in to flatfeestack</h2>

    <form on:submit|preventDefault="{handleSubmit}">
      <label for="email" class="py-1">Email address</label>
      <input required size="100" maxlength="100" type="email" id="email" name="email" bind:value={email} class="rounded py-2 border-primary-700" />

      <div class="flex py-1">
        <label for="password" class="">Password</label>
        <label for="password" class="">
          <Link to="/forgot">Forgot password?</Link>
        </label>
      </div>

      <input required size="100" maxlength="100" type="password" id="password" minlength="8" bind:value={password} class="rounded py-2 border-primary-700"/>
      <button class="btn my-4" disabled="{isSubmitting}" type="submit">Sign in
        {#if isSubmitting}<Dots />{/if}
      </button>

      {#if error}
        <div class="bg-red rounded p-2">{error}</div>
      {/if}

    </form>

    <div class="divider"></div>
    <div class="flex">
      New to flatfeestack?&nbsp;<Link to="/signup">Sign up</Link>
    </div>

  </div>
</div>
