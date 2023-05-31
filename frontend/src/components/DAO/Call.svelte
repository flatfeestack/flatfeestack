<script lang="ts">
  import type { Bytes, Contract } from "ethers";
  import {
    daoConfig,
    daoContract,
    membershipContract,
    walletContract,
  } from "../../ts/daoStore";
  import { onDestroy, onMount } from "svelte";
  import type { Readable, Unsubscriber } from "svelte/store";
  import type { TransactionDescription } from "ethers/lib/utils";

  export let calldata: string;
  export let target: string;
  export let value: number;
  export let index;

  let canDecode = false;
  let targetName: string;
  let transactionData: TransactionDescription;
  let unsubscribe: Unsubscriber;

  function findCorrespondingContract(): Readable<Contract | null> | undefined {
    switch (target.toLowerCase()) {
      case $daoConfig.dao.toLowerCase():
        targetName = "DAO Contract";
        return daoContract;
      case $daoConfig.membership.toLowerCase():
        targetName = "Membership Contract";
        return membershipContract;
      case $daoConfig.wallet.toLowerCase():
        targetName = "Wallet Contract";
        return walletContract;
      default:
        return;
    }
  }

  function decodeCall() {
    const contract = findCorrespondingContract();

    if (contract === undefined) {
      return;
    }

    unsubscribe = contract.subscribe((contractValue) => {
      if (contractValue !== null) {
        transactionData = contractValue.interface.parseTransaction({
          data: calldata,
          value: value,
        });
        canDecode = true;
      }
    });
  }

  onMount(decodeCall);
  onDestroy(() => {
    if (unsubscribe !== undefined) {
      unsubscribe();
    }
  });
</script>

<p class="mt-2 mb-2">Call {index + 1}:</p>
<ul class="mt-2 mb-2">
  {#if canDecode}
    <li class="break-all">
      <p class="bold inline">Target:</p>
      {targetName} ({target})
    </li>
  {:else}
    <li class="break-all">
      <p class="bold inline">Target:</p>
      {target}
    </li>
  {/if}
  <li>
    <p class="bold inline">Value:</p>
    {value}
  </li>
  {#if canDecode}
    <li>
      <p class="bold inline">Calldata:</p>
      <ul>
        <li>
          <p class="bold inline">Function name:</p>
          {transactionData.name}
        </li>
        <li>
          <p class="bold inline">Input arguments:</p>
          <ul>
            {#each transactionData.args as args}
              <li class="break-all">{args}</li>
            {/each}
          </ul>
        </li>
      </ul>
    </li>
  {:else}
    <li class="break-all">
      <p class="bold inline">Calldata:</p>
      {calldata}
    </li>
  {/if}
</ul>
