<script lang="ts">
  import {navigate, link} from "preveltekit";
  import {onMount} from "svelte";
  import {confirm} from "./auth.svelte.ts";
  import Modal from "./Modal.svelte";

  let searchParams = $state(new URLSearchParams(window.location.search));
  let email = $derived(searchParams.get("email"));
  let emailToken = $derived(searchParams.get("emailToken"));

  //console.log(email);
  //console.log(emailToken);

  let error =$state("");

  onMount(async () => {
      if(email && emailToken ) {
          console.log(email);
          try {
              await confirm(email, emailToken);
              navigate("/user/search");
          } catch(e:unknown) {
              error = String(e);
          }

      }
  });
</script>

{#if error}
<Modal>
    <h2>Something went wrong</h2>
    <div class="bg-red rounded p-025">{error}</div>
    <div class="divider"></div>
    <div class="pt-100 small">
        Try to  <a use:link href="/login">login</a> again
    </div>
</Modal>
{/if}