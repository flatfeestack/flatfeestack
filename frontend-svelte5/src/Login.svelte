<script lang="ts">
  import { login } from "ts/auth";
  import Dots from "./Dots.svelte";
  import {goto, route} from "@mateothegreat/svelte5-router";
  import { emailValidationPattern } from "./utils";
  import {appState} from "./ts/state.ts";

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
      goto("/user/search");
      email = "";
      password = "";
    } catch (e) {
      isSubmitting = false;
      password = "";
      appState.setError(e);
    }
  }
</script>

<style>
  button,
  input:focus {
    outline: none;
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
    <h2 class="py-5 text-center text-primary-900">Sign in to FlatFeeStack</h2>

    <form on:submit|preventDefault={handleSubmit}>
      <label for="email" class="py-1">Email address</label>
      <input
        required
        size="100"
        maxlength="100"
        type="email"
        id="email"
        pattern={emailValidationPattern}
        name="email"
        tabindex="-3"
        bind:value={email}
      />

      <div class="flex py-1">
        <label for="password">Password</label>
        <label for="password">
          <a href="/forgot" use:route tabindex="-1">Forgot password?</a>
        </label>
      </div>

      <input
        required
        size="100"
        maxlength="100"
        type="password"
        id="password"
        minlength="8"
        tabindex="-2"
        bind:value={password}
      />
      <button class="button1 btn my-4" disabled={isSubmitting} type="submit"
        >Sign in
        {#if isSubmitting}<Dots />{/if}
      </button>

      {#if error}
        <div class="bg-red rounded p-2">{error}</div>
      {/if}
    </form>

    <div class="divider" ></div>
    <div class="flex">
      New to FlatFeeStack?&nbsp;<a href="/signup" use:route>Sign up</a>
    </div>
  </div>
</div>
