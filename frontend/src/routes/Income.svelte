<script lang="ts">
  import { BigNumber, ethers, type Contract, type Signature } from "ethers";
  import { splitSignature } from "ethers/lib/utils";
  import { onMount } from "svelte";
  import Navigation from "../components/Navigation.svelte";
  import Spinner from "../components/Spinner.svelte";
  import { PayoutERC20ABI } from "../contracts/PayoutERC20";
  import { PayoutEthABI } from "../contracts/PayoutEth";
  import { API } from "../ts/api";
  import { provider, signer } from "../ts/ethStore";
  import { error } from "../ts/mainStore";
  import { formatBalance, formatDate, timeSince } from "../ts/services";
  import type { PayoutResponse } from "../types/backend";
  import type { PayoutConfig } from "../types/payout";
  import setSigner from "../utils/setSigner";
  import { ensureSameChainId } from "../utils/ethHelpers";

  let ethSignature: Signature;
  let isLoading = false;
  let payoutConfig: PayoutConfig;
  let payoutSignature: PayoutResponse;

  async function requestPayout(selectedCurrency: "ETH" | "GAS" | "USD") {
    payoutSignature = await API.user.requestPayout(selectedCurrency);

    if (selectedCurrency !== "GAS") {
      ethSignature = splitSignature(payoutSignature.signature);
    }
  }

  async function doEthPayout() {
    isLoading = true;

    if ($signer === null) {
      await setSigner($provider);
    }

    ensureSameChainId(payoutConfig.chainId);

    let contract: Contract;

    if (payoutSignature.currency === "ETH") {
      contract = new ethers.Contract(
        payoutConfig.payoutContractAddresses.eth,
        PayoutEthABI,
        $signer
      );
    } else {
      contract = new ethers.Contract(
        payoutConfig.payoutContractAddresses.usdc,
        PayoutERC20ABI,
        $signer
      );
    }

    try {
      await contract.withdraw(
        await $signer.getAddress(),
        payoutSignature.encodedUserId,
        BigNumber.from(String(payoutSignature.amount)),
        ethSignature.v,
        ethSignature.r,
        ethSignature.s
      );
    } catch (exception) {
      console.error(exception);
      $error = exception.data.data.reason;
      resetViewAfterPayout();
    }

    resetViewAfterPayout();
  }

  function resetViewAfterPayout() {
    payoutSignature = undefined;
    isLoading = false;
    document.body.scrollIntoView();
  }

  onMount(async () => {
    payoutConfig = await API.payout.payoutConfig();
  });
</script>

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
            <button on:click={() => doEthPayout()} class="button1"
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
                <td
                  ><a href={contribution.repoUrl}>{contribution.repoName}</a
                  ></td
                >
                <td
                  >{contribution.sponsorName
                    ? contribution.sponsorName
                    : "[no name]"}</td
                >
                <td
                  >{formatBalance(
                    BigInt(contribution.balance),
                    contribution.currency
                  )}</td
                >
                <td>{contribution.currency}</td>
                <td>
                  {#if contribution.claimedAt === null}
                    Unclaimed
                  {:else}
                    Realized
                  {/if}
                </td>
                <td title={formatDate(new Date(contribution.day))}>
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
    {($error = err.message)}
  {/await}
</Navigation>
