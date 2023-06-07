<script lang="ts">
  import { onMount } from "svelte";
  import { API } from "./../ts/api";
  import { error } from "../ts/mainStore";
  import { navigate } from "svelte-routing";

  export let email: string;
  export let token: string;

  onMount(async () => {
    try {
      await API.user.confirmGitEmail(email, token);
      navigate("/user/settings");
    } catch (e) {
      $error = e;
      navigate("/user/search");
    }
  });
</script>
