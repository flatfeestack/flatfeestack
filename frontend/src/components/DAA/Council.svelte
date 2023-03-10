<script lang="ts">
  import { navigate } from "svelte-routing";
  import { councilMembers, userEthereumAddress } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import AddVotingSlot from "./council/AddVotingSlot.svelte";
  import CancelVotingSlot from "./council/CancelVotingSlot.svelte";
  import Navigation from "./Navigation.svelte";
  import MembershipRequests from "./council/MembershipRequests.svelte";

  $: {
    if ($councilMembers === null || $userEthereumAddress === null) {
      $isSubmitting = true;
    } else if (
      !$councilMembers.some((member) => member == $userEthereumAddress)
    ) {
      $error = "You are not allowed to view this page.";
      navigate("/daa/home");
    } else {
      $isSubmitting = false;
    }
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Council Member functions</h1>

  <AddVotingSlot />
  <CancelVotingSlot />
  <MembershipRequests />
</Navigation>
