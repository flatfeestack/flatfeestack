<script lang="ts">
  import { onMount } from 'svelte';
  import { API } from "./../ts/api";
  import { error } from "../ts/store";
  import {navigate} from "svelte-routing";

  export let email;
  export let token;

  onMount(async () => {
    try {
      await API.user.confirmGitEmail(email, token);
      await navigate("/user/settings");
    } catch (e) {
      $error = e
    }
  });
</script>
