<script lang="ts">
  import {goto, route, QueryString} from "@mateothegreat/svelte5-router";
  import { emailValidationPattern } from "../utils.ts";
  import Modal from "./Modal.svelte";
  import {login} from "./auth.svelte.ts";

  let error = $state("");
  let email = $state("");

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    try {
      const hasLoggedIn = await login(email);
      if(hasLoggedIn) {
        goto("/user/search");
      } else {
        //goto("/login-wait/" + email);
        const query = new QueryString();
        query.set("email", email);
        query.goto("/login-wait");
        email = "";
      }
    } catch (e: unknown) {
      error = String(e);
    }
  }
</script>

<style>
  form {
    display: flex;
    flex-direction: column;
  }
</style>

<Modal>
  <form onsubmit={handleSubmit}>
    <label for="email" class="py-025">Email address</label>
    <input  required
            title="Please enter a valid email address"
            class="optional"
            maxlength="100"
            type="email"
            id="email"
            pattern={emailValidationPattern}
            name="email"
            tabindex="-3"
            placeholder="you@example.com"
            bind:value={email}
            aria-describedby="password-help-email"/>
    <div id="password-help-email" class="help-text">Enter a valid email (e.g., name@example.com)</div>

    <button class="button1 my-100" type="submit">
      Continue with email
    </button>

    {#if error}
      <div class="bg-red rounded p-025">{error}</div>
    {/if}
  </form>

  <div class="divider" ></div>
  <div class="pt-100 small">
    By continuing, you acknowledge FlatFeeStacks  <a use:route href="/toc">Privacy Policy</a>.
  </div>
</Modal>
