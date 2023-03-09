<script lang="ts">
  import { config, error } from "../../ts/mainStore";
  import { API } from "../../ts/api";
  // noinspection TypeScriptCheckImport
  import QR from "svelte-qr";
  import { formatBalance, minBalanceName, qrString } from "../../ts/services";
  import type { PaymentResponse } from "../../types/users";

  export let remaining: number;
  export let current: number;
  export let seats: number;
  export let freq: number;

  let selected;
  let paymentResponse: PaymentResponse;

  async function handleSubmit() {
    try {
      paymentResponse = await API.user.nowPayment(selected, freq, seats);
    } catch (e) {
      $error = e;
    }
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <div class="container">
    <div class="p-2">
      <select bind:value={selected}>
        {#each Object.entries($config.supportedCurrencies) as [key, value]}
          {#if value.isCrypto}
            <option value={key}>
              {value.name}
            </option>
          {/if}
        {/each}
      </select>
    </div>
    <div class="p-2">
      <button class="button1" type="submit" disabled={remaining < current / 2}
        >❤&nbsp;Support</button
      >
      {#if remaining < current / 2}
        (No need to top-up your account, you still funds)
      {:else}
        for ${remaining.toFixed(2)}
      {/if}
    </div>
  </div>
</form>
{#if paymentResponse}
  <div class="p-2">
    Pay in {formatBalance(
      paymentResponse.payAmount,
      paymentResponse.payCurrency
    )}
    {paymentResponse.payCurrency}
    to this address: <b>{paymentResponse.payAddress}</b>
    ({paymentResponse.payAmount}
    {minBalanceName(paymentResponse.payCurrency)})
  </div>
  <QR
    text={qrString(
      paymentResponse.payAddress,
      paymentResponse.payCurrency,
      paymentResponse.payAmount
    )}
    level="H"
  />
{/if}
