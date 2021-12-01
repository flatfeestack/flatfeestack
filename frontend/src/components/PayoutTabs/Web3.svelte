<script lang="ts">
  import { ethers, providers } from "ethers";
  import { ABI } from "../../types/contract";
  import { error, user, config } from "../../ts/store";
  import detectEthereumProvider from "@metamask/detect-provider";
  import { onMount } from "svelte";
  import Spinner from "../Spinner.svelte";
  import Dots from "../Dots.svelte";

  let storageContract;
  let viewContract;
  let balance = 0;

  onMount(async () => {
    const ethProv = await detectEthereumProvider();
    if (ethProv) {
      try {
        const provider = new providers.Web3Provider(<any>ethProv);
        const signer = provider.getSigner();
        const contractAddress = $config.contractAddr;
        storageContract = new ethers.Contract(contractAddress, ABI, signer);
        viewContract = new ethers.Contract(contractAddress, ABI, provider);
      } catch (e) {
        $error = e;
      }
    } else {
      $error = "Please install <a href=\"https://metamask.io/download.html\">MetaMask</a>";
    }
  });

  const requestFunds = async () => {
    try {
      ethereum.request({ method: 'eth_requestAccounts' });
      await storageContract.release();
    } catch (e) {
      $error = e;
    }
  };
</script>

<style>
    main {
        text-align: center;
        padding: 1em;
        max-width: 240px;
        margin: 0 auto;
    }
    h1 {
        color: #ff3e00;
        text-transform: uppercase;
        font-size: 4em;
        font-weight: 100;
    }
    @media (min-width: 640px) {
        main {
            max-width: none;
        }
    }
</style>

<div class="container">
  <label class="px-2">Request funds:</label>
  {#await balance}
    <Dots /> ETH
  {:then res}
    {res} ETH
  {:catch err}
    {$error = err}
  {/await}
  <button class="button2" on:click="{requestFunds}">Request funds</button>
</div>
