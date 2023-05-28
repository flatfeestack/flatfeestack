<script lang="ts">
  import { link } from "svelte-routing";
  import { API } from "./../ts/api";
  import Dots from "../components/Dots.svelte";
  import { removeToken } from "../ts/services";
  import { emailValidationPattern } from "../ts/utils";

  let email = "";
  let password = "";
  let confirmPassword = "";
  let error = "";
  let info = "";
  let isSubmittingSignup = false;

  async function handleSubmit() {
    if (password !== confirmPassword) {
      error = "Password and confirmation password do not match.";
      return;
    }
    try {
      error = "";
      isSubmittingSignup = true;
      await API.auth.signup(email, password);
      removeToken();
      email = "";
      password = "";
      confirmPassword = "";
      info =
        "Your email is on the way. To enable your account, click on the link in the email.";
    } catch (e) {
      error = "Something went wrong. Please try again.";
    } finally {
      isSubmittingSignup = false;
    }
  }
</script>

<style>
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
    <h2 class="py-5 text-center text-primary-900">Create your account</h2>

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
        />
        <label for="password" class="flex py-1">Password</label>
        <input
          required
          size="100"
          maxlength="100"
          type="password"
          id="password"
          minlength="8"
          bind:value={password}
        />
        <label for="confirmPassword" class="flex py-1">Confirm Password</label>
        <input
          required
          size="100"
          maxlength="100"
          type="password"
          id="confirmPassword"
          minlength="8"
          bind:value={confirmPassword}
        />
        <button class="button1 my-4" disabled={isSubmittingSignup} type="submit"
          >Sign up
          {#if isSubmittingSignup}<Dots />{/if}
        </button>

        {#if error}
          <div class="bg-red rounded p-2">{error}</div>
        {/if}
      </form>
    {/if}

    <div class="divider" />
    <div class="flex">
      Already have an account?&nbsp;<a href="/login" use:link>Log in</a>
    </div>
  </div>
</div>
