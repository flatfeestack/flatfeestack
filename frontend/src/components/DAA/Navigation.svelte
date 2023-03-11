<script type="ts">
  import { Web3Provider } from "@ethersproject/providers";
  import {
    faHippo,
    faList,
    faHome,
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
  import { isSubmitting, route } from "../../ts/mainStore";
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
    padding-top: 2rem;
    display: flex;
    flex-flow: column;
    min-width: 14rem;
    background-color: var(--tertiary-300);
    border-right: solid 1px var(--secondary-300);
    white-space: nowrap;
  }
  .navigation :global(a),
  .navigation .inactive {
    display: block;
    color: var(--secondary-700);
    padding: 1em;
    text-decoration: none;
    transition: color 0.3s linear, background-color 0.3s linear;
  }

  .navigation .inactive {
    color: var(--secondary-300);
  }

  .navigation :global(a:hover),
  .navigation .selected {
    background-color: var(--tertiary-900);
    color: var(--tertiary-700);
  }

  .memberArea {
    display: flex;
    flex-flow: column;
    overflow-wrap: anywhere;
    font-size: 0.8rem;
    margin-bottom: 0.5rem;
    padding: 1em;
  }

  @media (max-width: 36rem) {
    .page {
      flex-direction: column;
      display: flex;
    }
    .navigation {
      display: flex;
      flex-direction: row;
      justify-content: space-between;
      width: 99.9%;
      border-bottom: solid 1px var(--primary-500);
      padding: 0;
    }
    .navigation :global(a) {
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
      <Fa class="mb-5" icon={faUserAstronaut} size="3x" />

      {#if $userEthereumAddress === null}
        <button class="button4" on:click={connectWallet}>Connect wallet</button>
      {:else}
        <p>
          Hello {truncateEthAddress($userEthereumAddress)}! <br />
          Your status: {membershipStatus}
        </p>
      {/if}
    </div>
    <nav use:links>
      <NavItem href="/daa/home" icon={faHome} label="DAO Home" />
      <NavItem href="/daa/votes" icon={faList} label="Votes" />
      {#if $membershipStatusValue == 3}
        <NavItem href="/daa/treasury" icon={faMoneyBill} label="Treasury" />
      {/if}
      {#if $membershipStatusValue == 3 && $councilMembers?.includes($userEthereumAddress)}
        <NavItem href="/daa/council" icon={faHippo} label="Council functions" />
      {/if}
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
