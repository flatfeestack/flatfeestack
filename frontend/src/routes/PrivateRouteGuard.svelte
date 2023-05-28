<script>
  import { navigate } from "svelte-routing";
  import { onMount } from "svelte";
  import { API } from "../ts/api";

  import { user, error } from "../ts/mainStore";

  let userFromStoreOrAPI = {};
  onMount(async () => {
    // have to do this to prevent wrong behaviour from loading it solely via store...
    try {
      userFromStoreOrAPI =
        $user.id == undefined ? await API.user.get() : $user;
      if (!userFromStoreOrAPI.id) {
        navigate("/login");
      }
    } catch (e) {
      $error = "Please log in or create an account to access FlatFeeStack.";
      navigate("/login");
    }
  });
</script>

{#if userFromStoreOrAPI.id}
  <slot />
{/if}
