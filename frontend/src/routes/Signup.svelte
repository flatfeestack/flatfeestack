<script lang="ts">
  import { Link } from "svelte-routing";
  import { API } from "./../ts/api";
  import Dots from "../components/Dots.svelte";

  let email = "";
  let password = "";
  let error = "";
  let info = "";
  let isSubmittingSignup = false;

  async function handleSubmit() {
    try {
      error = "";
      isSubmittingSignup = true;
      await API.auth.signup(email, password);
      email = "";
      password = "";
      info = "Your email is on the way. To enable your account, click on the link in the email.";
    } catch (e) {
      error = "Something went wrong. Please try again. ("+ e +")";
    } finally {
      isSubmittingSignup = false;
    }
  }
</script>

<style>
    button, input:focus{
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
    <h2 class="py-5 text-center text-primary-900">Create your account</h2>

    {#if info}
      <div class="bg-green rounded p-2">{info}</div>
    {:else}
      <form on:submit|preventDefault="{handleSubmit}">
        <label for="email" class="py-1">Email address</label>
        <input required size="100" maxlength="100" type="email" id="email" name="email" bind:value={email} class="rounded py-2 border-primary-900" />
        <label for="password" class="flex py-1">Password</label>
        <input required size="100" maxlength="100" type="password" id="password" minlength="8" bind:value={password} class="rounded py-2 border-primary-900"/>
        <button class="btn my-4" disabled="{isSubmittingSignup}" type="submit">Sign up
          {#if isSubmittingSignup}<Dots />{/if}
        </button>

        {#if error}
          <div class="bg-red rounded p-2">{error}</div>
        {/if}

      </form>
    {/if}

    <div class="divider"></div>
    <div class="flex">
      Already have an account?&nbsp;<Link to="/signin">Log in</Link>
    </div>

  </div>
</div>
