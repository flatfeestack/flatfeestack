<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "@/api";
  import { error } from "@/mainStore";
  import {goto} from "@mateothegreat/svelte5-router";

  export let email: string;
  export let token: string;

  onMount(async () => {
    try {
      await API.user.confirmGitEmail(email, token);
      goto("/user/settings");
    } catch (e) {
      $error = e;
      goto("/user/search");
    }
  });
</script>
