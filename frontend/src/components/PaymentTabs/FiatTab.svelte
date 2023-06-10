<script lang="ts">
  import { error, user, config } from "../../ts/mainStore";
  import Dots from "../Dots.svelte";
  import { stripePayment, stripePaymentMethod } from "../../ts/services";
  import { loadStripe } from "@stripe/stripe-js/pure";
  import {
    CardCvc,
    CardExpiry,
    CardNumber,
    Elements,
    PaymentElement,
  } from "svelte-stripe";
  import { API } from "../../ts/api";
  import type { Stripe } from "@stripe/stripe-js";

  export let total: number;
  export let seats: number;
  export let freq: number;

  let stripe: Stripe | null = null;

  let isSubmitting = false;
  let cardElement;
  let paymentProcessing = false;
  let showSuccess = false;

  $: {
    if ($config.stripePublicApi) {
      load();
    }
  }

  async function load() {
    stripe = await loadStripe($config.stripePublicApi);
  }

  const handleSubmit = async () => {
    paymentProcessing = true;
    isSubmitting = true;
    try {
      if (!$user.paymentMethod) {
        await stripePaymentMethod(stripe, cardElement);
      }
      await stripePayment(stripe, freq, seats, $user.paymentMethod);
      showSuccess = true;
    } catch (e) {
      $error = e;
    } finally {
      paymentProcessing = false;
      isSubmitting = false;
    }
  };

  async function deletePaymentMethod() {
    isSubmitting = true;
    try {
      const p1 = API.user.deletePaymentMethod();
      const p2 = API.user.cancelSub();
      $user.paymentMethod = null;
      $user.last4 = null;
      await p1;
      await p2;
    } catch (e) {
      $error = e;
    } finally {
      isSubmitting = false;
    }
  }
</script>

<style>
  :global(.w20) {
    width: 20rem;
  }
  :global(.w4) {
    width: 4rem;
  }

  .stripe p {
    padding: 0.5rem;
    margin: 0;
  }
  @media screen and (max-width: 600px) {
    form .container {
      display: flex;
      flex-wrap: wrap;
    }
    .stripe,
    .stripe .container {
      margin: 0;
    }
    .stripe .container button {
      margin: 0.5rem;
    }
    :global(.w20) {
      width: 100%;
    }
  }
</style>

{#if $user.paymentMethod && $user.last4 !== null}
  <div class="container">
    <p class="nobreak">Credit card:</p>
    <div class="container">
      <span>*** *** *** {$user.last4}</span>
      <form class="p-2" on:submit|preventDefault={deletePaymentMethod}>
        <button class="button3" disabled={isSubmitting} type="submit"
          >Cancel
          {#if isSubmitting}
            <Dots />
          {/if}
        </button>
      </form>
    </div>
  </div>
{/if}

<div class="stripe container">
  <div class="p-2">
    {#if stripe}
      <form on:submit|preventDefault={handleSubmit}>
        <div class="container">
          <Elements {stripe}>
            <CardNumber
              classes={{ base: "w20 p-2 m-2 rounded border-primary-700" }}
              bind:element={cardElement}
            />
            <CardExpiry
              classes={{ base: "w4 p-2 m-2 rounded border-primary-700" }}
            />
            <CardCvc
              classes={{ base: "w4 p-2 m-2 rounded border-primary-700" }}
            />
          </Elements>
          <button class="button1" type="submit" disabled={seats <= 0}
            >❤&nbsp;Support{#if isSubmitting}<Dots />{/if}</button
          >
          <p>for ${total.toFixed(2)}</p>
        </div>
      </form>
    {/if}
  </div>
</div>

{#if showSuccess}<div class="p-2">Payment successful sent</div>{/if}
{#if paymentProcessing}<div class="p-2">Verifying payment<Dots /></div>{/if}
