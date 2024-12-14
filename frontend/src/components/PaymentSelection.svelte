<script lang="ts">
  import PaymentTab from "./PaymentTabs/PaymentTab.svelte";
  import MultiplierPaymentTab from "./PaymentTabs/MultiplierPaymentTab.svelte";
  import Tabs from "./Tabs.svelte";
  import { user } from "../ts/mainStore";

  let items = [
    { label: "Payment", value: 1, component: PaymentTab },
    { label: "Multiplier Payment", value: 0, component: MultiplierPaymentTab },
  ];

  let multiplierActive: undefined | boolean;

  $: {
    if (typeof multiplierActive === "undefined" && $user.multiplier) {
      multiplierActive = $user.multiplier;
    }
  }
</script>

<style>
</style>

{#if multiplierActive}
  <div class="p-2 m-2">
    <Tabs {items} />
  </div>
{:else}
  <PaymentTab />
{/if}
