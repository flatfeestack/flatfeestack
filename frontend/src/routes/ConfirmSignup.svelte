<script lang="ts">
  import { onMount } from "svelte";
  import { confirmEmail } from "../ts/services";
  import { navigate, link } from "svelte-routing";
  import Spinner from "../components/Spinner.svelte";

  export let email: string;
  export let token: string;
  let error = "";

  onMount(async () => {
    try {
      await confirmEmail(email, token);
      navigate("/user/settings");
    } catch (e) {
      error = "Something went wrong. Please try again.";
    }
  });
</script>

<style>
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

{#if error}
  <div class="max">
    <div class="box-container rounded p-5">
      <h2 class="py-5 text-center text-primary-900">Confirm your email</h2>

      <div class="bg-red rounded p-2">{error}</div>

      <div class="divider" />
      <div class="flex">
        Already have an account?&nbsp;<a href="/login" use:link>Log in</a>
      </div>
    </div>
  </div>
{:else}
  <Spinner />
{/if}
