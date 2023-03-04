<script type="ts">
  import { Web3Provider } from "@ethersproject/providers";
  import {
    faHand,
    faHippo,
    faList,
    faMoneyBill,
    faPerson,
    faUserAstronaut,
  } from "@fortawesome/free-solid-svg-icons";
  import detectEthereumProvider from "@metamask/detect-provider";
  import { getContext, onMount } from "svelte";
  import Fa from "svelte-fa";
  import { links, navigate } from "svelte-routing";
  import {
    councilMembers,
    membershipStatusValue,
    provider,
    signer,
    userEthereumAddress,
  } from "../../ts/daaStore";
  import { isSubmitting } from "../../ts/mainStore";
  import membershipStatusMapping from "../../utils/membershipStatusMapping";
  import NavItem from "../NavItem.svelte";
  import Spinner from "../Spinner.svelte";
  import truncateEthAddress from "../../utils/truncateEthereumAddress";

  let pathname = "/";
  if (typeof window !== "undefined") {
    pathname = window.location.pathname;
  }

  let membershipStatus: string;

  const showMetaMaskRequired = () => navigate("/daa/metamask");

  onMount(async () => {
    try {
      const ethProv = await detectEthereumProvider();
      $provider = new Web3Provider(<any>ethProv);
    } catch {
      showMetaMaskRequired();
    }
  });

  async function connectWallet() {
    if ($provider === null) {
      showMetaMaskRequired();
    } else {
      await $provider.send("eth_requestAccounts", []);
      $signer = $provider.getSigner();
    }
  }

  $: {
    if ($membershipStatusValue === null || $councilMembers === null) {
      membershipStatus = "Loading ...";
    } else {
      membershipStatus = resolveMembershipStatus();
    }
  }

  function resolveMembershipStatus(): string {
    if (
      $membershipStatusValue == 3 &&
      $councilMembers?.includes($userEthereumAddress)
    ) {
      return "Council Member";
    }

    return membershipStatusMapping[$membershipStatusValue];
  }
</script>

<style>
  .page {
    flex: 1 1 auto;
    display: flex;
  }

  .navigation {
    display: flex;
    flex-flow: column;
    min-width: 12rem;
    background-color: var(--secondary-100);
    border-right: solid 1px var(--secondary-300);
    white-space: nowrap;
    padding: 2rem 1em;
  }

  .navigation :global(a),
  .navigation {
    display: block;
    color: var(--secondary-700);
    text-decoration: none;
    transition: color 0.3s linear, background-color 0.3s linear;
    word-wrap: break-word;
  }

  .navigation :global(a) {
    padding: 1em 0;
  }

  .navigation :global(a:hover),
  .navigation {
    background-color: var(--primary-300);
    color: var(--primary-900);
  }

  .memberArea {
    display: flex;
    flex-flow: column;
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
    width: 100%;
  }
</style>

<div class="page">
  <div class="navigation">
    <div class="memberArea">
      <Fa icon={faUserAstronaut} size="3x" />

      {#if $userEthereumAddress === null}
        <button class="button1" on:click={connectWallet}>Connect wallet</button>
      {:else}
        <p>
          Hello {truncateEthAddress($userEthereumAddress)}! <br />
          Your status: {membershipStatus}
        </p>
      {/if}
    </div>
    <nav use:links>
      <NavItem href="/daa/votes" icon={faList} label="Votes" />
      <NavItem href="/daa/treasury" icon={faMoneyBill} label="Treasury" />
      <NavItem
        href="/daa/membershipRequests"
        icon={faHand}
        label="Membership requests"
      />
      <NavItem href="/daa/council" icon={faHippo} label="Council functions" />
      <NavItem href="/daa/membership" icon={faPerson} label="Membership" />
    </nav>
  </div>
  <div class="content">
    {#if $isSubmitting}
      <Spinner />
    {:else}
      <slot />
    {/if}
  </div>
</div>
