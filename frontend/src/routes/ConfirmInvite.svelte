<script lang="ts">
  import { navigate } from "svelte-routing";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { error } from "../ts/store";

  export let email;
  export let inviteEmail;
  export let inviteDate;
  export let inviteToken;

  onMount(async () => {
    try {
      await  API.auth.confirmInvite(email, inviteEmail, inviteDate, inviteToken);
      navigate("/user/payments");
    } catch (e) {
      $error = e
      console.log(e);
    }
  });
</script>
