<script lang="ts">
  import { navigate } from "svelte-routing";
  import { councilMembers, userEthereumAddress } from "../../ts/daoStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import AddVotingSlot from "../../components/DAO/council/AddVotingSlot.svelte";
  import CancelVotingSlot from "../../components/DAO/council/CancelVotingSlot.svelte";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import MembershipRequests from "../../components/DAO/council/MembershipRequests.svelte";
  import checkUndefinedProvider from "../../utils/checkUndefinedProvider";
  import { onDestroy } from "svelte";

  checkUndefinedProvider();

  $: {
    if ($councilMembers === null || $userEthereumAddress === null) {
      $isSubmitting = true;
    } else if (
      !$councilMembers.some((member) => member == $userEthereumAddress)
    ) {
      $error = "You are not allowed to view this page.";
      navigate("/dao/home");
    } else {
      $isSubmitting = false;
    }
  }

  onDestroy(() => {
    $isSubmitting = false;
  });
</script>

<Navigation>
  <h1 class="text-secondary-900">Council Member functions</h1>

  <AddVotingSlot />
  <CancelVotingSlot />
  <MembershipRequests />
</Navigation>
