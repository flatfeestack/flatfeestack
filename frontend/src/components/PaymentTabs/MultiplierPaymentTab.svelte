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

  let dailyLimit: number = 100;
  let newDailyLimit: any;
  let newDailyLimitForBackend: number;
  let total: number = 100;

  $: {
    if ($user.multiplierDailyLimit) {
      total = dailyLimit = $user.multiplierDailyLimit / 1000000;
    }
  }

  function setDailyLimit() {
    try {
      if (newDailyLimit >= 1) {
        newDailyLimitForBackend = parseInt(newDailyLimit) * 1000000;
        API.user.setMultiplierDailyLimit(newDailyLimitForBackend);
        total = dailyLimit = newDailyLimit;
        $user.multiplierDailyLimit = newDailyLimitForBackend;
        newDailyLimit = "";
      } else {
        $error = "The daily limit must be a number greater than or equalt to 1";
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
</script>

<style>
</style>

<h2 class="p-2 m-2">Multiplier Payment</h2>

<div class="container-col" id="tipping-limit-div">
  <p>
    Your Multiplier Sponsoring limit is set to <strong
      >${new Intl.NumberFormat("de-CH", { useGrouping: true }).format(
        dailyLimit
      )}</strong
    > per day.
  </p>
  <div class="container">
    <label for="daily-limit-input">Daily Limit </label>
    <input
      id="daily-limit-input"
      type="number"
      class="m-4 max-w20"
      bind:value={newDailyLimit}
      on:keydown={handleLimitKeyDown}
      placeholder="$"
    />
    <button on:click={setDailyLimit} class="ml-5 p-2 button1"
      >Set Daily Limit</button
    >
  </div>
  <div class="p-2 m-2">
    <Tabs {items} {total} seats={1} freq={1} />
  </div>
</div>
