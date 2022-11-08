<script type="ts">
  import {
    faHand,
    faHippo,
    faList,
    faMoneyBill,
    faUserAstronaut,
  } from "@fortawesome/free-solid-svg-icons";
  import { links } from "svelte-routing";
  import { error, isSubmitting } from "../../ts/mainStore";
  import Spinner from "../Spinner.svelte";
  import NavItem from "../NavItem.svelte";
  import detectEthereumProvider from "@metamask/detect-provider";
  import { Web3Provider } from "@ethersproject/providers";
  import { ethers } from "ethers";
  import { MembershipABI } from "../../contracts/Membership";
  import Fa from "svelte-fa";
  import { getContext } from "svelte";
  import MembershipStatus from "./MembershipStatus.svelte";
  import { membershipContract, provider, signer } from "../../ts/daaStore";

  let pathname = "/";
  if (typeof window !== "undefined") {
    pathname = window.location.pathname;
  }

  const membershipStatusMapping = [
    "Not a member",
    "Membership requested",
    "Whitelisted by one",
    "Member",
  ];

  export let ethereumAddress = null;
  export let membershipStatus;

  const { open } = getContext("simple-modal");
  const showMembershipStatus = () => open(MembershipStatus);

  async function connectWallet() {
    let ethProv = await detectEthereumProvider();

    if (ethProv) {
      $provider = new Web3Provider(<any>ethProv);
      await $provider.send("eth_requestAccounts", []);
      $signer = $provider.getSigner();
      ethereumAddress = await $signer.getAddress();

      $membershipContract = new ethers.Contract(
        import.meta.env.VITE_MEMBERSHIP_CONTRACT_ADDRESS,
        MembershipABI,
        $signer
      );

      membershipStatus =
        membershipStatusMapping[
          await $membershipContract.getMembershipStatus(ethereumAddress)
        ];
    } else {
      $error = "MetaMask not detected in your browser.";
    }
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
    <NavItem href="/daa/delegate" icon={faHippo} label="Delegate functions" />
  </nav>
  <div>
    {#if $isSubmitting}
      <Spinner />
    {/if}
    <slot />
  </div>

  <div class="memberArea">
    <Fa icon={faUserAstronaut} size="3x" />

    {#if ethereumAddress === null}
      <button class="button1" on:click={connectWallet}>Connect wallet</button>
    {:else}
      <p>
        Hello {ethereumAddress}! <br />
        Your status: {membershipStatus}
      </p>
      <button class="py-2 button3" on:click={showMembershipStatus}
        >Approval process</button
      >
    {/if}
  </div>
</div>
