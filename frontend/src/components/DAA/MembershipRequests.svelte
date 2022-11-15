<script lang="ts">
  import { onMount } from "svelte";
  import { navigate } from "svelte-routing";
  import { userEthereumAddress, membershipContract } from "../../ts/daaStore";
  import { error } from "../../ts/mainStore";
  import membershipStatusMapping from "../../utils/membershipStatusMapping";
  import Navigation from "./Navigation.svelte";

  let isWhiteLister;
  let membersToBeConfirmed: MemberToConfirm[] = [];

  interface MemberToConfirm {
    address: String;
    status: String;
    canConfirm: boolean;
  }

  onMount(async () => {
    if ($userEthereumAddress === null || $membershipContract === null) {
      moveToVotesPage();
      return;
    }

    isWhiteLister = await $membershipContract.isWhitelister(
      $userEthereumAddress
    );

    if (!isWhiteLister) {
      moveToVotesPage();
    }

    const [requestingMembers, confirmedMembers] = await Promise.all([
      $membershipContract.queryFilter(
        $membershipContract.filters.ChangeInMembershipStatus(null, [1, 2])
      ),
      $membershipContract.queryFilter(
        $membershipContract.filters.ChangeInMembershipStatus(null, 3)
      ),
    ]);

    // immense sorting function, but in a nutshell
    // 1. remove all members that are already confirmed
    // 2. reverse sort the array by highest membership status
    // 3. filter out duplicate addresses while preserving the order
    // 4. map it to a more handy object to display it
    membersToBeConfirmed = await Promise.all(
      requestingMembers
        .filter(
          (requestEvent) =>
            !confirmedMembers.some(
              (confirmedEvent) =>
                confirmedEvent.args[0] === requestEvent.args[0]
            )
        )
        .sort((requestEvent) => requestEvent.args[1])
        .reverse()
        .filter(
          (element, index, array) =>
            array.findIndex((v2) => v2.args[0] === element.args[0]) === index
        )
        .map(async (requestingMember): Promise<MemberToConfirm> => {
          const canConfirm =
            requestingMember.args[1] === 1
              ? true
              : (await $membershipContract.getFirstWhitelister(
                  requestingMember.args[0]
                )) !== $userEthereumAddress;

          return {
            address: requestingMember.args[0],
            status: membershipStatusMapping[requestingMember.args[1]],
            canConfirm: canConfirm,
          };
        })
    );
  });

  function moveToVotesPage() {
    $error = "You are not allowed to review this page.";
    navigate("/daa/votes");
  }

  async function whitelistMember(address: String) {
    try {
      await $membershipContract.whitelistMember(address);
    } catch (e) {
      $error = e.message;
    }
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Current membership requests</h1>

  {#if membersToBeConfirmed.length > 0}
    <table>
      <thead>
        <tr>
          <th>Address</th>
          <th>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each membersToBeConfirmed as memberToBeConfirmed}
          <tr>
            <td>{memberToBeConfirmed.address}</td>
            <td>{memberToBeConfirmed.status}</td>
            <td
              ><button
                class="button1"
                disabled={!memberToBeConfirmed.canConfirm}
                on:click={() => whitelistMember(memberToBeConfirmed.address)}
                >Confirm member</button
              ></td
            >
          </tr>
        {/each}
      </tbody>
    </table>
  {:else}
    There are no members to confirm.
  {/if}
</Navigation>
