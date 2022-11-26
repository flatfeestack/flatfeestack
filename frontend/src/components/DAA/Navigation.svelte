<script type="ts">
  import { Web3Provider } from "@ethersproject/providers";
  import {
    faHand,
    faHippo,
    faList,
    faMoneyBill,
    faUserAstronaut,
  } from "@fortawesome/free-solid-svg-icons";
  import detectEthereumProvider from "@metamask/detect-provider";
  import { getContext, onMount } from "svelte";
  import Fa from "svelte-fa";
  import { links } from "svelte-routing";
  import {
    userEthereumAddress,
    membershipContract,
    membershipStatusValue,
    provider,
    signer,
    whitelisters,
    chairmanAddress,
  } from "../../ts/daaStore";
  import { isSubmitting } from "../../ts/mainStore";
  import membershipStatusMapping from "../../utils/membershipStatusMapping";
  import NavItem from "../NavItem.svelte";
  import Spinner from "../Spinner.svelte";
  import MembershipStatus from "./MembershipStatus.svelte";
  import MetaMaskRequired from "./MetaMaskRequired.svelte";
  import { error } from "../../ts/mainStore";
  import Dialog from "../Dialog.svelte";

  let pathname = "/";
  if (typeof window !== "undefined") {
    pathname = window.location.pathname;
  }

  let membershipStatus = "Loading ...";
  let metaMaskMissing = false;

  const { open } = getContext("simple-modal");
  const showMembershipStatus = () => open(MembershipStatus);
  const showMetaMaskRequired = () =>
    open(MetaMaskRequired, {}, { closeButton: false });

  onMount(async () => {
    try {
      const ethProv = await detectEthereumProvider();
      $provider = new Web3Provider(<any>ethProv);
    } catch {
      metaMaskMissing = true;
      showMetaMaskRequired();
    }
  });

  async function connectWallet() {
    if ($provider === null) {
      metaMaskMissing = true;
      showMetaMaskRequired();
    } else {
      await $provider.send("eth_requestAccounts", []);
      $signer = $provider.getSigner();
    }
  }

  const onCancel = () => {};
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

  $: {
    if (
      $membershipStatusValue === null ||
      $whitelisters === null ||
      $chairmanAddress === null
    ) {
      membershipStatus = "Loading ...";
    } else {
      membershipStatus = resolveMembershipStatus();
    }
  }

  function resolveMembershipStatus(): string {
    if ($membershipStatusValue == 3) {
      if ($chairmanAddress == $userEthereumAddress) {
        return "Chairman";
      } else if ($whitelisters?.includes($userEthereumAddress)) {
        return "Whitelister";
      }
    }

    return membershipStatusMapping[$membershipStatusValue];
  }
</script>

<style>
  .page {
    flex: 1 1 auto;
    display: flex;
  }

  nav {
    padding-top: 2rem;
    display: flex;
    flex-flow: column;
    min-width: 12rem;
    background-color: var(--secondary-100);
    border-right: solid 1px var(--secondary-300);
    white-space: nowrap;
  }

  nav :global(a),
  nav {
    display: block;
    color: var(--secondary-700);
    padding: 1em;
    text-decoration: none;
    transition: color 0.3s linear, background-color 0.3s linear;
  }

  nav :global(a:hover),
  nav {
    background-color: var(--primary-300);
    color: var(--primary-900);
  }

  .memberArea {
    display: flex;
    flex-flow: column;
    max-width: 12rem;
    margin-left: auto;
    padding: 1rem;
    border-left: solid 1px var(--secondary-300);
    overflow-wrap: anywhere;
    font-size: 0.8rem;
  }

  @media (max-width: 36rem) {
    .page {
      flex-direction: column;
      display: flex;
    }

    nav {
      display: flex;
      flex-direction: row;
      justify-content: space-between;
      width: 99.9%;
      border-bottom: solid 1px var(--primary-500);
      padding: 0;
    }

    nav :global(a) {
      text-align: center;
      width: 100%;
      float: left;
    }
  }

  .content {
    padding: 1rem;
  }
</style>

<div class="page">
  <nav use:links>
    <NavItem href="/daa/votes" icon={faList} label="Votes" />
    <NavItem href="/daa/treasury" icon={faMoneyBill} label="Treasury" />
    <NavItem
      href="/daa/membershipRequests"
      icon={faHand}
      label="Membership requests"
    />
    <NavItem href="/daa/chairman" icon={faHippo} label="Chairman functions" />
  </nav>
  <div class="content">
    {#if $isSubmitting}
      <Spinner />
    {:else if metaMaskMissing}
      <div />
    {:else}
      <slot />
    {/if}
  </div>

  <div class="memberArea">
    <Fa icon={faUserAstronaut} size="3x" />

    {#if $userEthereumAddress === null}
      <button class="button1" on:click={connectWallet}>Connect wallet</button>
    {:else}
      <p>
        Hello {$userEthereumAddress}! <br />
        Your status: {membershipStatus}
      </p>
      <button class="py-2 button3 my-2" on:click={showMembershipStatus}>
        Approval process
      </button>
      {#if $membershipStatusValue > 0}
        <button class="py-2 button3 my-2" on:click={leaveFlatFeeStack}>
          Leave FlatFeeStack
        </button>
      {/if}
    {/if}
  </div>
</div>
