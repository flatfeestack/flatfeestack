<script lang="ts">
  import { Link, navigate } from "svelte-routing";
  import { onMount } from 'svelte';
  import { confirmEmail, updateUser } from "ts/authService";

  export let email;
  export let token;
  let error = "";

  onMount(async () => {
    try {
      await confirmEmail(email, token);
      await updateUser();
      navigate("/dashboard");
    } catch (e) {
      error = e
      console.log(e);
    }
  });
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
    <h2 class="py-5 text-center text-primary-700">Confirm your email</h2>

    {#if error}
      <div class="bg-red rounded p-2">{error}</div>
    {/if}

    <div class="divider"></div>
    <div class="flex">
      Already have an account?&nbsp;<Link to="/signin">Log in</Link>
    </div>

  </div>
</div>
