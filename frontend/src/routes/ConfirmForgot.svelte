<script lang="ts">
  import { Link, navigate } from "svelte-routing";
  import Dots from "../components/Dots.svelte";
  import { confirmReset, updateUser } from "./../ts/authService";

  export let email;
  export let token;

  let password = "";
  let error = "";
  let isSubmitting = false;

  async function handleSubmit() {
    try {
      error = "";
      isSubmitting = true;
      await confirmReset(email, password, token);
      await updateUser();
      navigate("/dashboard");
      isSubmitting = false;
      email = "";
      password = "";
    } catch (e) {
      isSubmitting = false;
      console.log(e);
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
    <h2 class="py-5 text-center text-primary-700">Create your account</h2>

    <form on:submit|preventDefault="{handleSubmit}">
      <label for="email" class="py-1">Email address</label>
      <input required size="100" maxlength="100" type="email" id="email" name="email" bind:value={email} class="rounded py-2 border-primary-700" />
      <label for="password" class="flex py-1">Set new password</label>
      <input required size="100" maxlength="100" type="password" id="password" minlength="8" bind:value={password} class="rounded py-2 border-primary-700"/>
      <button class="btn my-4" disabled="{isSubmitting}" type="submit">Reset password
        {#if isSubmitting}<Dots />{/if}
      </button>

      {#if error}
        <div class="bg-red rounded p-2">{error}</div>
      {/if}

    </form>

    <div class="divider"></div>
    <div class="flex">
      Already have an account?&nbsp;<Link to="/signin">Log in</Link>
    </div>

  </div>
</div>
