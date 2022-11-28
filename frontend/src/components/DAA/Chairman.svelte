<script lang="ts">
  import type { Signer } from "ethers";
  import { navigate } from "svelte-routing";
  import {
    chairmanAddress,
    daaContract,
    membershipContract,
    userEthereumAddress,
    whitelisters,
  } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import AddWhitelister from "./chairman/AddWhitelister.svelte";
  import RemoveWhitelister from "./chairman/RemoveWhitelister.svelte";
  import Navigation from "./Navigation.svelte";

  let minimumWhitelister = 0;
  let toBeAdded = "";
  let toBeRemoved = "";

  let nonWhitelisters: Signer[] = [];

  $: {
    if (
      $daaContract === null ||
      $membershipContract === null ||
      $whitelisters === null
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
</Navigation>
