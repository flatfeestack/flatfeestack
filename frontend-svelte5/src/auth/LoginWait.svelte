<script lang="ts">
  import Modal from "./Modal.svelte";
  import {query} from "@mateothegreat/svelte5-router";
  let email = query("email");
</script>

<style>

</style>

<Modal>
  <h2>Check your inbox</h2>
  <div>We've sent a login link to {email}</div>
  <div class="divider"></div>
  <div class="pt-100 small">
    Click the link in the email to sign in. The link expires in 10 minutes.
  </div>
</Modal>
