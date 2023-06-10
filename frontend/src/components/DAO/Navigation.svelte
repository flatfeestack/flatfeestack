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
  import { links } from "svelte-routing";
  import {
    councilMembers,
    daoConfig,
    membershipStatusValue,
  } from "../../ts/daoStore";
  import { provider, signer, userEthereumAddress } from "../../ts/ethStore";
  import { isSubmitting } from "../../ts/mainStore";
  import { ensureSameChainId } from "../../utils/ethHelpers";
  import membershipStatusMapping from "../../utils/membershipStatusMapping";
  import setSigner from "../../utils/setSigner";
  import truncateEthAddress from "../../utils/truncateEthereumAddress";
  import NavItem from "../NavItem.svelte";
  import Spinner from "../Spinner.svelte";

  let pathname = "/";
  if (typeof window !== "undefined") {
    pathname = window.location.pathname;
  }

  let membershipStatus: string;

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
    ensureSameChainId($daoConfig.chainId);
    await setSigner($provider);
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
        <button class="button4" on:click={() => triggerSetSigner()}
          >Connect wallet</button
        >
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
  <div class="content">
    {#if $isSubmitting}
      <Spinner />
    {:else}
      <slot />
    {/if}
  </div>
</div>
