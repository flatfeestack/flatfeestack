<script lang="ts">
    import { onMount } from "svelte";
    import { writable, get } from "svelte/store";
    import { createLogger } from "./lib/logger";
    import { createAAContext, type AAContext } from "./lib/aa/context";
    import { listMembershipTokens, renewMembership } from "./lib/aa/membership";
    import { createProposal } from "./lib/aa/proposals";
    import { voteOnProposal, loadAllProposals } from "./lib/dao/index";
    import { getPaymasterDeposit, isMember } from "./lib/aa/paymaster";
    import { sendUserOp } from "./lib/viem/write";
    import { finalizeUserOp } from "./lib/aaClient";
    import { COUNTER_CONTRACT_ADDRESS, NFT_COUNCIL_1, NFT_COUNCIL_2 } from "./config";
    import { executeVotingDelay, prepareVotingDelaySignature, executeMint, prepareMintSignature } from "./lib/dao/council";
    import { encodeFunctionData, formatEther } from "viem";

    const COUNTER_ABI = [
        { name: "increment", type: "function", stateMutability: "nonpayable", inputs: [], outputs: [] }
    ] as const;

    const statusFeed = writable<string[]>([]);
    const eoa = writable<`0x${string}` | null>(null);
    const ctx = writable<AAContext | null>(null);

    const proposals = writable<any[]>([]);
    const membershipTokens = writable<any[]>([]);
    const selectedTokenId = writable<string | null>(null);

    const eoaBalance = writable<string | null>(null);
    const smartAccountBalance = writable<string | null>(null);
    const usingPaymaster = writable<boolean>(false);

    const lastTxHash = writable<`0x${string}` | null>(null);
    const gasCost = writable<string | null>(null);
    const paymasterFunds = writable<string | null>(null);
    const loading = writable(false);
    const retrieving = writable(false);
    const showProposalForm = writable(false);

    const proposalTitle = writable("");
    const proposalDescription = writable("");

    let isDebug = false;
    let mintTargetAddress = "";

    onMount(() => {
        if (typeof window !== "undefined") {
            isDebug = window.location.hash === "#debug";
        }

        tryAutoConnect();
    });

    let log = createLogger((msg: string) => {
        statusFeed.update((l) => [...l, `[${new Date().toLocaleTimeString()}] ${msg}`]);
    });

    function shortAddress(addr: string, first = 6, last = 4) {
        return `${addr.slice(0, first)}...${addr.slice(-last)}`;
    }

    function pushStatus(msg: string) {
        log.ui(msg);
    }

    function resetStatusFeed(){
        statusFeed.set([]);
    }

    async function tryAutoConnect() {
        if (typeof window === "undefined") return;

        const ethereum = (window as any).ethereum;
        if (!ethereum) {
            log.warn("MetaMask not found — cannot auto-connect");
            return;
        }

        try {
            const accounts: string[] = await ethereum.request({
                method: "eth_accounts"
            });

            if (!accounts || accounts.length === 0) {
                return;
            }

            const locEoa = accounts[0] as `0x${string}`;
            eoa.set(locEoa);

            ctx.set(await createAAContext(locEoa));
            
            await loadBalances();

            log.info("Auto-connect successful.");

        } catch (err: any) {
            console.error("Auto-connect failed", err);
        }
    }

    async function loadBalances() {
        const c = get(ctx);
        if (!c) return;

        let ebal = await c.public.getBalance({ address: c.eoa });
        let sbal = await c.public.getBalance({ address: c.smartAccount });

        console.log()
        eoaBalance.set(formatEther(ebal));
        smartAccountBalance.set(formatEther(sbal));
    }

    async function handleConnect() {
        loading.set(true);
        try {
            const ethereum = (window as any).ethereum;
            if (!ethereum) throw new Error("MetaMask not found");

            await ethereum.request({
                method: "wallet_requestPermissions",
                params: [{ eth_accounts: {} }],
            });

            const accounts = await ethereum.request({ method: "eth_requestAccounts" });
            const eo = accounts[0] as `0x${string}`;

            eoa.set(eo);

            const context = await createAAContext(eo, log);
            ctx.set(context);
            usingPaymaster.set(context.usePaymaster);

            await loadBalances();
        } catch (err) {
            log.error("Connection failed", err);
        } finally {
            loading.set(false);
        }
    }

    function handleDisconnect() {
        eoa.set(null);
        ctx.set(null);
        proposals.set([]);
        membershipTokens.set([]);
        eoaBalance.set(null);
        smartAccountBalance.set(null);
        lastTxHash.set(null);
        gasCost.set(null);
        paymasterFunds.set(null);
        statusFeed.set([]);
        usingPaymaster.set(false);
        log.info("Disconnected wallet");
    }

    async function handleIncrementCounter() {
        const c = get(ctx);
        if (!c) return pushStatus("Connect first");

        loading.set(true);
        pushStatus("Submitting counter increment...");

        try {
            const data = encodeFunctionData({
                abi: COUNTER_ABI,
                functionName: "increment",
                args: []
            });

            const { hash, receipt } = await sendUserOp(
                c.smartClient,
                COUNTER_CONTRACT_ADDRESS,
                data
            );

            //lastTxHash.set(receipt.receipt.transactionHash);
            pushStatus("Counter increment submitted");

            await finalizeUserOp(hash, c, {
                logger: log,
                setLastTxHash: (h) => lastTxHash.set(h),
                onBalancesUpdated: loadBalances,
                onPaymasterFunds: (h) => paymasterFunds.set(h),
                onGasCost: (h) => gasCost.set(h)
            });
            
            //const paymasterDep = await getPaymasterDeposit(c.public);
            //paymasterFunds.set(c.public.formatEther ? c.public.formatEther(paymasterDep) : String(paymasterDep));

            //await loadBalances();
        } catch (err) {
            log.error("Increment counter error", err);
        } finally {
            loading.set(false);
        }
    }

    async function handleLoadProposals() {
        const c = get(ctx);
        if (!c) return pushStatus("Connect wallet first");

        retrieving.set(true);
        loading.set(true);

        pushStatus("Loading proposals...");

        try {
            const items = await loadAllProposals(c);
            proposals.set(items);
            pushStatus(`Loaded ${items.length} proposals`);
        } catch (err) {
            log.error("Error loading proposals", err);
        } finally {
            retrieving.set(false);
            loading.set(false);
        }
    }

    async function handleCreateProposal() {
        const c = get(ctx);
        if (!c) return pushStatus("Connect first");

        loading.set(true);
        pushStatus("Preparing proposal...");

        try {
            const proposal = {
                targets: [c.smartAccount],
                values: [0n],
                calldatas: ["0x"],
                description: `${get(proposalTitle)}\n\n${get(proposalDescription)}`
            };

            const { receipt } = await createProposal(c, proposal, log);
            lastTxHash.set(receipt.receipt.transactionHash);

            pushStatus("Proposal submitted");
            showProposalForm.set(false);
            proposalTitle.set("");
            proposalDescription.set("");

            await handleLoadProposals();
            await loadBalances();
        } catch (err) {
            log.error("Proposal error", err);
        } finally {
            loading.set(false);
        }
    }

    async function handleLoadMembershipTokens() {
        const c = get(ctx);
        const eo = get(eoa);
        if (!c || !eo) return pushStatus("Connect first");

        loading.set(true);
        pushStatus("Loading membership tokens...");

        try {
            const tokens = await listMembershipTokens(c, eo);
            membershipTokens.set(tokens);

            if (tokens.length > 0)
                selectedTokenId.set(tokens[0].tokenId.toString());

            pushStatus(`Loaded ${tokens.length} token(s)`);
        } catch (err) {
            log.error("Token load error", err);
        } finally {
            loading.set(false);
        }
    }

    async function handleRenewMembership() {
        const tokenId = get(selectedTokenId);
        const c = get(ctx);
        if (!tokenId || !c) return pushStatus("Select a token");

        loading.set(true);
        try {
            await renewMembership(c, BigInt(tokenId), log);
            pushStatus("Membership renewed");
            await handleLoadMembershipTokens();
        } catch (err) {
            log.error("Renew membership error", err);
        } finally {
            loading.set(false);
        }
    }

    async function handleVote(id: bigint, support: number) {
        const c = get(ctx);
        if (!c) return pushStatus("Connect first");

        loading.set(true);
        pushStatus("Submitting vote...");

        try {
            await voteOnProposal(c, id, support, log);
            pushStatus("Vote submitted");
            await handleLoadProposals();
        } catch (err) {
            log.error("Vote error", err);
        } finally {
            loading.set(false);
        }
    }

    async function handleCheckIsDAOMember() {
        const c = get(ctx);
        if (!c) return pushStatus("Connect wallet first");

        loading.set(true);

        try {
            const member = await isMember(c.public, c.smartAccount, c.eoa);
            pushStatus(`DAO Membership: ${member ? "YES" : "NO"}`);
        } catch (err) {
            log.error("Membership check failed", err);
        } finally {
            loading.set(false);
        }
    }

    async function handleSetVotingDelay() {
        const c = get(ctx);
        if (!c) return pushStatus("Connect wallet first");

        const delay = 60n;
        const council1 = NFT_COUNCIL_1;
        const council2 = NFT_COUNCIL_2;

        loading.set(true);

        try {
            if (c.eoa.toLowerCase() === council2.toLowerCase()) {
                // Council2 signs
                const { signature } = await prepareVotingDelaySignature(c, delay);
                sessionStorage.setItem("council2_sig_vdelay", signature);
                pushStatus("Signature saved. Switch to Council1.");
            } else if (c.eoa.toLowerCase() === council1.toLowerCase()) {
                // Council1 executes
                const sig2 = sessionStorage.getItem("council2_sig_vdelay") as `0x${string}`;
                if (!sig2) throw new Error("Council2 signature missing!");

                await executeVotingDelay(c, delay, sig2, log);
                pushStatus("Voting delay updated!");

                sessionStorage.removeItem("council2_sig_vdelay");
            } else {
                pushStatus("You must be a council member.");
            }
        } catch (err) {
            log.error("Voting delay error", err);
        } finally {
            loading.set(false);
        }
    }

    async function handleMintNewMembership() {
        const c = get(ctx);
        if (!c) return pushStatus("Connect first");

        const newMember = mintTargetAddress as `0x${string}`;
        const tokenId = BigInt(Date.now());
        const council1 = NFT_COUNCIL_1;
        const council2 = NFT_COUNCIL_2;

        loading.set(true);

        try {
            if (c.eoa.toLowerCase() === council2.toLowerCase()) {
                const { signature } = await prepareMintSignature(c, newMember, tokenId);
                sessionStorage.setItem("debug_mint_sig2", signature);
                pushStatus("Council2 signature saved. Switch to Council1.");
            } 
            else if (c.eoa.toLowerCase() === council1.toLowerCase()) {
                const sig2 = sessionStorage.getItem("debug_mint_sig2") as `0x${string}`;
                if (!sig2) throw new Error("Missing council2 signature!");

                const sig1 = await (window as any).ethereum.request({
                    method: "personal_sign",
                    params: [/* same payloadHash */, c.eoa]
                });

                await executeMint(c, newMember, tokenId, sig1, sig2, log);

                pushStatus("Membership minted!");
                sessionStorage.removeItem("debug_mint_sig2");
            } 
            else {
                pushStatus("You must be council1 or council2");
            }
        } catch (err) {
            log.error("Mint membership error", err);
        } finally {
            loading.set(false);
        }
    }
</script>

<div class="app-layout">
    <div class="top-right">
        {#if !$eoa}
            <button on:click={handleConnect} disabled={$loading} class="btn btn-secondary">
                Connect
            </button>
        {:else}
            <div class="connection-status">
                Connected as<br />
                <span>{shortAddress($eoa)}</span>
            </div>

            <button on:click={handleDisconnect} disabled={$loading} class="btn btn-secondary">
                Disconnect
            </button>
        {/if}
    </div>

    <div class="container">
        <h1>Paymaster DAO</h1>

        <div class="controls">
            {#if $ctx}
                <section style="margin-top: 1rem;">
                    <p>
                        {#if $eoaBalance}
                            <strong>EOA Balance:</strong> {$eoaBalance} ETH <br/>
                        {/if}

                        <strong>Smart account:</strong> {shortAddress($ctx.smartAccount)} <br/>
                        {#if $smartAccountBalance}
                            <strong>Smart Account Balance:</strong> {$smartAccountBalance} ETH <br/>
                        {/if}
                        
                        <label for="paymasterToggle"><strong>Using Paymaster:</strong></label>

                        <label class="switch">
                            <input
                                id="paymasterToggle"
                                type="checkbox"
                                bind:checked={$ctx.usePaymaster}
                            />
                            <span class="slider round"></span>
                        </label>

                        <span class="toggle-state">{ $ctx.usePaymaster ? "Yes" : "No" }</span>
                    </p>

                    <div class="button-column">
                        <button on:click={handleCheckIsDAOMember} disabled={$loading} class="btn btn-primary">
                            Check Is DAO Member
                        </button>

                        <button on:click={handleIncrementCounter} disabled={$loading} class="btn btn-primary">
                            [TEST] Increment Counter
                        </button>

                        <button on:click={handleLoadMembershipTokens} disabled={$loading} class="btn btn-primary">
                            Load Membership Tokens (EOA or Smart Account)
                        </button>

                        {#if $membershipTokens.length > 0}
                            <div class="form-group">
                                <label for="membership-token">Select token</label>
                                <select
                                    id="membership-token"
                                    bind:value={$selectedTokenId}
                                    disabled={$loading}
                                >
                                    {#each $membershipTokens as token}
                                        <option value={token.tokenId}>
                                            Token #{token.tokenId} -
                                            {#if token.membershipPaidUntil == 281474976710655n}
                                                valid forever (council)
                                            {:else}
                                                valid until {new Date(Number(token.membershipPaidUntil) * 1000).toLocaleDateString()}
                                            {/if}
                                        </option>
                                    {/each}
                                </select>
                            </div>
                            {#if $selectedTokenId}
                                <button on:click={handleRenewMembership} disabled={$loading} class="btn btn-primary">
                                    Renew Membership for Token
                                </button>
                            {/if}
                        {/if}

                        <div class="result-box">
                            <h3>Result</h3>
                            {#if $lastTxHash}
                                <p>
                                    <strong>Last tx hash: </strong>
                                    <a href={`https://sepolia.etherscan.io/tx/${$lastTxHash}`} target="_blank">
                                        View on Etherscan
                                    </a> <br>

                                    {#if $gasCost}
                                        <strong>Gas cost:</strong> {$gasCost} ETH <br>
                                    {/if}

                                    {#if $paymasterFunds}
                                        <strong>Remaining Paymaster Funds:</strong> {$paymasterFunds} ETH
                                    {/if}
                                </p>
                            {/if}
                        </div>

                        {#if isDebug}
                            <hr style="margin: 1rem 0; border: 1px solid #ccc;" />
                            <h4 style="margin-top: 0;">🔍 Debugging Tools</h4>

                            <label for="mintTargetInput"><strong>Mint Membership For:</strong></label>
                            <input
                                id="mintTargetInput"
                                type="text"
                                placeholder="0x1234... EOA or Smart Account address"
                                bind:value={mintTargetAddress}
                            />
                            <button
                                on:click={handleMintNewMembership}
                                class="btn btn-secondary"
                                style="background: #ff9800;">
                                Mint new Membership (Councils)
                            </button>

                            <button on:click={handleSetVotingDelay} disabled={$loading} class="btn btn-secondary" style="background: #ff9800;">
                                Set 1-Minute Voting Delay (Councils)
                            </button>
                        {/if}
                    </div>
                </section>
            {/if}            
        </div>
    </div>

    <div class="middle-section">
        <h2 style="margin-top: 0;">Governance Proposals</h2>

        <button on:click={handleLoadProposals} disabled={$loading} class="btn btn-secondary">
            {$retrieving ? "Retrieving..." : "Load All Proposals"}
        </button>

        <button on:click={() => showProposalForm.set(!$showProposalForm)} disabled={$loading} class="btn btn-primary">
            {$showProposalForm ? "Cancel" : "Create Proposal"}
        </button>
        {#if $showProposalForm}
            <section class="proposal-section" style="margin-top: 1rem; border-top: 1px solid #eee; padding-top: 1rem;">
                <h3>Create New Proposal</h3>
                
                <div class="form-group">
                    <label for="proposal-title">Title</label>
                    <input 
                        id="proposal-title"
                        type="text" 
                        bind:value={$proposalTitle}
                        placeholder="Proposal title"
                        disabled={$loading}
                    />
                </div>

                <div class="form-group">
                    <label for="proposal-description">Description</label>
                    <textarea 
                        id="proposal-description"
                        bind:value={$proposalDescription}
                        placeholder="Proposal description"
                        rows="4"
                        disabled={$loading}
                    ></textarea>
                </div>

                <div class="form-note">
                    <small>Note: This simple form creates a proposal with no actions. For complex proposals with contract calls, use custom calldata.</small>
                </div>

                <button 
                    on:click={handleCreateProposal} 
                    disabled={$loading || !$proposalTitle.trim()}
                    class="btn btn-primary"
                    style="width: 100%;"
                >
                    {$loading ? "Submitting..." : "Submit Proposal"}
                </button>
            </section>
        {/if}

        {#if $proposals.length > 0}
            <section class="proposal-section" style="margin-top: 1rem; border-top: 1px solid #eee; padding-top: 1rem;">
                <h3>Proposals ({$proposals.length})</h3>
                
                <div class="proposal-list">
                    {#each $proposals as proposal}
                        <div class="proposal-card">
                            <div class="proposal-header">
                                <span class="proposal-id">#{proposal.id.toString().slice(0, 8)}...</span>
                                <span class="proposal-state state-{proposal.proposalState}">
                                    {['Pending', 'Active', 'Canceled', 'Defeated', 'Succeeded', 'Queued', 'Expired', 'Executed'][proposal.proposalState] || 'Unknown'}
                                </span>
                            </div>
                            
                            <div class="proposal-description">
                                {proposal.description.split('\n\n')[0] || proposal.description.substring(0, 100)}
                                {#if proposal.description.length > 100}...{/if}
                            </div>
                            
                            <div class="proposal-meta">
                                <small>
                                    <strong>Proposer:</strong> {proposal.proposer.slice(0, 6)}...{proposal.proposer.slice(-4)}
                                </small>
                                <small>
                                    <strong>Voting:</strong> 
                                    For: {proposal.forVotes.toString()} | 
                                    Against: {proposal.againstVotes.toString()} | 
                                    Abstain: {proposal.abstainVotes.toString()}
                                </small>
                                <small>
                                    <strong>Start:</strong> {new Date(Number(proposal.startTime) * 1000).toLocaleDateString()}
                                </small>
                                <small>
                                    <strong>End:</strong> {new Date(Number(proposal.endTime) * 1000).toLocaleDateString()}
                                </small>
                            </div>
                            
                            {#if proposal.proposalState === 1}
                                <div class="vote-buttons">
                                    <button 
                                        on:click={() => handleVote(proposal.id, 1)} 
                                        disabled={$loading}
                                        class="btn-vote btn-vote-for"
                                    >
                                        👍 Vote For
                                    </button>
                                    <button 
                                        on:click={() => handleVote(proposal.id, 0)} 
                                        disabled={$loading}
                                        class="btn-vote btn-vote-against"
                                    >
                                        👎 Vote Against
                                    </button>
                                    <button 
                                        on:click={() => handleVote(proposal.id, 2)} 
                                        disabled={$loading}
                                        class="btn-vote btn-vote-abstain"
                                    >
                                        🤷 Abstain
                                    </button>
                                </div>
                            {:else if proposal.proposalState === 0}
                                <div class="vote-notice">
                                    ⏳ Voting starts on {new Date(Number(proposal.startTime) * 1000).toLocaleString()}
                                </div>
                            {:else}
                                <div class="vote-notice">
                                    Voting ended
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </section>
        {/if}
    </div>

    <div class="status-feed">
        <button 
            on:click={() => resetStatusFeed()}
            class="btn"
        >
            Clear
        </button>
        {#each $statusFeed as entry}
            <div class="feed-item">{entry}</div>
        {/each}
    </div>
</div>

<style>
.app-layout {
    display: grid;
    grid-template-columns: 1fr 1fr 300px;
    gap: 20px;
    position: relative;
    min-height: 100vh;
    overflow: visible;
}

.top-right {
    position: absolute;
    top: 20px;
    right: 20px;
    text-align: right;
    display: flex;
    flex-direction: row;
    align-items: flex-end;
    gap: 10px;
}

.connection-status {
    font-size: 14px;
    color: #555;
}
.connection-status span {
    font-size: 14px;
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

h1 {
    margin-top: 0;
}

.container {
    max-width: 520px;
    padding: 24px;
    border-radius: 14px;
    background: #ffffff;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.08);
    font-family: system-ui, sans-serif;
    margin-top: 20px;
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

.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 600;
    font-size: 14px;
}

.form-group input,
.form-group textarea {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ddd;
    border-radius: 6px;
    font-family: system-ui, sans-serif;
    font-size: 14px;
    box-sizing: border-box;
}

.form-group input:disabled,
.form-group textarea:disabled {
    background-color: #f5f5f5;
    cursor: not-allowed;
}

.form-group textarea {
    resize: vertical;
}

.form-note {
    margin: 1rem 0;
    padding: 8px 12px;
    background-color: #f0f0f0;
    border-left: 3px solid #999;
    border-radius: 4px;
    font-size: 12px;
    color: #666;
}

.middle-section {
    display: flex;
    flex-direction: column;
    width: 100%;
    margin-top: 1rem;
    border-radius: 14px;
    background: #ffffff;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.08);
    font-family: system-ui, sans-serif;
    margin-top: 60px;
    max-height: 100vh;
    overflow-y: auto;
    padding-right: 8px;
}

.middle-section::-webkit-scrollbar {
    width: 8px;
}

.middle-section::-webkit-scrollbar-thumb {
    background: #ccc;
    border-radius: 4px;
}

.proposal-section {
    width: 100%;
    max-width: 600px;
}

.proposal-card {
    background: #f9f9f9;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    padding: 12px;
    margin-bottom: 12px;
}

.proposal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
}

.proposal-id {
    font-family: monospace;
    font-size: 12px;
    color: #666;
}

.proposal-state {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
}

.state-0 { background: #ffc107; color: #000; } /* Pending */
.state-1 { background: #2196f3; color: #fff; } /* Active */
.state-2 { background: #9e9e9e; color: #fff; } /* Canceled */
.state-3 { background: #f44336; color: #fff; } /* Defeated */
.state-4 { background: #4caf50; color: #fff; } /* Succeeded */
.state-5 { background: #ff9800; color: #fff; } /* Queued */
.state-6 { background: #757575; color: #fff; } /* Expired */
.state-7 { background: #8bc34a; color: #fff; } /* Executed */

.proposal-description {
    margin-bottom: 8px;
    font-size: 14px;
    line-height: 1.4;
    color: #333;
}

.proposal-meta {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 12px;
    color: #666;
    border-top: 1px solid #e0e0e0;
    padding-top: 8px;
}

.proposal-meta small {
    display: block;
}

.vote-buttons {
    display: flex;
    gap: 8px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #e0e0e0;
}

.btn-vote {
    flex: 1;
    padding: 8px 12px;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.2s;
}

.btn-vote:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.btn-vote-for {
    background: #4caf50;
    color: white;
}

.btn-vote-for:hover:not(:disabled) {
    background: #419445;
}

.btn-vote-against {
    background: #f44336;
    color: white;
}

.btn-vote-against:hover:not(:disabled) {
    background: #d32f2f;
}

.btn-vote-abstain {
    background: #9e9e9e;
    color: white;
}

.btn-vote-abstain:hover:not(:disabled) {
    background: #757575;
}

.vote-notice {
    margin-top: 12px;
    padding: 8px 12px;
    background: #fff3cd;
    border: 1px solid #ffc107;
    border-radius: 6px;
    font-size: 13px;
    text-align: center;
    color: #856404;
}

.switch {
    position: relative;
    display: inline-block;
    width: 46px;
    height: 24px;
}

/* Hide HTML checkbox */
.switch input {
    opacity: 0;
    width: 0;
    height: 0;
}

/* The slider UI */
.slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #aaa;
    transition: 0.2s;
    border-radius: 24px;
}

.slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: 0.2s;
    border-radius: 50%;
}

/* Checked state */
input:checked + .slider {
    background-color: #4CAF50;
}

input:checked + .slider:before {
    transform: translateX(22px);
}

.toggle-state {
    min-width: 32px;
}

.proposal-list {
    max-height: 60vh;
    overflow-y: auto;
    padding-right: 8px;
}

.proposal-list::-webkit-scrollbar {
    width: 8px;
}

.proposal-list::-webkit-scrollbar-thumb {
    background: #ccc;
    border-radius: 4px;
}

.result-box {
    margin-top: 20px;
    padding: 16px;
    background: #ffffff;
    border-radius: 10px;
    box-shadow: 0 4px 10px rgba(0,0,0,0.05);
    font-size: 14px;
    max-width: 520px;
}

.result-box h3 {
    margin: 0 0 10px 0;
    font-size: 18px;
    font-weight: 600;
}

.result-box p {
    margin: 8px 0;
    line-height: 1.4;
}

button:disabled,
.btn:disabled {
    opacity: 0.75 !important;
    cursor: not-allowed;
    filter: brightness(0.8);
}
</style>