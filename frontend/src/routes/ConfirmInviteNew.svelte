<script lang="ts">
  import { Link, navigate } from "svelte-routing";
  import Dots from "../components/Dots.svelte";
  import { confirmInviteNew } from "../ts/services";
  import { API } from "../ts/api";

  export let email;
  export let emailToken;
  export let inviteEmail;
  export let expireAt;
  export let inviteToken;

  let password = "";
  let error = "";
  let isSubmitting = false;

  async function handleSubmit() {
    try {
      error = "";
      isSubmitting = true;
      await confirmInviteNew(email, password, emailToken, inviteEmail, expireAt, inviteToken);
      await API.user.topup();
      email = "";
      password = "";
      isSubmitting = false;
      navigate("/user/payments");
    } catch (e) {
      isSubmitting = false;
      error = "Something went wrong. Please try again.";
      console.log(e);
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
    <h2 class="py-5 text-center text-primary-900">{inviteEmail} invited you</h2>
      <form on:submit|preventDefault="{handleSubmit}">
        <label for="email" class="py-1">Email address</label>
        <input required size="100" maxlength="100" type="email" id="email" name="email" bind:value={email} class="rounded py-2 border-primary-900" />
        <label for="password" class="flex py-1">Password</label>
        <input required size="100" maxlength="100" type="password" id="password" minlength="8" bind:value={password} class="rounded py-2 border-primary-900"/>
        <button class="btn my-4" disabled="{isSubmitting}" type="submit">Sign up
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
