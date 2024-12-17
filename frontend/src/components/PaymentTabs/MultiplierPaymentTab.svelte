<script lang="ts">
  import { API } from "../../ts/api";
  import { error, user, config } from "../../ts/mainStore";
  import Tabs from "../Tabs.svelte";
  import FiatTab from "./FiatTab.svelte";
  import CryptoTab from "./CryptoTab.svelte";

  // List of tab items with labels, values and assigned components
  let items = [{ label: "Credit Card", value: 1, component: FiatTab }];

  if ($config.env == "local" || $config.env == "staging") {
    items.push({
      label: "Crypto Currencies",
      value: 2,
      component: CryptoTab,
    });
  }

  let dailyLimit: number;
  let newDailyLimit: any;
  let newDailyLimitForBackend: number;
  let total: number;

  $: if (!$user.multiplierDailyLimit && dailyLimit === undefined) {
    dailyLimit = total = 100;
    newDailyLimit = dailyLimit;
  } else if ($user.multiplierDailyLimit && dailyLimit === undefined) {
    dailyLimit = total = $user.multiplierDailyLimit / 1000000;
    newDailyLimit = dailyLimit;
  }

  function setDailyLimit() {
    try {
      if (newDailyLimit >= 1) {
        newDailyLimitForBackend = parseInt(newDailyLimit) * 1000000;
        API.user.setMultiplierDailyLimit(newDailyLimitForBackend);
        total = dailyLimit = newDailyLimit;
        $user.multiplierDailyLimit = newDailyLimitForBackend;
      } else {
        $error = "The daily limit must be a number greater than or equalt to 1";
        newDailyLimit = dailyLimit;
      }
    } catch (e) {
      $error = e;
    }
  }

  function handleLimitKeyDown(event) {
    if (event.key === "Enter") {
      setDailyLimit();
    }
  }

  function handleLimitChange() {
    setDailyLimit();
  }
</script>

<style>
  .input-wrapper {
    position: relative;
    display: inline-block;
  }

  .currency-symbol {
    position: absolute;
    right: 0px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 1em;
    color: #666;
  }
</style>

<h2 class="p-2 m-2">Multiplier Payment</h2>

<div class="container-col" id="tipping-limit-div">
  <div class="container">
    <label for="daily-limit-input">Daily Limit </label>
    <div class="input-wrapper">
      <input
        id="daily-limit-input"
        type="number"
        class="m-4 max-w20 input-field"
        bind:value={newDailyLimit}
        on:change={handleLimitChange}
        on:keydown={handleLimitKeyDown}
      />
      <span class="currency-symbol">$</span>
    </div>
    <button on:click={setDailyLimit} class="ml-5 p-2 button1"
      >Set Daily Limit</button
    >
  </div>
  <div class="p-2 m-2">
    <Tabs {items} {total} seats={1} freq={1} />
  </div>
</div>
