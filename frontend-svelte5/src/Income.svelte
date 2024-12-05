<script lang="ts">
  import { BrowserProvider, Contract, Signature } from "ethers";
  import { onMount } from "svelte";
  //import {goto} from "@mateothegreat/svelte5-router";
  import Navigation from "./Navigation.svelte";
  import Spinner from "./Spinner.svelte";
  //import { PayoutERC20ABI } from "../contracts/PayoutERC20";
  //import { PayoutEthABI } from "../contracts/PayoutEth";
  import { API } from "./ts/api.ts";
  import { appState } from "ts/state.svelte.ts";
  import { formatBalance, formatDate, timeSince } from "./ts/services.svelte.ts";
  import type { PayoutResponse } from "./types/backend";
  import type { PayoutConfig } from "./types/payout";
  //import setSigner from "./setSigner";
  //import showMetaMaskRequired from "./showMetaMaskRequired";

  let ethSignature: Signature;
  let isLoading = false;
  let payoutConfig: PayoutConfig;
  let payoutSignature: PayoutResponse;

  async function requestPayout(selectedCurrency: "ETH" | "GAS" | "USD") {
    try {
      payoutSignature = await API.user.requestPayout(selectedCurrency);

      if (selectedCurrency !== "GAS") {
        ethSignature = Signature.from(payoutSignature.signature);
      }
    } catch (e) {
      appState.setError(e);
    }
  }

  /*async function doEthPayout() {
    isLoading = true;

    try {
      const ethProv = await detectEthereumProvider();
      $provider = new BrowserProvider(<any>ethProv);
    } catch (exception) {
      $provider = undefined;
    }

    if ($signer === null) {
      await setSigner($provider);
    }

    const currentChainId = await getChainId();

    if (currentChainId === undefined) {
      showMetaMaskRequired();
    } else if (currentChainId !== payoutConfig.chainId) {
      $lastEthRoute = window.location.pathname;
      goto(
        `/differentChainId?required=${payoutConfig.chainId}&actual=${currentChainId}`
      );
    } else {
      let contract: Contract;

      if (payoutSignature.currency === "ETH") {
        contract = new Contract(
          payoutConfig.payoutContractAddresses.eth,
          PayoutEthABI,
          $signer
        );
      } else {
        contract = new Contract(
          payoutConfig.payoutContractAddresses.usdc,
          PayoutERC20ABI,
          $signer
        );
      }

      try {
        await contract.withdraw(
          await $signer.getAddress(),
          payoutSignature.encodedUserId,
          BigInt(payoutSignature.amount),
          ethSignature.v,
          ethSignature.r,
          ethSignature.s
        );
      } catch (exception) {
        if (exception.data?.data === undefined) {
          // we deal with a regular error, not one a "revert" from the blockchain
          $error = exception.message;
        } else {
          $error = exception.data.data.reason;
        }
      } finally {
        resetViewAfterPayout();
      }
    }
  }

  function resetViewAfterPayout() {
    payoutSignature = undefined;
    isLoading = false;
    document.body.scrollIntoView();
  }*/

  onMount(async () => {
    payoutConfig = await API.payout.payoutConfig();
  });
</script>

<style>
  @media screen and (max-width: 600px) {
    table {
      width: 100%;
    }
    table thead {
      border: none;
      clip: rect(0 0 0 0);
      height: 1px;
      margin: -1px;
      overflow: hidden;
      padding: 0;
      position: absolute;
      width: 1px;
    }

    table tr {
      border-bottom: 3px solid #fff;
      display: block;
    }

    table td {
      border-bottom: 1px solid #fff;
      display: block;
      font-size: 0.8em;
      text-align: right;
    }

    table td::before {
      content: attr(data-label);
      float: left;
      font-weight: bold;
      text-transform: uppercase;
    }

    table td:last-child {
      border-bottom: 0;
    }
  }
</style>

<Navigation>
  <h2 class="p-2 m-2">Income</h2>
  <p class="p-2 m-2 bold">
    How does FlatFeeStack send the contributions you received?
  </p>
  <p class="p-2 m-2">
    That's the neat thing - we don't! All jokes aside, when clicking the button
    below, we will generate a signature that allows you to withdraw your earned
    contributions directly from the smart contract.
  </p>

  <p class="p-2 m-2 bold">
    I received contributions in different currencies. How does that work?
  </p>

  <p class="p-2 m-2">
    You need to withdraw the received contributions in the original currency.
    Notable exception are contributions in US dollars, which are payed out using
    the USDC stable coin.
  </p>

  {#await API.user.contributionsRcv()}
    ...waiting
  {:then contributions}
    {#if isLoading}
      <Spinner />
    {:else}
      <div class="container">
        {#if contributions.some((contribution) => contribution.currency === "USD")}
          <button on:click={() => requestPayout("USD")} class="button1"
            >Request USDC payout</button
          >
        {/if}

        {#if contributions.some((contribution) => contribution.currency === "ETH")}
          <button on:click={() => requestPayout("ETH")} class="button1"
            >Request ETH payout</button
          >
        {/if}

        {#if contributions.some((contribution) => contribution.currency === "GAS")}
          <button on:click={() => requestPayout("GAS")} class="button1"
            >Request NEO Gas payout</button
          >
        {/if}
      </div>

      {#if payoutSignature !== undefined}
        <p class="p-2 m-2">
          Great, a signature for payout has been generated! Please call the
          withdraw function of our smart contract at {payoutConfig
            .payoutContractAddresses[payoutSignature.currency.toLowerCase()]} with
          the following parameters:
        </p>

        <ul>
          <li>The address where you want to receive the payout</li>
          <li>{payoutSignature.encodedUserId}</li>
          <li>{payoutSignature.amount}</li>
          {#if ethSignature !== undefined}
            <li>{ethSignature.v}</li>
            <li>{ethSignature.r}</li>
            <li>{ethSignature.s}</li>
          {:else}
            <li>{payoutSignature.signature}</li>
          {/if}
        </ul>

        {#if ethSignature !== undefined}
          <p class="p-2 m-2">
            ... or click the button below to connect MetaMask and let us prepare
            the transaction.
          </p>

          <div class="container">
            <button class="button1"
              >Withdraw!</button
            >
          </div>
        {/if}
      {/if}
    {/if}

    {#if contributions.length > 0}
      <div class="container">
        <table>
          <thead>
            <tr>
              <th>Repository</th>
              <th>From</th>
              <th>Balance</th>
              <th>Currency</th>
              <th>Realized</th>
              <th>Date</th>
            </tr>
          </thead>
          <tbody>
            {#each contributions as contribution}
              <tr>
                <td data-label="Repository"
                  ><a href={contribution.repoUrl}>{contribution.repoName}</a
                  ></td
                >
                <td data-label="From"
                  >{contribution.sponsorName
                    ? contribution.sponsorName
                    : "[no name]"}</td
                >
                <td data-label="Balance"
                  >{formatBalance(
                    BigInt(contribution.balance),
                    contribution.currency
                  )}</td
                >
                <td data-label="Currency">{contribution.currency}</td>
                <td data-label="Realized">
                  {#if contribution.claimedAt === null}
                    Unclaimed
                  {:else}
                    Realized
                  {/if}
                </td>
                <td
                  data-label="Date"
                  title={formatDate(new Date(contribution.day))}
                >
                  {timeSince(new Date(contribution.day), new Date())} ago
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="p-2 m-2">No contributions received so far.</p>
    {/if}
  {:catch err}
    {(appState.error = err)}
  {/await}
</Navigation>
