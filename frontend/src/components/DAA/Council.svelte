<script lang="ts">
  import { navigate } from "svelte-routing";
  import { councilMembers, userEthereumAddress } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import AddVotingSlot from "./council/AddVotingSlot.svelte";
  import CancelVotingSlot from "./council/CancelVotingSlot.svelte";
  import Navigation from "./Navigation.svelte";

  $: {
    if ($councilMembers === null || $userEthereumAddress === null) {
      $isSubmitting = true;
    } else if (
      !$councilMembers.some((member) => member == $userEthereumAddress)
    ) {
      $error = "You are not allowed to review this page.";
      navigate("/daa/votes");
    } else {
      $isSubmitting = false;
    }
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Council Member functions</h1>

  <AddVotingSlot />
  <CancelVotingSlot />
</Navigation>
