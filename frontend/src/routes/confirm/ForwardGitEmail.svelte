<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "./../ts/api";
  import { error } from "../ts/mainStore";
  import { goto } from "$app/navigation";

  export let email: string;
  export let token: string;

  onMount(async () => {
    try {
      await API.user.confirmGitEmail(email, token);
      goto("/user/settings");
    } catch (e) {
      $error = e;
    }
  });
</script>
