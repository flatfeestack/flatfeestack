<script lang="ts">
  import { navigate } from "svelte-routing";
  import { chairmen, userEthereumAddress } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import AddVotingSlot from "./chairman/AddVotingSlot.svelte";
  import CancelVotingSlot from "./chairman/CancelVotingSlot.svelte";
  import Navigation from "./Navigation.svelte";

  $: {
    if ($chairmen === null || $userEthereumAddress === null) {
      $isSubmitting = true;
    } else if (
      !$chairmen.some((chairman) => chairman == $userEthereumAddress)
    ) {
      $error = "You are not allowed to review this page.";
      navigate("/daa/votes");
    } else {
      $isSubmitting = false;
    }
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Chairman functions</h1>

  <AddVotingSlot />
  <CancelVotingSlot />
</Navigation>
