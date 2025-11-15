<script lang="ts">
    import { ethers } from "ethers";
    import { nftAbi, daoAbi, paymasterAbi, entryPointAbi } from "../lib/abis";

    let daoAddress = "P0x5c35320231aa8677b65c8231ef2637111e7354fb";
    let entryPointAddress = "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789";
    let paymasterAddress = "0x62e3a30b7b3a737c114aa99ece0d49aad58528c4";
    let nftAddress = "0x08b0049895ce4c87749b7439cb2ad553cec7caf9";

    let provider: ethers.BrowserProvider;
    let signer: ethers.JsonRpcSigner;

    let dao: any;
    let nft: any;
    let paymaster: any;
    let entryPoint: any;

    let walletAddress = "";

    async function connect() {
        console.log("connect pressed");

        if (!window.ethereum) {
            alert("Install MetaMask!");
            return;
        }

        provider = new ethers.BrowserProvider(window.ethereum);
        signer = await provider.getSigner();
        walletAddress = await signer.getAddress();

        // Load contracts
        dao = new ethers.Contract(daoAddress, daoAbi, signer);
        nft = new ethers.Contract(nftAddress, nftAbi, signer);
        paymaster = new ethers.Contract(paymasterAddress, paymasterAbi, signer);
        entryPoint = new ethers.Contract(entryPointAddress, entryPointAbi, signer);

        console.log("Connected:", walletAddress);
    }

    async function createProposal() {
        console.log("create proposal pressed");

        if (!dao) {
            alert("Wallet not connected");
            return;
        }

        // Example proposal: no-op function call on NFT
        const targets = [nftAddress];
        const values = [0];
        const calldata = [
            nft.interface.encodeFunctionData("balanceOf", [walletAddress])
        ];

        const description = "Test Proposal: Read membership.";

        try {
            const tx = await dao.propose(targets, values, calldata, description);
            const receipt = await tx.wait();

            console.log("Proposal created:", receipt);
            alert("Proposal created!");
        } catch (err) {
            console.error("Proposal error:", err);
            alert("Error creating proposal");
        }
    }

    let isMember = false;

    async function checkMembership() {
        const balance = await nft.balanceOf(walletAddress);
        isMember = balance > 0;
        console.log("is Member " + isMember);
    }
</script>

<style>
  .full-page-container {
    margin: 1rem 20vw 3rem;
  }
</style>

<div class="container-col2 full-page-container">
    <h1>FlatFeeStack DAO</h1>

    <button on:click={connect}>Connect Wallet</button>

    {#if walletAddress}
        <p>Connected: {walletAddress}</p>
    {/if}

    <hr />

    <button on:click={createProposal}>Create Proposal</button>
    <button on:click={checkMembership}>Is Member</button>
</div>
