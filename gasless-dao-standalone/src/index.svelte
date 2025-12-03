<script lang="ts">
    import { onMount } from "svelte";
    import { writable, get } from 'svelte/store';
    import {
        handleWalletConnect, getPaymasterDeposit,
        getGasCost, waitForConfirmations,
        incrementCounter, waitForTransactionReceipt,
        disconnectWallet, autoConnectIfAuthorized,
        checkIsMember
    } from './lib/aaClient';

    const statusFeed = writable<string[]>([]);
    const eoa = writable<`0x${string}` | null>(null);
    const smartAccount = writable<string | null>(null);
    const lastTxHash = writable<`0x${string}` | null>(null);
    const loading = writable<boolean>(false);
    const gasCost = writable<string | null>(null);
    const paymasterFunds = writable<string | null>(null);
    const usePaymaster = writable<Boolean | null>(null);

    onMount(async () => {
        const res = await autoConnectIfAuthorized();
        eoa.set(res.eoa);
        smartAccount.set(res.smartAccountAddress);
        usePaymaster.set(res.paymasterUsed);
    });

    function formatTime(date = new Date()) {
        return date.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
        });
    }

    function pushStatus(msg: string) {
        const stampedMsg = `[${formatTime()}] ${msg}`;
        statusFeed.update((l) => [...l, stampedMsg]);
    }

    async function handleConnect() {
        loading.set(true);

        try {
            const res = await handleWalletConnect();
            eoa.set(res.eoa);
            smartAccount.set(res.smartAccountAddress);
            usePaymaster.set(res.paymasterUsed);
        } catch (err: any) {
            console.error(err);
        } finally {
            loading.set(false);
        }
    }

    async function handleDisconnect(){
        disconnectWallet();

        lastTxHash.set(null);
        eoa.set(null);
        gasCost.set(null);
        loading.set(false);
        smartAccount.set(null);
        paymasterFunds.set(null);
        usePaymaster.set(null);
    }

    async function handleIncrementCounter() {
        loading.set(true);
        pushStatus("Preparing UserOperation ...");
        gasCost.set(null);
        paymasterFunds.set(null);

        try {
            const { userOpHash, receipt } = await incrementCounter(
                (text) => pushStatus(text)
            );

            lastTxHash.set(receipt.receipt.transactionHash);
            pushStatus("UserOperation submitted");

            const txReceipt = await waitForTransactionReceipt(get(lastTxHash));
            pushStatus("Transaction included in a block");

            const cost = await getGasCost(get(lastTxHash));
            gasCost.set(cost.totalCostEth);

            const paymasterDeposit = await getPaymasterDeposit();
            paymasterFunds.set(paymasterDeposit.eth);

            await waitForConfirmations(txReceipt, 3,
                (text) => pushStatus(text)
            );

            pushStatus("UserOperation Success");
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error occurred');
        } finally {
            loading.set(false);
        }
    }

    async function checkIsDAOMember(){
        let result = await checkIsMember(get(smartAccount) as '0x{string}', get(eoa));
        pushStatus("Is Member " + String(result));
    }

    async function createProposal() {
        //TODO tbd
    }

    async function voteOnProposal() {
        //TODO tbd
    }
</script>

<div class="app-layout">
    <div class="top-right">
        {#if !$eoa}
            <button on:click={handleConnect} disabled={$loading} class="btn btn-secondary">
                Connect
            </button>
        {:else}
            <button on:click={handleDisconnect} disabled={$loading} class="btn btn-secondary">
                Disconnect
            </button>

            <div class="connection-status">
                Connected as<br />
                <span>{$eoa}</span>
            </div>
        {/if}
    </div>

    <div class="container">
        <h1>Paymaster DAO</h1>

        <div class="controls">
            {#if $smartAccount}
                <section style="margin-top: 1rem;">
                    <p>
                        <strong>Smart account:</strong> {$smartAccount}<br />
                        <strong>Using Paymaster:</strong> {String($usePaymaster)}
                    </p>

                    <div class="button-column">
                        <button on:click={checkIsDAOMember} disabled={$loading} class="btn btn-primary">
                            Check Is DAO Member
                        </button>

                        <button on:click={handleIncrementCounter} disabled={$loading} class="btn btn-primary">
                            [TEST] Increment Counter
                        </button>
                    </div>
                </section>
            {/if}

            {#if $lastTxHash}
                <section style="margin-top: 1rem;">
                    <p>
                        Last tx hash:
                        <a href={`https://sepolia.etherscan.io/tx/${$lastTxHash}`} target="_blank">
                            View on Etherscan
                        </a>
                    </p>

                    {#if $gasCost}
                        <p><strong>Gas cost:</strong> {$gasCost} ETH</p>
                    {/if}

                    {#if $paymasterFunds}
                        <p><strong>Remaining Paymaster Funds:</strong> {$paymasterFunds} ETH</p>
                    {/if}
                </section>
            {/if}
        </div>
    </div>

    <div class="status-feed">
        {#each $statusFeed as entry}
            <div class="feed-item">{entry}</div>
        {/each}
    </div>
</div>


<style>
.app-layout {
    display: grid;
    grid-template-columns: 1fr 300px;
    gap: 20px;
    position: relative;
}

.top-right {
    position: absolute;
    top: 20px;
    right: 20px;
    text-align: right;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 8px;
}

.connection-status {
    font-size: 12px;
    color: #555;
}
.connection-status span {
    font-size: 12px;
    word-break: break-all;
}

.status-feed {
    margin-top: 80px;
    height: calc(100vh - 120px);
    overflow-y: auto;
    padding: 12px;
    border-left: 1px solid #ddd;
    background: #fafafa;
    font-size: 14px;
}

.feed-item {
    padding: 8px 0;
    border-bottom: 1px solid #eee;
    word-break: break-word;
}

.button-column {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.container {
    max-width: 520px;
    padding: 24px;
    border-radius: 14px;
    background: #ffffff;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.08);
    font-family: system-ui, sans-serif;
    margin-top: 60px;
}

.btn {
    padding: 10px 16px;
    border-radius: 8px;
    border: none;
    cursor: pointer;
    font-size: 14px;
    font-weight: 600;
}

.btn-primary {
    background: #4caf50;
    color: white;
}

.btn-secondary {
    background: #444;
    color: white;
}

.btn-primary:hover {
    background: #419445;
}
</style>