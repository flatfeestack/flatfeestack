<script lang="ts">
  import detectEthereumProvider from "@metamask/detect-provider";
  import { ethers, providers } from "ethers";
  import { onMount } from "svelte";
  import { MembershipABI } from "../../contracts/Membership";
  import { error } from "../../ts/store";

  export let isRepresentative: boolean = false;

  // https://ethereum.stackexchange.com/a/42810
  onMount(async () => {
    const ethProv = await detectEthereumProvider();
    
    if (ethProv) {
      try {
        const provider = new providers.Web3Provider(<any>ethProv);
        await provider.send("eth_requestAccounts", []);
        const signer = provider.getSigner();

        const membershipContract = new ethers.Contract(
          import.meta.env.VITE_MEMBERSHIP_CONTRACT_ADDRESS,
          MembershipABI,
          signer
        );
        const representative = await membershipContract.representative();
        isRepresentative = representative === (await signer.getAddress());
      } catch (e) {
        $error = e;
      }
    } else {
      $error =
        'Please install <a href="https://metamask.io/download.html">MetaMask</a>';
    }
  });
</script>

<div class="container">
  {#if isRepresentative}
    <p>You are the representative!</p>
  {:else}
    <p>You are not the representative!</p>
  {/if}
</div>
