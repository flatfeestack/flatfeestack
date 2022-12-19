<script lang="ts">
  import { BigNumber, ethers } from "ethers";
  import { onMount } from "svelte";
  import { navigate } from "svelte-routing";
  import {
    currentBlockNumber,
    membershipStatusValue,
    provider,
    userEthereumAddress,
    walletContract,
  } from "../../ts/daaStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import formateDateTime from "../../utils/formatDateTime";
  import { secondsPerBlock } from "../../utils/futureBlockDate";
  import truncateEthAddress from "../../utils/truncateEthereumAddress";
  import Navigation from "./Navigation.svelte";

  enum TransactionEventType {
    AcceptPayment,
    IncreaseAllowance,
    WithdrawFunds,
  }

  interface TransactionEvent {
    amount: BigNumber;
    blockDate: string;
    blockNumber: number;
    source: string; // wallet
    type: TransactionEventType;
  }

  let availableFunds: BigNumber = BigNumber.from("0");
  let transactions: TransactionEvent[] | null = null;
  let totalBalance: BigNumber = BigNumber.from("0");
  let totalAllowance: BigNumber = BigNumber.from("0");

  currentBlockNumber.subscribe(async (_currentBlockNumber) => {
    await prepareView();
  });

  provider.subscribe(async (_provider) => {
    await prepareView();
  });

  walletContract.subscribe(async (_walletContract) => {
    await prepareView();
  });

  userEthereumAddress.subscribe(async (_userEthereumAddress) => {
    await prepareView();
  });

  onMount(() => {
    if ($membershipStatusValue != 3) {
      $error = "You are not allowed to review this page.";
      navigate("/daa/votes");
      return;
    }

    $isSubmitting = true;
  });

  async function prepareView() {
    if (
      $currentBlockNumber === null ||
      $provider === null ||
      $userEthereumAddress === null ||
      $walletContract === null
    ) {
      $isSubmitting = true;
      return;
    }

    const startingBlock =
      $currentBlockNumber - (60 * 60 * 24 * 7 * 4 * 3) / secondsPerBlock;

    const transactionsEvents = await Promise.all([
      $walletContract.queryFilter(
        $walletContract.filters.IncreaseAllowance(),
        startingBlock,
        $currentBlockNumber
      ),
      $walletContract.queryFilter(
        $walletContract.filters.AcceptPayment(),
        startingBlock,
        $currentBlockNumber
      ),
      $walletContract.queryFilter(
        $walletContract.filters.WithdrawFunds(),
        startingBlock,
        $currentBlockNumber
      ),
    ]);

    transactions = (
      await Promise.all(
        transactionsEvents.flat().map(async (event) => {
          let eventType;

          if (event.event === "IncreaseAllowance") {
            eventType = TransactionEventType.IncreaseAllowance;
          } else if (event.event === "AcceptPayment") {
            eventType = TransactionEventType.AcceptPayment;
          } else {
            eventType = TransactionEventType.WithdrawFunds;
          }

          const blockTimestamp = (await $provider.getBlock(event.blockNumber))
            .timestamp;

          return {
            amount: event.args[1],
            blockDate: formateDateTime(new Date(blockTimestamp * 1000)),
            blockNumber: event.blockNumber,
            source: event.args[0],
            type: eventType,
          };
        })
      )
    )
      .sort((a, b) => a.blockNumber - b.blockNumber)
      .reverse();

    [availableFunds, totalBalance, totalAllowance] = await Promise.all([
      $walletContract.allowance($userEthereumAddress),
      $walletContract.totalBalance(),
      $walletContract.totalAllowance(),
    ]);

    $isSubmitting = false;
  }

  async function withdrawFunds() {
    await $walletContract.withdrawMoney($userEthereumAddress);
  }
</script>

<Navigation>
  <h1 class="text-secondary-900">Treasury</h1>

  <ul>
    <li>Total balance: {ethers.utils.formatEther(totalBalance)} ETH</li>
    <li>
      Total funds to be claimed: {ethers.utils.formatEther(totalAllowance)} ETH
    </li>
    <li>
      Available funds: {ethers.utils.formatEther(
        totalBalance.sub(totalAllowance)
      )} ETH
    </li>
  </ul>

  <h2>Withdraw funds</h2>
  {#if availableFunds.gt(BigNumber.from("0"))}
    <p>
      You have {ethers.utils.formatEther(availableFunds)} ETH available to be withdrawn.
    </p>

    <button on:click={() => withdrawFunds()} class="button1">Withdraw</button>
  {:else}
    You have no funds available.
  {/if}

  <h2>Activities in the last 90 days</h2>

  {#if transactions?.length > 0}
    <table>
      <thead>
        <tr>
          <th>Date</th>
          <th>Block number</th>
          <th>Type</th>
          <th>Source</th>
          <th>Amount</th>
        </tr>
      </thead>
      <tbody>
        {#each transactions as transaction}
          <tr>
            <td>{transaction.blockDate}</td>
            <td>{transaction.blockNumber}</td>
            <td
              >{#if transaction.type === TransactionEventType.AcceptPayment}
                Payment received
              {:else if transaction.type === TransactionEventType.IncreaseAllowance}
                Allowance increased
              {:else}
                Funds withdrawn
              {/if}</td
            >
            <td>{truncateEthAddress(transaction.source)}</td>
            <td>{ethers.utils.formatEther(transaction.amount)} ETH</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else}
    No transactions in the last 90 days.
  {/if}
</Navigation>
