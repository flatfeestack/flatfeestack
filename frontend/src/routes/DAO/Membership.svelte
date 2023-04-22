<script lang="ts">
  import {
    faSquareCheck,
    faSquareXmark,
  } from "@fortawesome/free-solid-svg-icons";
  import { getContext } from "svelte";
  import Fa from "svelte-fa";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import RequestMembership from "../../components/DAO/membership/RequestMembership.svelte";
  import Dialog from "../../components/Dialog.svelte";
  import { membershipContract, membershipStatusValue } from "../../ts/daoStore";
  import { userEthereumAddress } from "../../ts/ethStore";
  import { error } from "../../ts/mainStore";
  import checkUndefinedProvider from "../../utils/checkUndefinedProvider";

  let membershipFeePaid = false;
  let walletConnected: boolean;
  let nextMembershipFeePayment = 0;

  checkUndefinedProvider();

  const approvalProcessSteps = [
    "Membership to the DAO has been requested",
    "Membership has been approved by one council member",
    "Membership has been approved by a second council member",
  ];

  const { open } = getContext("simple-modal");

  async function prepareView() {
    if ($membershipStatusValue == 3) {
      nextMembershipFeePayment =
        await $membershipContract.nextMembershipFeePayment(
          $userEthereumAddress
        );
      membershipFeePaid =
        nextMembershipFeePayment > Math.floor(Date.now() / 1000);
    }
  }

  $: {
    if (
      $membershipStatusValue !== null &&
      $membershipContract !== null &&
      $userEthereumAddress !== null
    ) {
      walletConnected = true;
      prepareView();
    } else {
      walletConnected = false;
    }
  }

  async function payMembershipFees() {
    try {
      const membershipFee = await $membershipContract.membershipFee();
      await $membershipContract.payMembershipFee({
        value: membershipFee,
      });
    } catch (e) {
      showError(e);
    }
  }

  const onMembershipConfirm = async () => {
    try {
      await $membershipContract.requestMembership();
    } catch (e) {
      showError(e);
    }
  };

  function showError(e: Error) {
    $error = e.message;
  }

  const onCancel = () => {};
  const onMembershipCancel = () => {};
  const onConfirm = async () => {
    try {
      await $membershipContract.removeMember($userEthereumAddress);
    } catch (e) {
      $error = e.data.data.reason;
    }
  };

  const leaveFlatFeeStack = () => {
    open(
      Dialog,
      {
        title: "Leave FlatFeeStack",
        message: "Are you sure you want to leave FlatFeeStack immediately?",
        onCancel,
        onConfirm,
      },
      {
        closeButton: false,
        closeOnEsc: false,
        closeOnOuterClick: false,
      }
    );
  };

  const requestMembership = () => {
    open(
      RequestMembership,
      {
        onMembershipCancel,
        onMembershipConfirm,
      },
      {
        closeButton: false,
        closeOnEsc: true,
        closeOnOuterClick: true,
      }
    );
  };
</script>

<style>
  ul {
    list-style-type: none;
    padding: 0;
    margin: 0;
  }
</style>

<Navigation>
  {#if !walletConnected}
    <div class="centerContainer">
      <p>Please connect you Wallet.</p>
    </div>
  {:else}
    <div>
      <h1 class="text-secondary-900">Membership</h1>

      <h3 class="text-secondary-900">Membership approval process</h3>

      <ul>
        {#each Array(3) as _, i}
          <li
            class={$membershipStatusValue >= i + 1
              ? "text-tertiary-500"
              : "text-red"}
          >
            {#if $membershipStatusValue >= i + 1}
              <Fa icon={faSquareCheck} size="sm" class="icon" />
            {:else}
              <Fa icon={faSquareXmark} size="sm" class="icon" />
            {/if}
            {approvalProcessSteps[i]}
          </li>
        {/each}

        <li class={membershipFeePaid ? "text-tertiary-500" : "text-red"}>
          {#if membershipFeePaid}
            <Fa icon={faSquareCheck} size="sm" class="icon" />
          {:else}
            <Fa icon={faSquareXmark} size="sm" class="icon" />
          {/if}
          Membership fees paid {#if membershipFeePaid}
            (until {new Date(
              Number(nextMembershipFeePayment) * 1000
            ).toLocaleDateString()}){/if}
        </li>
      </ul>

      <div class="py-2">
        {#if $membershipStatusValue == 0}
          <p>You seem not to be a member of the FlatFeeStack Association.</p>
          <p>If you want to shape the future of this product, join us!</p>
          <button on:click={requestMembership} class="button4"
            >Request membership</button
          >
        {/if}
      </div>

      <div class="py-2">
        {#if $membershipStatusValue == 3 && !membershipFeePaid}
          <button on:click={payMembershipFees} class="button4"
            >Pay membership fees</button
          >
        {/if}
      </div>

      {#if $membershipStatusValue > 0}
        <button class="py-2 button3 my-2" on:click={leaveFlatFeeStack}>
          Leave FlatFeeStack
        </button>
      {/if}
    </div>
  {/if}
</Navigation>
