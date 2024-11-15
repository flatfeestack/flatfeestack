<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "./ts/api.ts";
  import { appState } from "ts/state.ts";
  import {goto} from "@mateothegreat/svelte5-router";

  let {email, token} = $props<{email: string; name: string;}>();

  onMount(async () => {
    try {
      await API.user.confirmGitEmail(email, token);
      goto("/user/settings");
    } catch (e) {
      appState.setError(e);
      goto("/user/search");
    }
  });
</script>
