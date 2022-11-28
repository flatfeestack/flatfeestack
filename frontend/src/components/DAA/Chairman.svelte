<script lang="ts">
  import type { Signer } from "ethers";
  import { navigate } from "svelte-routing";
  import {
    chairmanAddress,
    daaContract,
    membershipContract,
    provider,
    userEthereumAddress,
    whitelisters,
  } from "../../ts/daaStore";
  import { votingSlots } from "../../ts/proposalStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import AddVotingSlot from "./chairman/AddVotingSlot.svelte";
  import AddWhitelister from "./chairman/AddWhitelister.svelte";
  import CancelVotingSlot from "./chairman/CancelVotingSlot.svelte";
  import RemoveWhitelister from "./chairman/RemoveWhitelister.svelte";
  import Navigation from "./Navigation.svelte";

  let currentBlockNumber = 0;
  let minimumWhitelister = 0;
  let nonWhitelisters: Signer[] = [];

  $: {
    if (
      $daaContract === null ||
      $membershipContract === null ||
      $whitelisters === null ||
      $provider === null
    ) {
      $isSubmitting = true;
    } else if ($chairmanAddress !== $userEthereumAddress) {
      $error = "You are not allowed to review this page.";
      navigate("/daa/votes");
    } else if (nonWhitelisters.length === 0) {
      prepareView();
    }
  }

  async function prepareView() {
    minimumWhitelister = minimumWhitelister = (
      await $membershipContract.minimumWhitelister()
    ).toNumber();
    currentBlockNumber = await $provider.getBlockNumber();

    await setMembers();

    $isSubmitting = false;
  }

  async function setMembers() {
    const membersLength = await $membershipContract.getMembersLength();

    const allMembers = await Promise.all(
      [...Array(membersLength.toNumber()).keys()].map(async (index: Number) => {
        return await $membershipContract.members(index);
      })
    );

    nonWhitelisters = allMembers.filter(
      (address) =>
        !$whitelisters.some((whitelister) => whitelister == address) &&
        $chairmanAddress != address
    );
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Chairman functions</h1>

  <AddWhitelister {nonWhitelisters} membershipContract={$membershipContract} />
  <RemoveWhitelister
    {minimumWhitelister}
    membershipContract={$membershipContract}
    whitelisters={$whitelisters}
  />
  <AddVotingSlot {currentBlockNumber} daaContract={$daaContract} />
  <CancelVotingSlot daaContract={$daaContract} votingSlots={$votingSlots} />
</Navigation>
