<script lang="ts">
  import {ethers, providers} from "ethers";
  import {ABI} from "../../types/contract";
  import {error, user, config} from "../../ts/store";
  import detectEthereumProvider from "@metamask/detect-provider";
  import {onMount} from "svelte";
  import Spinner from "../Spinner.svelte";
  import Dots from "../Dots.svelte";
  import { TempleWallet } from "@temple-wallet/dapp";


  let balance

  $: {
    if ($user.payout_eth) {

    }
  }

  onMount(async () => {
    const available = await TempleWallet.isAvailable();
    if (!available) {
      throw new Error("Temple Wallet not installed");
    }

    // Note:

    // use `TempleWallet.isAvailable` method only after web application fully loaded.

    // Alternatively, you can use the method `TempleWallet.onAvailabilityChange`
    // that tracks availability in real-time .

    const wallet = new TempleWallet("My Super DApp");
    await wallet.connect("granadanet");
    const tezos = wallet.toTezos();

    const accountPkh = await tezos.wallet.pkh();
    balance = await tezos.tz.getBalance(accountPkh);
  });

  const requestFunds = async () => {
    try {

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
    <Dots/>
    TZ
  {:then res}
    {res} TZ
  {:catch err}
    {$error = err}
  {/await}
  <button class="button2" on:click="{requestFunds}">Request funds</button>
</div>
