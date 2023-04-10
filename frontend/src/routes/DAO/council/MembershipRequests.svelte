<script lang="ts">
  import {
    membershipContract,
    provider,
    userEthereumAddress,
  } from "$lib/ts/daoStore";
  import { error } from "$lib/ts/mainStore";
  import membershipStatusMapping from "$lib/utils/membershipStatusMapping";
  import Spinner from "$lib/components/Spinner.svelte";

  let isLoading = true;
  let membersToBeConfirmed: MemberToConfirm[] = [];

  interface MemberToConfirm {
    address: String;
    status: String;
    canConfirm: boolean;
  }

  $: {
    if ($provider === null || $membershipContract === null) {
      isLoading = true;
    } else {
      prepareView();
    }
  }

  async function prepareView() {
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
              : (await $membershipContract.getFirstApproval(
                  requestingMember.args[0]
                )) !== $userEthereumAddress;

          return {
            address: requestingMember.args[0],
            status: membershipStatusMapping[requestingMember.args[1]],
            canConfirm: canConfirm,
          };
        })
    );

    isLoading = false;
  }

  async function approveMember(address: String) {
    try {
      await $membershipContract.approveMembership(address);
    } catch (e) {
      $error = e.message;
    }
  }
</script>

<h1 class="text-secondary-900">Current membership requests</h1>

{#if isLoading}
  <Spinner />
{:else if membersToBeConfirmed.length > 0}
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
          <td>
            <button
              class="button4"
              disabled={!memberToBeConfirmed.canConfirm}
              on:click={() => approveMember(memberToBeConfirmed.address)}
              >Confirm member
            </button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else}
  There are no members to confirm.
{/if}
