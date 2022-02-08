<script lang="ts">
  import { navigate } from "svelte-routing";
  import { onMount } from 'svelte';
  import { API } from "./../ts/api";
  import { error } from "../ts/store";

  export let email;
  export let token;

  onMount(async () => {
    try {
      await API.user.confirmGitEmail(email, token);
      navigate("/user/settings");
    } catch (e) {
      $error = e
    }
  });
</script>
