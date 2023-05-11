<script lang="ts">
  import FiatTab from "./PaymentTabs/FiatTab.svelte";
  import CryptoTab from "./PaymentTabs/CryptoTab.svelte";
  import Tabs from "./Tabs.svelte";
  import { config, error } from "../ts/mainStore";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import type { Plan } from "../types/backend";

  // List of tab items with labels, values and assigned components
  let items = [
    { label: "Credit Card", value: 1, component: FiatTab },
    { label: "Crypto Currencies", value: 2, component: CryptoTab },
  ];

  let currentFreq: number = 365;
  let currentSeats = 1;
  let selectedPlan: Plan;

  $: {
    if ($config && $config.plans) {
      selectedPlan = $config.plans.find((e) => e.freq === currentFreq);
      if (!selectedPlan) {
        selectedPlan = $config.plans[0];
      }
    }
  }

  let total: number = 0;
  $: {
    if (selectedPlan) {
      total = selectedPlan.price * currentSeats;
    }
  }

  onMount(async () => {
    try {
      const res = await API.user.get();
      if (res.freq && res.seats) {
        currentFreq = res.freq;
        currentSeats = res.seats;
      }
    } catch (e) {
      $error = e;
    }
  });
</script>

<style>
  .small {
    font-size: x-small;
  }
  .child {
    margin: 0.5em;
    box-shadow: 0.25em 0.25em 0.25em #e1e1e3;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }
</style>

<h2 class="p-2 m-2">Payment</h2>
<p class="p-2 m-2">
  We request your permission that we initiate a payment or a series of payments
  on your behalf. By continuing, I authorize FlatFeeStack to send instructions
  to the financial institution that issued my card to take payments from my card
  account in accordance with the terms of my agreement with you.
</p>
<div class="container-stretch">
  {#if $config.plans}
    {#each $config.plans as { title, desc, disclaimer, freq }}
      <div class="flex-grow child p-2 m-2 w1-2 card border-primary-500 rounded {currentFreq === freq ? 'bg-green': ''}">
        <button class="accessible-btn" on:click={() => (currentFreq = freq)}>
          <h3 class="text-center font-bold text-lg">{title}</h3>
          <div class="text-center">{@html desc}</div>
          <div class="small text-center">{@html disclaimer}</div>
        </button>
      </div>
    {/each}
  {/if}
</div>

<div class="container page">
  <div class="p-2">
    <input size="5" type="number" min="1" bind:value={currentSeats} /> Seats
  </div>
  <div class="p-2">
    {#if $config.plans}
      <div>
        Sponsoring Amount:<span class="bold">$ {total.toFixed(2)}</span>
      </div>
      <div class="small">
        ([{currentSeats} x {selectedPlan.price}])
      </div>
    {/if}
  </div>
</div>

<div class="p-2 m-2">
  <Tabs {items} {total} seats={currentSeats} freq={currentFreq} />
</div>
