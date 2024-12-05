<script lang="ts">
  import { appState } from "ts/state.svelte.ts";
  import { API } from "./ts/api.ts";
  // noinspection TypeScriptCheckImport
  //import QR from "svelte-qr";
  import { formatBalance, minBalanceName, qrString } from "./ts/services.svelte.ts";
  import type { PaymentResponse } from "./types/backend";

  export let total: number;
  export let seats: number;
  export let freq: number;

  let selected;
  let paymentResponse: PaymentResponse;

  async function handleSubmit() {
    try {
      paymentResponse = await API.user.nowPayment(selected, freq, seats);
    } catch (e) {
      appState.setError(e);
    }
  }
</script>

<style>
  @media screen and (max-width: 600px) {
    form {
      margin: 0;
    }
    form .container {
      flex-direction: column;
    }
  }
</style>

<form on:submit|preventDefault={handleSubmit}>
  <div class="container">
    <div class="p-2">
      <select bind:value={selected}>
        {#each Object.entries(appState.config.supportedCurrencies) as [key, value]}
          {#if value.isCrypto}
            <option value={key}>
              {value.name}
            </option>
          {/if}
        {/each}
      </select>
    </div>
    <div class="p-2">
      <button class="button1" type="submit" disabled={seats <= 0}
        >❤&nbsp;Support</button
      >
      for ${total.toFixed(2)}
    </div>
  </div>
</form>
{#if paymentResponse}
  <div class="p-2">
    Pay in {formatBalance(
      BigInt(paymentResponse.payAmount),
      paymentResponse.payCurrency
    )}
    {paymentResponse.payCurrency}
    to this address: <b>{paymentResponse.payAddress}</b>
    ({paymentResponse.payAmount}
    {minBalanceName(paymentResponse.payCurrency)})
  </div>
  <!--<QR
    text={qrString(
      paymentResponse.payAddress,
      paymentResponse.payCurrency,
      paymentResponse.payAmount
    )}
    level="H"
  />-->
{/if}
