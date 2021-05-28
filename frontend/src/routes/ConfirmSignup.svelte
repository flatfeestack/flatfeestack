<script lang="ts">
  import { Link, navigate } from "svelte-routing";
  import { onMount } from 'svelte';
  import { confirmEmail } from "../ts/services";
  import { firstTime } from "../ts/store";

  export let email;
  export let token;
  let error = "";

  onMount(async () => {
    try {
      await confirmEmail(email, token);
      $firstTime = true;
      navigate("/user/settings");
    } catch (e) {
      error = e;
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

<div class="max">
  <div class="box-container rounded p-5">
    <h2 class="py-5 text-center text-primary-900">Confirm your email</h2>

    {#if error}
      <div class="bg-red rounded p-2">{error}</div>
    {/if}

    <div class="divider"></div>
    <div class="flex">
      Already have an account?&nbsp;<Link to="/signin">Log in</Link>
    </div>

  </div>
</div>
