<script lang="ts">
    import { onMount } from "svelte";
    import { writable } from 'svelte/store';
    import {
        handleWalletConnect, getPaymasterDeposit,
        getGasCost, waitForConfirmations,
        incrementCounter, waitForTransactionReceipt,
        checkIsMember, disconnectWallet,
        autoConnectIfAuthorized
    } from './lib/aaClient';

    const status = writable<string>('Not connected');
    const eoa = writable<`0x${string}` | null>(null);
    const smartAccount = writable<string | null>(null);
    const lastTxHash = writable<`0x${string}` | null>(null);
    const loading = writable<boolean>(false);
    const gasCost = writable<string | null>(null);
    const paymasterFunds = writable<string | null>(null);
    const usePaymaster = writable<Boolean | null>(null);

    onMount(async () => {
        const { eoa: addr, smartAccountAddress, paymasterUsed } = await autoConnectIfAuthorized();
        eoa.set(addr);
        smartAccount.set(smartAccountAddress);
        usePaymaster.set(paymasterUsed);

        if (addr != null) status.set("Auto Connected");
	});

    async function handleConnect() {
        loading.set(true);

        try {
        const { eoa: addr, smartAccountAddress, paymasterUsed } = await handleWalletConnect();
        eoa.set(addr);
        smartAccount.set(smartAccountAddress);
        usePaymaster.set(paymasterUsed);
        if (addr != null) status.set("Connected");
        } catch (err) {
        console.error(err);
        status.set((err).message);
        } finally {
        loading.set(false);
        }
    }

    async function handleDisconnect(){
        disconnectWallet();
        lastTxHash.set(null);
        status.set(null);
        eoa.set(null);
        gasCost.set(null);
        loading.set(false);
        smartAccount.set(null);
        paymasterFunds.set(null);
    }

    async function checkIsDAOMember(){
        loading.set(true);
        status.set("pending...");
        gasCost.set(null);
        paymasterFunds.set(null);

        try {
        const isMember = await checkIsMember();
        status.set("Is Member: " + isMember);
        } catch (err) {
        console.error(err);
        status.set((err as Error).message);
        } finally {
        loading.set(false);
        }
    }

    async function handleIncrementCounter() {
        loading.set(true);
        status.set("pending...");
        gasCost.set(null);
        paymasterFunds.set(null);

        try {
        const { userOpHash, txHash } = await incrementCounter(
            (text) => status.set(text)
        );
        
        lastTxHash.set(txHash.receipt.transactionHash);
        status.set("userOp submitted");

        const receipt = await waitForTransactionReceipt($lastTxHash);
        status.set("transaction mined");

        const cost = await getGasCost($lastTxHash);
        gasCost.set(cost.totalCostEth);

        const paymasterDeposit = await getPaymasterDeposit();
        paymasterFunds.set(paymasterDeposit.eth);

        await waitForConfirmations(receipt, 3,
            (text) => status.set(text)
        );

        status.set("success");
        } catch (err) {
        console.error(err);
        status.set((err as Error).message);
        } finally {
        loading.set(false);
        }
    }

    async function createProposal() {
        //TODO tbd
    }

    async function voteOnProposal() {
        //TODO tbd
    }
</script>

<div class="container">
    <h1>Paymaster DAO</h1>

    <div class="status-box">
        <span>{$status}</span>
    </div>

    <div class="controls">
        {#if !$eoa}
            <button on:click={handleConnect} disabled={$loading} class="btn btn-primary">
            Connect MetaMask
            </button>
        {:else}
            <button on:click={handleDisconnect} disabled={$loading} class="btn btn-primary">
            Disconnect MetaMask
            </button>
        {/if}

        {#if $smartAccount}
            <section style="margin-top: 1rem;">
            <p>
                <strong>EOA:</strong> {$eoa}<br />
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
                <a
                href={`https://sepolia.etherscan.io/tx/${$lastTxHash}`}
                target="_blank"
                rel="noreferrer"
                >
                View on Etherscan
                </a>
            </p>
            </section>
            
            {#if $gasCost}
            <section style="margin-top: 1rem;">
                <p>
                <strong>Gas cost (paid by paymaster):</strong> {$gasCost} ETH
                </p>
            </section>
            {/if}
            {#if $paymasterFunds}
            <section style="margin-top: 1rem;">
                <p>
                <strong>Remaining Paymaster Funds:</strong> {$paymasterFunds} ETH
                </p>
            </section>
            {/if}
        {/if}

        <!--<h2>DAO Proposal</h2>

        <div class="input-group">
        <span>Proposal Target</span>
        <input type="text" bind:value={proposalTarget} placeholder="0xContract..." class="address-input" />
        </div>

        <div class="input-group">
        <span>Calldata (hex)</span>
        <input type="text" bind:value={proposalCalldata} placeholder="0x..." class="address-input" />
        </div>

        <div class="input-group">
        <span>Description</span>
        <input type="text" bind:value={proposalDescription} placeholder="Description" class="address-input" />
        </div>

        <button onclick={createProposal} class="btn btn-primary">
            Create Proposal
        </button>

        <hr>

        <h2>Vote on Proposal</h2>

        <div class="input-group">
        <span>Proposal ID</span>
            <input type="text" bind:value={proposalId} placeholder="e.g. 123" class="address-input" />
        </div>

        <div class="controls-row" style="display:flex; gap:10px;">
            <button onclick={() => voteOnProposal(0)} class="btn btn-secondary">Against</button>
            <button onclick={() => voteOnProposal(1)} class="btn btn-primary">For</button>
            <button onclick={() => voteOnProposal(2)} class="btn btn-secondary">Abstain</button>
        </div>-->
    </div>
</div>

<style>
    .button-column {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .container {
        max-width: 520px;
        margin: 40px auto;
        padding: 24px;
        border-radius: 14px;
        background: #ffffff;
        box-shadow: 0 4px 14px rgba(0,0,0,0.08);
        font-family: system-ui, sans-serif;
    }

    h1 {
        text-align: center;
        margin-bottom: 28px;
        font-size: 28px;
    }

    .status-box {
        margin-bottom: 24px;
        padding: 16px;
        background: #f7f7f9;
        border-radius: 10px;
        border: 1px solid #e3e3e8;
        font-size: 14px;
        line-height: 1.5;
        word-break: break-all;
    }

    .controls {
        display: flex;
        flex-direction: column;
        gap: 14px;
    }

    .btn {
        padding: 12px 20px;
        border-radius: 10px;
        border: none;
        cursor: pointer;
        font-size: 15px;
        font-weight: 600;
        transition: all 0.15s ease;
    }

    .btn-primary {
        background: #4caf50;
        color: white;
    }

    .btn-primary:hover {
        background: #419445;
    }

    .status-box span {
        display: block;
        margin-bottom: 4px;
    }
</style>