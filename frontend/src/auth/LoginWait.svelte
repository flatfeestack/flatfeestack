<script lang="ts">
  import Modal from "./Modal.svelte";
  let { params }: { params: { email: string } } = $props();
  let email = params.email;
</script>

<style>

</style>

<Modal>
  <h1>Check your inbox</h1>
  <div>We've sent a login link to {email}</div>
  <div class="divider"></div>
  <div class="pt-100 small">
    Click the link in the email to sign in. The link expires in 10 minutes.
  </div>
</Modal>
