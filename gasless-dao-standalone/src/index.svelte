<script lang="ts">
    import { onMount } from "svelte";
    import { writable, get } from 'svelte/store';
    import {
        handleWalletConnect, getPaymasterDeposit,
        getGasCost, waitForConfirmations,
        incrementCounter, waitForTransactionReceipt,
        disconnectWallet, autoConnectIfAuthorized,
        checkIsMember, getBalance, fundPaymaster,
        delegateVotesToSmartAccount
    } from './lib/aaClient';
    import { loadAllProposals, voteOnProposal, createProposal, getProposal } from "./lib/dao";
    import { pushStatus, statusFeed } from "./lib/statusfeed";
    import { blockToDate } from "./lib/blockToDate";
    import { ProposalView, ProposalViewExtended } from "./lib/types";

    const eoa = writable<`0x${string}` | null>(null);
    const smartAccount = writable<`0x${string}` | null>(null);
    const lastTxHash = writable<`0x${string}` | null>(null);
    const loading = writable<boolean>(false);
    const gasCost = writable<string | null>(null);
    const paymasterFunds = writable<string | null>(null);
    const usePaymaster = writable<Boolean | null>(null);
    const eoaBalance = writable<string | null>(null);
    const smartAccountBalance = writable<string | null>(null);
    const proposals = writable<ProposalViewExtended[]>([]);

    onMount(async () => {
        const res = await autoConnectIfAuthorized();
        eoa.set(res.eoa);
        smartAccount.set(res.smartAccountAddress);
        usePaymaster.set(res.paymasterUsed);

        await getBalances();

        await refreshProposals();
    });

    let newDescription = "";
    let newBylawsHash = "";
    let fundPaymasterAmount = "";

    async function refreshProposals() {
        pushStatus("Loading proposals ...");
        try{
            const proposalList = await loadAllProposals();
            console.log("# of retrieved proposals: " + proposalList.length);
            proposals.set([]);
            const extendedList = [];

            for (const p of proposalList) {
                let prop = await getProposal(p.id);
                const extended: ProposalViewExtended = {
                    ...prop,
                    startDate: new Date(Number(prop.startTime) * 1000),
                    endDate: new Date(Number(prop.endTime) * 1000),
                };
                console.log("Pushing: " + extended);
                extendedList.push(extended);
            }

            proposals.set(extendedList);
            pushStatus("Proposals loaded");
        } catch (err) {
            console.error(err);
            pushStatus("Failed to load proposals");
        }
    }

    async function getBalances(){
        if($eoa) eoaBalance.set(await getBalance($eoa));
        if($smartAccount) smartAccountBalance.set(await getBalance($smartAccount));
    }

    async function submitProposal() {
        if (!newDescription || !newBylawsHash) return;
        await createProposal(newDescription, Number(newBylawsHash));
        await refreshProposals();
    }

    async function vote(id, support) {
        pushStatus("Casting vote ...");
        await voteOnProposal(id, support);
        await refreshProposals();
    }

    async function handleConnect() {
        loading.set(true);

        try {
            const res = await handleWalletConnect();
            eoa.set(res.eoa);
            smartAccount.set(res.smartAccountAddress);
            usePaymaster.set(res.paymasterUsed);

            await getBalances();
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
        eoaBalance.set(null);
        smartAccountBalance.set(null);
    }

    async function handleIncrementCounter() {
        loading.set(true);
        pushStatus("Preparing UserOperation ...");
        gasCost.set(null);
        paymasterFunds.set(null);

        try {
            const { userOpHash, receipt } = await incrementCounter();

            lastTxHash.set(receipt.receipt.transactionHash);
            pushStatus("UserOperation submitted");

            const txReceipt = await waitForTransactionReceipt(get(lastTxHash));
            pushStatus("Transaction included in a block");

            const cost = await getGasCost(get(lastTxHash));
            gasCost.set(cost.totalCostEth);

            const paymasterDeposit = await getPaymasterDeposit();
            paymasterFunds.set(paymasterDeposit.eth);
            await getBalances();

            await waitForConfirmations(txReceipt, 3);

            pushStatus("UserOperation Success");
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error occurred');
        } finally {
            loading.set(false);
        }
    }

    async function checkIsDAOMember(){
        console.log("Checking if member")
        let result = await checkIsMember(get(smartAccount), get(eoa));
        console.log("IsMember: " + result);
        pushStatus("Is Member " + String(result));
    }

    async function handleFundPaymaster(){
        await fundPaymaster(fundPaymasterAmount);
    }

    async function handleDelegateToSmartAccount(){
        await delegateVotesToSmartAccount($smartAccount);
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

        <section class="new-proposal">
            <h2>Create New Proposal</h2>

            <input
                bind:value={newDescription}
                placeholder="Proposal Description"
            />

            <input
                bind:value={newBylawsHash}
                placeholder="New Bylaws Hash (uint)"
                type="number"
            />

            <button on:click={submitProposal}>
                Submit Proposal
            </button>
        </section>

        <!--<button on:click={handleDelegateToSmartAccount}>
            Delegate To Smart Account
        </button>-->

        <button on:click={refreshProposals}>
            {loading ? "Loading..." : "Refresh"}
        </button>

        {#if $proposals.length === 0}
            <p>No proposals found.</p>
        {:else}
            <div class="proposal-list">
                {#each $proposals as p}
                    <div class="proposal-card">
                        <h2>Proposal<!-- #{p.id.toString()}--></h2>

                        <p><strong>Description:</strong> {p.description}</p>
                        <p><strong>Proposer:</strong> {p.proposer}</p>

                        <p>
                            <strong>State:</strong>
                            {#if p.proposalState === 0} Pending
                            {:else if p.proposalState === 1} Active
                            {:else if p.proposalState === 2} Canceled
                            {:else if p.proposalState === 3} Defeated
                            {:else if p.proposalState === 4} Succeeded
                            {:else if p.proposalState === 5} Queued
                            {:else if p.proposalState === 6} Expired
                            {:else if p.proposalState === 7} Executed
                            {/if}
                        </p>

                        <p><strong>Starts:</strong> {p.startDate.toLocaleString()}</p>
                        <p><strong>Ends:</strong> {p.endDate.toLocaleString()}</p>

                        <p><strong>Votes:</strong></p>
                        <ul>
                            <li>For: {p.forVotes}</li>
                            <li>Against: {p.againstVotes}</li>
                            <li>Abstain: {p.abstainVotes}</li>
                        </ul>

                        <div class="actions">
                            <button on:click={() => vote(p.id, 1)}>Vote FOR</button>
                            <button on:click={() => vote(p.id, 0)}>Vote AGAINST</button>
                            <button on:click={() => vote(p.id, 2)}>ABSTAIN</button>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}

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

            <section style="margin-top: 1rem;">
                {#if $eoaBalance}
                    <p><strong>EOA Balance:</strong> {$eoaBalance} ETH</p>
                {/if}

                {#if $smartAccountBalance}
                    <p><strong>Smart Account Balance:</strong> {$smartAccountBalance} ETH</p>
                {/if}
            </section>

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
        {#each globalThis.$statusFeed as entry}
            <div class="feed-item">{entry}</div>
        {/each}
    </div>
</div>

<style>
    .new-proposal, .proposal-list { margin-bottom: 2rem; }

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

    .proposal-list {
        margin-top: 1.5rem;
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .proposal-card {
        padding: 1rem;
        border-radius: 8px;
        border: 1px solid #ddd;
        background: #fafafa;
    }
</style>