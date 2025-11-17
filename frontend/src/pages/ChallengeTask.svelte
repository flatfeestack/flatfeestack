<script>
    import { onMount } from "svelte";
    import { createWalletConnection, callContractFunction, sendContractTx } from "../lib/wallet.js";
    import { executeSmartAccountTransaction, pollUserOp } from "../lib/aa.js";
    import { setupStaticRoutes } from "preveltekit";
    import { FLATFEESTACK_NFT_ABI } from "../lib/abis"

    const abi = [
        {
            inputs: [{ internalType: "bytes32", name: "hash", type: "bytes32" }],
            name: "store",
            outputs: [],
            stateMutability: "nonpayable",
            type: "function",
        },
        {
            inputs: [
                { internalType: "address", name: "recipient", type: "address" },
                { internalType: "bytes32", name: "hash", type: "bytes32" },
            ],
            name: "verify",
            outputs: [{ internalType: "uint256", name: "", type: "uint256" }],
            stateMutability: "view",
            type: "function",
        },
    ];

    const contractAddress = "0x08b0049895ce4c87749b7439cb2ad553cec7caf9";

    let walletConnection;
    let account = $state(null);
    let isConnected = $state(false);
    let isTestnet = $state(true);
    let usePaymaster = $state(true);
    let hash = $state(null);
    let status = $state("Choose network and connect MetaMask");
    let fileName = $state(null);
    let isVerified = $state(false);
    let txHash = $state("");
    let userOpHash = $state("");
    let smartAccountAddress = $state("");

    async function hashFile(file) {
        const buffer = await file.arrayBuffer();
        const hashBuffer = await crypto.subtle.digest("SHA-256", buffer);
        return (
            "0x" +
            Array.from(new Uint8Array(hashBuffer))
                .map((b) => b.toString(16).padStart(2, "0"))
                .join("")
        );
    }

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
                isVerified = false;
                status = "Choose network and connect MetaMask";
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

    async function filesChange(fileList) {
        if (!isConnected) {
            status = "Please connect MetaMask first";
            return;
        }
    
        hash = await hashFile(fileList[0]);
        fileName = fileList[0].name;
        isVerified = false; // Reset verification state
    
        try {
            const addr = smartAccountAddress || account;
            const timestamp = await callContractFunction(walletConnection.wallet, contractAddress, abi, "verify", [addr, hash]);
    
            if (timestamp.toString() === "0") {
                status = `Not yet stored from account: ${addr}`;
            } else {
                isVerified = true;
                status = `<b>VERIFIED</b> in the blockchain! Timestamp: ${timestamp.toString()}`;
            }
        } catch (error) {
            hash = null;
            fileName = null;
            status = `Error: ${error.message}`;
        }
    }

    const DAO_ABI = [
        {
            name: "setNewBylawsHash",
            type: "function",
            stateMutability: "nonpayable",
            inputs: [{ name: "newHash", type: "uint256" }],
            outputs: [],
        },
        ];

    async function store() {
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
                txHash = await sendContractTx(walletConnection.wallet, account, contractAddress, abi, "store", [hash]);
                status = `Stored, tx is: ${txHash}`;
            }
            isVerified = true;
        } catch (error) {
            status = `Error: ${error.message}`;
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
    <h1>Notarize PDF</h1>

    <div class="dropbox" class:disabled={!isConnected}>
        <input type="file" onchange={(e) => filesChange(e.target.files)} class="input-file" disabled={!isConnected} />
        {#if isConnected}
            Drag your file here or click to browse
        {:else}
            Connect MetaMask to upload files
        {/if}
        {#if hash !== null}
            <div class="file-info">
                <div>
                    <strong>File:</strong>
                    {fileName}
                </div>
                <div>
                    <strong>SHA256:</strong>
                    {hash}
                </div>
            </div>
        {/if}
    </div>

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
            <button onclick={() => walletConnection.disconnect()} class="btn btn-secondary">Disconnect</button>
        {/if}

        <button onclick={store} class="btn btn-primary">Store</button>

        <label class="network-toggle">
            <input type="checkbox" bind:checked={usePaymaster} />
            <span>Gasless</span>
        </label>
    </div>
</div>
