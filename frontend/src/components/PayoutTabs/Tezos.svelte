<script lang="ts">
  import {error, user} from "../../ts/store";
  import {onMount} from "svelte";
  import Dots from "../Dots.svelte";
  import { TempleWallet } from "@temple-wallet/dapp";


  let balance
  let contract

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

    const permission = await TempleWallet.getCurrentPermission();
    const wallet = new TempleWallet("Flatfeestack", permission);
    await wallet.connect("granadanet");
    const tezos = wallet.toTezos();
    // example contract with amount which can be changed
    contract = await tezos.wallet.at('KT1K22GJXnz7ufXJbqyjQ859HHN3AAaU9act');
    balance = await contract.storage()
  });

  const requestFunds = async () => {
    try {
      const op = await contract.methods.replace(balance * 2).send();
      await op.confirmation();
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
