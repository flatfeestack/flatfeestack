<script lang="ts">
  import {
    faComment,
    faHippo,
    faHome,
    faList,
    faMoneyBill,
    faPerson,
    faUserAstronaut,
  } from "@fortawesome/free-solid-svg-icons";
  import detectEthereumProvider from "@metamask/detect-provider";
  import { BrowserProvider } from "ethers";
  import { onMount } from "svelte";
  import Fa from "svelte-fa";
  import { links, navigate } from "svelte-routing";
  import {
    councilMembers,
    daoConfig,
    membershipStatusValue,
  } from "../../ts/daoStore";
  import {
    getChainId,
    lastEthRoute,
    provider,
    signer,
    userEthereumAddress,
  } from "../../ts/ethStore";
  import { isSubmitting } from "../../ts/mainStore";
  import membershipStatusMapping from "../../utils/membershipStatusMapping";
  import setSigner from "../../utils/setSigner";
  import showMetaMaskRequired from "../../utils/showMetaMaskRequired";
  import truncateEthAddress from "../../utils/truncateEthereumAddress";
  import NavItem from "../NavItem.svelte";
  import Spinner from "../Spinner.svelte";
  import EnsureSameChainId from "../EnsureSameChainId.svelte";

  let pathname = "/";
  if (typeof window !== "undefined") {
    pathname = window.location.pathname;
  }

  let membershipStatus: string;

  export let requiresChainId: number | undefined = undefined;

  onMount(async () => {
    await setProvider();

    if ($provider !== undefined) {
      window.ethereum.on("chainChanged", (_networkId: string) => {
        setProvider();
      });

      window.ethereum.on("accountsChanged", (_accounts: string[]) => {
        if ($signer !== null) {
          setSigner($provider);
        }
      });
    }
  });

  $: {
    if ($membershipStatusValue === null || $councilMembers === null) {
      membershipStatus = "Loading ...";
    } else {
      membershipStatus = resolveMembershipStatus();
    }
  }

  function resolveMembershipStatus(): string {
    if (
      $membershipStatusValue == 3n &&
      $councilMembers?.includes($userEthereumAddress)
    ) {
      return "Council Member";
    }

    return membershipStatusMapping[Number($membershipStatusValue)];
  }

  async function setProvider() {
    try {
      const ethProv = await detectEthereumProvider();
      $provider = new BrowserProvider(<any>ethProv);
    } catch (error) {
      $provider = undefined;
    }
  }

  async function triggerSetSigner() {
    const currentChainId = await getChainId();

    if (currentChainId === undefined) {
      showMetaMaskRequired();
    } else if (currentChainId !== $daoConfig.chainId) {
      $lastEthRoute = window.location.pathname;
      navigate(
        `/differentChainId?required=${$daoConfig.chainId}&actual=${currentChainId}`
      );
    } else {
      await setSigner($provider);
    }
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
    background-color: var(--tertiary-300);
    border-right: solid 1px var(--secondary-300);
    white-space: nowrap;
  }

  nav :global(a) {
    display: block;
    color: var(--secondary-700);
    padding: 1em;
    text-decoration: none;
    transition: color 0.3s linear, background-color 0.3s linear;
  }

  nav :global(a:hover) {
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
      display: block;
    }

    .navigation {
      display: flex;
      flex-direction: column;
    }

    nav {
      display: flex;
      flex-direction: row;
      width: 100%;
      border-bottom: solid 1px var(--primary-500);
      padding: 0;
    }

    nav :global(a) {
      text-align: center;
      padding: 0.5em;
      flex: 1 0 auto;
    }
  }
</style>

<div class="page">
  <div class="navigation">
    <div class="memberArea">
      <Fa class="mb-5" icon={faUserAstronaut} size="3x" />

      {#if $userEthereumAddress === null}
        <button class="button4" on:click={() => triggerSetSigner()}
          >Connect wallet
        </button>
      {:else}
        <p>
          Hello {truncateEthAddress($userEthereumAddress)}! <br />
          Your status: {membershipStatus}
        </p>
      {/if}
    </div>
    <nav use:links>
      <NavItem href="/dao/home" icon={faHome} label="DAO Home" />
      <NavItem href="/dao/discussions" icon={faComment} label="Discussions" />
      <NavItem href="/dao/votes" icon={faList} label="Votes" />
      {#if $membershipStatusValue == 3n}
        <NavItem href="/dao/treasury" icon={faMoneyBill} label="Treasury" />
      {/if}
      {#if $membershipStatusValue == 3n && $councilMembers?.includes($userEthereumAddress)}
        <NavItem href="/dao/council" icon={faHippo} label="Council functions" />
      {/if}
      <NavItem href="/dao/membership" icon={faPerson} label="Membership" />
    </nav>
  </div>
  <div class="p-2">
    {#if requiresChainId}
      <EnsureSameChainId requiredChainId={requiresChainId} />
    {/if}

    {#if $isSubmitting}
      <Spinner />
    {:else}
      <slot />
    {/if}
  </div>
</div>
