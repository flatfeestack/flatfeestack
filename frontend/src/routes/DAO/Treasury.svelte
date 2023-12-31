<script lang="ts">
  import { EventLog, formatEther } from "ethers";
  import { onDestroy, onMount } from "svelte";
  import { navigate } from "svelte-routing";
  import Navigation from "../../components/DAO/Navigation.svelte";
  import {
    currentBlockNumber,
    daoConfig,
    membershipStatusValue,
    walletContract,
  } from "../../ts/daoStore";
  import { provider, userEthereumAddress } from "../../ts/ethStore";
  import { error, isSubmitting } from "../../ts/mainStore";
  import { checkUndefinedProvider } from "../../utils/ethHelpers";
  import formatDateTime from "../../utils/formatDateTime";
  import { secondsPerBlock } from "../../utils/futureBlockDate";
  import truncateEthAddress from "../../utils/truncateEthereumAddress";

  enum TransactionEventType {
    AcceptPayment,
    IncreaseAllowance,
    WithdrawFunds,
  }

  interface TransactionEvent {
    amount: bigint;
    blockDate: string;
    blockNumber: number;
    source: string; // wallet
    type: TransactionEventType;
  }

  let availableFunds: bigint = 0n;
  let transactions: TransactionEvent[] | null = null;
  let totalBalance: bigint = 0n;
  let totalAllowance: bigint = 0n;

  checkUndefinedProvider();

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
    if ($membershipStatusValue != 3n) {
      $error = "You are not allowed to view this page.";
      navigate("/dao/home");
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
        transactionsEvents.flat().map(async (event: EventLog) => {
          let eventType;

          if (event.eventName === "IncreaseAllowance") {
            eventType = TransactionEventType.IncreaseAllowance;
          } else if (event.eventName === "AcceptPayment") {
            eventType = TransactionEventType.AcceptPayment;
          } else {
            eventType = TransactionEventType.WithdrawFunds;
          }

          const blockTimestamp = (await $provider.getBlock(event.blockNumber))
            .timestamp;

          return {
            amount: event.args[1],
            blockDate: formatDateTime(new Date(blockTimestamp * 1000)),
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

  onDestroy(() => {
    $isSubmitting = false;
  });
</script>

<Navigation requiresChainId={$daoConfig?.chainId}>
  <h1 class="text-secondary-900">Treasury</h1>

  <ul>
    <li>Total balance: {formatEther(totalBalance)} ETH</li>
    <li>
      Total funds to be claimed: {formatEther(totalAllowance)} ETH
    </li>
    <li>
      Available funds: {formatEther(totalBalance - totalAllowance)} ETH
    </li>
  </ul>

  <h2>Withdraw funds</h2>
  {#if availableFunds > 0n}
    <p>
      You have {formatEther(availableFunds)} ETH available to be withdrawn.
    </p>

    <button on:click={() => withdrawFunds()} class="button4">Withdraw</button>
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
            <td>{formatEther(transaction.amount)} ETH</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else}
    No transactions in the last 90 days.
  {/if}
</Navigation>
