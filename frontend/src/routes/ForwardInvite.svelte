<script lang="ts">
  import { navigate } from "svelte-routing";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { error } from "../ts/store";

  export let email;
  export let inviteEmail;
  export let expireAt;
  export let inviteToken;

  onMount(async () => {
    try {
      await API.auth.confirmInvite(email, inviteEmail, expireAt, inviteToken);
      await API.user.topup();
      navigate("/user/invitations");
    } catch (e) {
      $error = e
    }
  });
</script>
