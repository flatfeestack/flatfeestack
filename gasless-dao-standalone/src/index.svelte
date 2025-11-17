<script>
    import { onMount } from "svelte";
    import { getTransactionReceipt, decodeEventLog } from "viem";
    import { createWalletConnection, sendContractTx } from "./lib/wallet.js";
    import { executeSmartAccountTransaction, pollUserOp } from "./lib/aa.js";
    import { setupStaticRoutes } from "preveltekit";
    import { FLATFEESTACK_NFT_ABI, daoAbi } from "./lib/abi"
    import { NFT_ADDRESS, DAO_ADDRESS } from "./lib/config";
    import { publicClient } from "./lib/publicClient.js"

    const contractAddress = NFT_ADDRESS;

    let walletConnection;
    let account = $state(null);
    let isConnected = $state(false);
    let isTestnet = $state(true);
    let usePaymaster = $state(true);
    let hash = $state(null);
    let status = $state("Connect MetaMask");
    let fileName = $state(null);
    let txHash = $state("");
    let userOpHash = $state("");
    let smartAccountAddress = $state("");

    let proposalDescription = $state("");
    let proposalTarget = $state("");
    let proposalCalldata = $state("0x");
    let proposalId = $state("");

    onMount(() => {
        walletConnection = createWalletConnection({
            onConnected: (acc, testnet = true, switchCancelled = false) => {
                account = acc;
                isConnected = true;
                isTestnet = testnet;
                status = switchCancelled
                    ? "Connected but network switch cancelled"
                    : `Connected to ${acc.slice(0, 6)}...${acc.slice(-4)} on ${testnet ? "Sepolia" : "Mainnet"}`;
            },
            onDisconnected: () => {
                account = null;
                isConnected = false;
                hash = null;
                fileName = null;
                status = "Connect MetaMask";
            },
            onChainChanged: (testnet) => {
                isTestnet = testnet;
                if (account) {
                    status = `Connected to ${account.slice(0, 6)}...${account.slice(-4)} on ${testnet ? "Sepolia" : "Mainnet"}`;
                }
            },
            onAccountChanged: (acc) => {
                account = acc;
                if (acc) {
                    status = `Connected to ${acc.slice(0, 6)}...${acc.slice(-4)} on ${isTestnet ? "Sepolia" : "Mainnet"}`;
                }
            },
            onError: (message) => {
                status = message;
            },
        });

        walletConnection.initialize();
    });

    async function checkIsCouncil() {
        try {
            userOpHash = "";
            txHash = "";

            if (smartAccountAddress) {
                userOpHash = await executeSmartAccountTransaction(
                    walletConnection.wallet,
                    smartAccountAddress,
                    account,
                    contractAddress,
                    FLATFEESTACK_NFT_ABI, //abi
                    "isCouncilIndex",
                    [account, 0],
                    usePaymaster
                );
                status = `UserOp submitted: ${userOpHash}`;
                console.log("User Op Hash: " + userOpHash);

                function updateStatus(msg) { status = msg; }

                const receipt = await pollUserOp(userOpHash, { onUpdate: updateStatus });
                console.log("Final result:", receipt);
                txHash = receipt.receipt.transactionHash || "";
            } else {
                txHash = await sendContractTx(walletConnection.wallet, account, contractAddress, FLATFEESTACK_NFT_ABI, "isCouncilIndex", [account, 0]);
                status = `Stored, tx is: ${txHash}`;
            }
        } catch (error) {
            status = `Error: ${error.message}`;
        }
    }

    async function createProposal() {
        status = "Submitting proposal...";

        try {
            //TODO
            /*const encodedCalldata = encodeFunctionData({
                abi: FLATFEESTACK_NFT_ABI,
                functionName: "isCouncil", //TODO
                args: proposalCalldata,
            });*/

            const targets = [proposalTarget];
            const values = [0];
            const calldatas = ["0x"];//TODO

            if (smartAccountAddress) {
                userOpHash = await executeSmartAccountTransaction(
                    walletConnection.wallet,
                    smartAccountAddress,
                    account,
                    DAO_ADDRESS,
                    daoAbi,
                    "propose",
                    [targets, values, calldatas, proposalDescription],
                    usePaymaster
                );

                status = `UserOp submitted: ${userOpHash}`;

                const receipt = await pollUserOp(userOpHash, {
                    onUpdate: (m) => (status = m),
                });

                status = `AA proposal created in tx: ${receipt?.receipt?.transactionHash || "unknown"}`;

                return;
            }

            txHash = await sendContractTx(
                walletConnection.wallet,
                account,
                DAO_ADDRESS,
                daoAbi,
                "propose",
                [targets, values, calldatas, proposalDescription]
            );

            status = `TX sent: ${txHash}<br>Waiting for proposalId...`;

            const receipt = await publicClient.waitForTransactionReceipt({
                hash: txHash,
            });

            const logs = publicClient.decodeEventLog({
                abi: daoAbi,
                eventName: "ProposalCreated",
                data: receipt.logs[0].data,
                topics: receipt.logs[0].topics,
            });

            const proposalId = logs.args.proposalId;

            status = `Proposal created! ID: ${proposalId}<br>TX: ${txHash}`;
        } catch (e) {
            status = `Error: ${e.message}`;
        }
    }

    async function voteOnProposal(support) {
        if (!proposalId) {
            status = "Enter a proposalId";
            return;
        }

        status = "Submitting vote...";

        try {
            if (smartAccountAddress) {
            userOpHash = await executeSmartAccountTransaction(
                walletConnection.wallet,
                smartAccountAddress,
                account,
                DAO_ADDRESS,
                daoAbi,
                "castVote",
                [proposalId, support],
                usePaymaster
            );

            status = `Vote sent through AA: ${userOpHash}`;
            return;
            }

            txHash = await sendContractTx(
            walletConnection.wallet,
            account,
            DAO_ADDRESS,
            daoAbi,
            "castVote",
            [proposalId, support]
            );

            status = `Vote submitted: ${txHash}`;
            
        } catch (e) {
            status = `Error voting: ${e.message}`;
        }
    }

    const routes = {
        staticRoutes: [
            {
                path: "/",
                htmlFilename: "index.html",
            },
        ],
    };

    setupStaticRoutes(routes);
</script>

<div class="container">
    <h1>Paymaster DAO</h1>

    <div class="status-box">
        {#if userOpHash.trim() !== ""}
            <span>
                {@html "User Op Hash: " + (userOpHash.replace(/\n/g, "<br>") + "<br>")}
            </span>
        {/if}
        {#if txHash.trim() !== ""}
            <span>
                {@html "Tx Hash: " + (txHash.replace(/\n/g, "<br>") + "<br>")}
            </span>
        {/if}
        <span>{@html status.replace(/\n/g, "<br>")}</span>
    </div>

    <div class="controls">
        {#if !isConnected}
            <button onclick={() => walletConnection.connect(isTestnet)} class="btn btn-connect">
                Connect MetaMask
            </button>
        {:else}
            <label class="input-group">
                <span>Smart Account (optional):</span>
                <input type="text" bind:value={smartAccountAddress} placeholder="0x..." class="address-input" />
            </label>
            
            <button onclick={() => walletConnection.disconnect()} class="btn btn-secondary">Disconnect</button>
        {/if}

        <button onclick={checkIsCouncil} class="btn btn-primary">Check is Council</button>

        <label class="network-toggle">
            <input type="checkbox" bind:checked={usePaymaster} />
            <span>Use Paymaster (Gasless)</span>
        </label>

        <hr>

        <h2>DAO Proposal</h2>

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
        </div>
    </div>
</div>

<style>
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

    .input-group {
        display: flex;
        flex-direction: column;
        gap: 6px;
        font-size: 14px;
    }

    .address-input {
        padding: 10px;
        border-radius: 8px;
        border: 1px solid #ccc;
        font-size: 14px;
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

    .btn-connect {
        background: #0077ff;
        color: white;
    }

    .btn-connect:hover {
        background: #005fcc;
    }

    .btn-secondary {
        background: #e9e9e9;
    }

    .btn-secondary:hover {
        background: #d5d5d5;
    }

    .btn-primary {
        background: #4caf50;
        color: white;
    }

    .btn-primary:hover {
        background: #419445;
    }

    .network-toggle {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-top: 4px;
        font-size: 14px;
    }

    .status-box span {
        display: block;
        margin-bottom: 4px;
    }

    input[type="checkbox"] {
        width: 18px;
        height: 18px;
    }

</style>