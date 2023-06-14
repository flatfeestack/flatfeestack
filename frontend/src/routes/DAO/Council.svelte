<script lang="ts">
  import { onDestroy } from "svelte";
  import { navigate } from "svelte-routing";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import AddVotingSlot from "../../components/DAO/council/AddVotingSlot.svelte";
  import CancelVotingSlot from "../../components/DAO/council/CancelVotingSlot.svelte";
  import MembershipRequests from "../../components/DAO/council/MembershipRequests.svelte";
  import { councilMembers, daoConfig } from "../../ts/daoStore";
  import { userEthereumAddress } from "../../ts/ethStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import { checkUndefinedProvider } from "../../utils/ethHelpers";

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

<Navigation requiresChainId={$daoConfig?.chainId}>
  <h1 class="text-secondary-900">Council Member functions</h1>

  <AddVotingSlot />
  <CancelVotingSlot />
  <MembershipRequests />
</Navigation>
