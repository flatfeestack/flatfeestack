<script lang="ts">
    import { onMount } from "svelte";
    import { writable, get } from 'svelte/store';
    import { createWalletClient, custom, encodePacked, keccak256 } from 'viem';
    import { sepolia } from 'viem/chains';
    import {
        handleWalletConnect, getPaymasterDeposit,
        getGasCost, waitForConfirmations,
        incrementCounter, waitForTransactionReceipt,
        disconnectWallet, autoConnectIfAuthorized,
        checkIsMember, getBalance, listMembershipTokens,
        createProposal, ensurePublicClient,
        renewMembership
    } from './lib/aaClient';
    import { loadAllProposals, voteOnProposal as castVote, getProposal } from "./lib/dao";
    import { ProposalViewExtended, ProposalDetails } from "./lib/types";
    import { DAO_CONTRACT_ADDRESS, NFT_CONTRACT_ADDRESS } from './config';
    import FlatFeeStackNFT from '../artifacts/contracts/FlatFeeStackNFTandDAO.sol/FlatFeeStackNFT.json' assert { type: "json" };

    const statusFeed = writable<string[]>([]);
    const eoa = writable<`0x${string}` | null>(null);
    const smartAccount = writable<`0x${string}` | null>(null);
    const lastTxHash = writable<`0x${string}` | null>(null);
    const loading = writable<boolean>(false);
    const retrieving = writable<boolean>(false);
    const gasCost = writable<string | null>(null);
    const paymasterFunds = writable<string | null>(null);
    const usePaymaster = writable<boolean | null>(null);
    const eoaBalance = writable<string | null>(null);
    const smartAccountBalance = writable<string | null>(null);
    const proposals = writable<ProposalViewExtended[]>([]);
    const showProposalForm = writable<boolean>(false);
    const proposalTitle = writable<string>("");
    const proposalDescription = writable<string>("");
    const membershipTokens = writable<{ tokenId: bigint; membershipPaidUntil: bigint; }[]>([]);
    const selectedTokenId = writable<string | null>(null);

    let isDebug = false;
    let mintTargetAddress = "";

    onMount(async () => {
        if(typeof window != "undefined"){
            isDebug = window.location.hash === "#debug";
        }

        const res = await autoConnectIfAuthorized();
        eoa.set(res.eoa);
        smartAccount.set(res.smartAccountAddress);
        usePaymaster.set(res.paymasterUsed);

        await getBalances();
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

    function resetStatusFeed(){
        statusFeed.set([]);
    }

    export function shortAddress(addr: string, first = 7, last = 5) {
        if (!addr) return "";
        return `${addr.slice(0, first)}...${addr.slice(-last)}`;
    }

    async function getBalances(){
        if($eoa) eoaBalance.set(await getBalance($eoa));
        if($smartAccount) smartAccountBalance.set(await getBalance($smartAccount));
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
        proposalTitle.set("");
        proposalDescription.set("");
        membershipTokens.set([]);
        selectedTokenId.set(null);
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
        console.log("Checking if member")
        let result = await checkIsMember(get(smartAccount), get(eoa));
        console.log("IsMember: " + result);
        pushStatus("Is Member " + String(result));
    }

    async function handleLoadProposals() {
        loading.set(true);
        retrieving.set(true);
        
        try {
            pushStatus("Loading all proposals...");
            const allProposals = await loadAllProposals();
            proposals.set(allProposals);
            pushStatus(`✓ Loaded ${allProposals.length} proposal(s)`);
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error loading proposals');
        } finally {
            loading.set(false);
            retrieving.set(false);
        }
    }

    async function handleSetVotingDelay() {
        loading.set(true);
        try {
            pushStatus("🏛️ Setting voting delay to 60 seconds...");
            
            const daoAddress = DAO_CONTRACT_ADDRESS as `0x${string}`;
            const newDelay = 60n;
            const council1 = "0x4a152972bc6fec8fd44c716b4f994090cca835d9";
            const council2 = "0x11e12a6fbd1126187502fd5430253ad189d0831f";
            const currentEOA = $eoa;
            
            pushStatus(`Connected as: ${currentEOA}`);
            pushStatus(`Is Council2? ${currentEOA === council2.toLowerCase()}`);
            pushStatus(`Is Council1? ${currentEOA === council1.toLowerCase()}`);
            
            // Step 1: Check which council member is connected
            if (currentEOA === council2.toLowerCase()) {
                // Council2 is connected - generate signature
                pushStatus("✓ Council2 detected - generating signature...");
                
                // @ts-ignore
                if (!window.ethereum) {
                    throw new Error("MetaMask not found");
                }
                
                // @ts-ignore
                const walletClient = createWalletClient({
                    chain: sepolia,
                    // @ts-ignore
                    transport: custom(window.ethereum)
                });
                
                const packed = encodePacked(
                    ["address", "string", "uint256"],
                    [daoAddress, "setVotingDelay", newDelay]
                );
                const innerHash = keccak256(packed);
                
                pushStatus(`Hash to sign: ${innerHash}`);
                
                // Use personal_sign directly to avoid viem adding extra prefix
                // The contract expects: ECDSA.recover(keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", innerHash)), signature)
                // personal_sign will add the prefix for us
                // @ts-ignore
                const signature = await window.ethereum.request({
                    method: 'personal_sign',
                    params: [innerHash, currentEOA]
                }) as `0x${string}`;
                
                pushStatus("✓ Signature generated!");
                pushStatus(`Signature: ${signature.substring(0, 20)}...`);
                pushStatus("");
                pushStatus("⚠️ Now switch to Council1 and click the button again");
                pushStatus(`Council1 address: ${council1}`);
                pushStatus("");
                pushStatus("Signature will be stored temporarily...");
                
                // Store signature in sessionStorage
                sessionStorage.setItem('council2_signature', signature);
                
            } else if (currentEOA === council1.toLowerCase()) {
                // Council1 is connected - execute with council2's signature
                pushStatus("✓ Council1 detected - executing transaction...");
                
                const signature2 = sessionStorage.getItem('council2_signature') as `0x${string}`;
                if (!signature2) {
                    throw new Error("Council2 signature not found. Please connect as Council2 first and click the button to generate signature.");
                }
                
                pushStatus("✓ Using Council2 signature from session");
                pushStatus(`Signature: ${signature2.substring(0, 20)}...`);
                
                // Verify the signature before sending
                const { recoverMessageAddress } = await import('viem');
                
                const packed2 = encodePacked(
                    ["address", "string", "uint256"],
                    [daoAddress, "setVotingDelay", newDelay]
                );
                const innerHash2 = keccak256(packed2);
                
                const recoveredAddress = await recoverMessageAddress({
                    message: { raw: innerHash2 },
                    signature: signature2
                });
                
                pushStatus(`Signature recovers to: ${recoveredAddress}`);
                pushStatus(`Expected Council2: ${council2}`);
                pushStatus(`Match? ${recoveredAddress.toLowerCase() === council2.toLowerCase()}`);
                
                if (recoveredAddress.toLowerCase() !== council2.toLowerCase()) {
                    throw new Error(`Signature verification failed! Recovered ${recoveredAddress} but expected ${council2}`);
                }
                
                // @ts-ignore
                if (!window.ethereum) {
                    throw new Error("MetaMask not found");
                }
                
                // @ts-ignore
                const walletClient = createWalletClient({
                    chain: sepolia,
                    // @ts-ignore
                    transport: custom(window.ethereum)
                });
                
                const daoAbi = [
                    {
                        name: 'setCouncilVotingDelayOverride',
                        type: 'function',
                        stateMutability: 'nonpayable',
                        inputs: [
                            { name: 'newDelay', type: 'uint256' },
                            { name: 'index1', type: 'uint256' },
                            { name: 'index2', type: 'uint256' },
                            { name: 'signature2', type: 'bytes' }
                        ],
                        outputs: []
                    },
                    {
                        name: 'votingDelay',
                        type: 'function',
                        stateMutability: 'view',
                        inputs: [],
                        outputs: [{ type: 'uint256' }]
                    }
                ] as const;
                
                const publicClient = ensurePublicClient();
                
                // Double-check both council members are valid
                const nftAddress = await publicClient.readContract({
                    address: daoAddress,
                    abi: [{ name: 'token', type: 'function', stateMutability: 'view', inputs: [], outputs: [{ type: 'address' }] }],
                    functionName: 'token'
                }) as `0x${string}`;
                
                pushStatus(`NFT contract: ${nftAddress}`);
                
                const isCouncil1 = await publicClient.readContract({
                    address: nftAddress,
                    abi: [{ name: 'isCouncilIndex', type: 'function', stateMutability: 'view', inputs: [{ type: 'address' }, { type: 'uint256' }], outputs: [{ type: 'bool' }] }],
                    functionName: 'isCouncilIndex',
                    args: [currentEOA as `0x${string}`, 0n]
                });
                
                const isCouncil2 = await publicClient.readContract({
                    address: nftAddress,
                    abi: [{ name: 'isCouncilIndex', type: 'function', stateMutability: 'view', inputs: [{ type: 'address' }, { type: 'uint256' }], outputs: [{ type: 'bool' }] }],
                    functionName: 'isCouncilIndex',
                    args: [recoveredAddress as `0x${string}`, 0n]
                });
                
                pushStatus(`Council1 (${currentEOA}) valid at index 0? ${isCouncil1}`);
                pushStatus(`Council2 (${recoveredAddress}) valid at index 0? ${isCouncil2}`);
                
                if (!isCouncil1 || !isCouncil2) {
                    throw new Error("One or both council members not valid at index 0");
                }
                
                pushStatus("Sending transaction...");
                const txHash = await walletClient.writeContract({
                    address: daoAddress,
                    abi: daoAbi,
                    functionName: 'setCouncilVotingDelayOverride',
                    args: [newDelay, 0n, 0n, signature2],
                    account: currentEOA,
                    gas: 500000n
                });
                
                pushStatus(`Transaction sent: ${txHash}`);
                pushStatus("Waiting for confirmation...");
                
                const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
                
                if (receipt.status === 'success') {
                    const actualDelay = await publicClient.readContract({
                        address: daoAddress,
                        abi: daoAbi,
                        functionName: 'votingDelay'
                    });
                    
                    pushStatus("✅ SUCCESS!");
                    pushStatus(`Voting delay now: ${actualDelay.toString()} seconds`);
                    pushStatus("New proposals will start voting in ~1 minute!");
                    
                    // Clear the stored signature
                    sessionStorage.removeItem('council2_signature');
                } else {
                    throw new Error("Transaction failed");
                }
                
            } else {
                throw new Error(`Please connect as Council1 (${council1}) or Council2 (${council2})`);
            }
            
            console.log("Council execution completed");
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error preparing voting delay change');
        } finally {
            loading.set(false);
        }
    }

    export async function handleMintNewMembership(newMember: `0x${string}`, tokenId: bigint) {
        const nftAddress = NFT_CONTRACT_ADDRESS as `0x${string}`;
        const council1 = "0x4a152972bc6fec8fd44c716b4f994090cca835d9"; 
        const council2 = "0x11e12a6fbd1126187502fd5430253ad189d0831f"; 

        const publicClient = ensurePublicClient();
        const currentEOA = $eoa;

        const payloadHash = keccak256(
            encodePacked(
                ["address", "string", "address", "string", "uint256"],
                [nftAddress, "safeMint", newMember, "#", tokenId]
            )
        );

        // Council2 signs first
        if (currentEOA === council2.toLowerCase()) {
            const sig2 = await window.ethereum.request({
                method: "personal_sign",
                params: [payloadHash, currentEOA]
            });

            sessionStorage.setItem("debug_mint_sig2", sig2);
            alert("Council2 signature saved! Now switch wallet to Council1.");
            return;
        }

        // Council1 executes mint
        if (currentEOA === council1.toLowerCase()) {
            const sig2 = sessionStorage.getItem("debug_mint_sig2");
            if (!sig2) throw new Error("Council2 signature missing!");

            // Council1 signs
            const sig1 = await window.ethereum.request({
                method: "personal_sign",
                params: [payloadHash, currentEOA]
            });

            const walletClient = createWalletClient({
                chain: sepolia,
                transport: custom(window.ethereum)
            });

            const txHash = await walletClient.writeContract({
                address: nftAddress,
                abi: FlatFeeStackNFT.abi,
                functionName: "safeMint",
                args: [newMember, 1n, sig1, 2n, sig2],
                account: currentEOA,
                gas: 500000n
            });

            const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });

            sessionStorage.removeItem("debug_mint_sig2");
            return receipt;
        }

        throw new Error(`Connect as Council1 (${council1}) or Council2 (${council2})`);
    }

    async function handleCreateProposal() {
        loading.set(true);
        gasCost.set(null);
        paymasterFunds.set(null);

        try {
            pushStatus("Preparing proposal UserOperation ...");
            
            const proposal: ProposalDetails = {
                targets: [get(smartAccount) as `0x${string}`],
                values: [0n],
                calldatas: ['0x'],
                description: `${get(proposalTitle)}\n\n${get(proposalDescription)}`,
            };

            const { receipt } = await createProposal(proposal, (text) => pushStatus(text));

            lastTxHash.set(receipt.receipt.transactionHash);
            pushStatus("Proposal UserOperation submitted");

            const txReceipt = await waitForTransactionReceipt(get(lastTxHash));
            pushStatus("Proposal transaction included in a block");

            const cost = await getGasCost(get(lastTxHash));
            gasCost.set(cost.totalCostEth);

            const paymasterDeposit = await getPaymasterDeposit();
            paymasterFunds.set(paymasterDeposit.eth);

            await waitForConfirmations(txReceipt, 3, (text) => pushStatus(text));

            pushStatus("Proposal created successfully!");
            showProposalForm.set(false);
            proposalTitle.set("");
            proposalDescription.set("");
            
            // Reload proposals list
            await handleLoadProposals();

            await getBalances();
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error occurred');
        } finally {
            loading.set(false);
        }
    }

    async function handleLoadMembershipTokens() {
        if (!get(eoa)) {
            pushStatus("Connect wallet first");
            return;
        }

        loading.set(true);
        try {
            const tokens = await listMembershipTokens(get(eoa) as `0x${string}`);
            console.log("" + tokens);
            membershipTokens.set(tokens);
            if (tokens.length > 0) {
                selectedTokenId.set(tokens[0].tokenId.toString());
            }
            pushStatus(tokens.length ? "Membership tokens loaded" : "No membership tokens found");
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error loading membership tokens');
        } finally {
            loading.set(false);
        }
    }

    async function handleRenewMembership() {
        const tokenId = get(selectedTokenId);
        if (!tokenId) {
            pushStatus("Select a token id first");
            return;
        }

        loading.set(true);
        try {
            pushStatus("Renewing membership for token " + tokenId + "...");
            const tokenIdBigInt = BigInt(tokenId as string);
            await renewMembership(tokenIdBigInt, (msg) => pushStatus(msg));
            pushStatus("Membership renewed successfully!");
            
            // Reload tokens to refresh state
            await handleLoadMembershipTokens();

            await getBalances();
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error renewing membership');
        } finally {
            loading.set(false);
        }
    }

    async function handleVote(proposalId: bigint, support: number) {
        loading.set(true);
        try {
            pushStatus(`Submitting vote (${support === 1 ? 'For' : support === 0 ? 'Against' : 'Abstain'})...`);
            await castVote(proposalId, support, (msg) => pushStatus(msg));
            pushStatus("✓ Vote submitted successfully!");
            
            // Reload proposals to show updated vote counts
            await handleLoadProposals();

            await getBalances();
        } catch (err: any) {
            console.error(err);
            pushStatus(err?.message ?? 'Error submitting vote');
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
            {#if $smartAccount}
                <section style="margin-top: 1rem;">
                    <p>
                        {#if $eoaBalance}
                            <strong>EOA Balance:</strong> {$eoaBalance} ETH <br/>
                        {/if}

                        <strong>Smart account:</strong> {shortAddress($smartAccount)} <br/>
                        {#if $smartAccountBalance}
                            <strong>Smart Account Balance:</strong> {$smartAccountBalance} ETH <br/>
                        {/if}
                        
                        <label for="paymasterToggle"><strong>Using Paymaster:</strong></label>

                        <label class="switch">
                            <input
                                id="paymasterToggle"
                                type="checkbox"
                                bind:checked={$usePaymaster}
                            />
                            <span class="slider round"></span>
                        </label>

                        <span class="toggle-state">{ $usePaymaster ? "Yes" : "No" }</span>
                    </p>

                    <div class="button-column">
                        <button on:click={checkIsDAOMember} disabled={$loading} class="btn btn-primary">
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
                                on:click={() => handleMintNewMembership(mintTargetAddress as `0x${string}`, BigInt(Date.now()))}
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