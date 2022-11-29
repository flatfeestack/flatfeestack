<script lang="ts">
  import { navigate } from "svelte-routing";
  import { chairmanAddress, userEthereumAddress } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import AddVotingSlot from "./chairman/AddVotingSlot.svelte";
  import AddWhitelister from "./chairman/AddWhitelister.svelte";
  import CancelVotingSlot from "./chairman/CancelVotingSlot.svelte";
  import RemoveWhitelister from "./chairman/RemoveWhitelister.svelte";
  import Navigation from "./Navigation.svelte";

  $: {
    if ($chairmanAddress === null || $userEthereumAddress === null) {
      $isSubmitting = true;
    } else if ($chairmanAddress !== $userEthereumAddress) {
      $error = "You are not allowed to review this page.";
      navigate("/daa/votes");
    } else {
      $isSubmitting = false;
    }
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Chairman functions</h1>

  <AddWhitelister />
  <RemoveWhitelister />
  <AddVotingSlot />
  <CancelVotingSlot />
</Navigation>
