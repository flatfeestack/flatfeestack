<script lang="ts">
  import { onMount } from "svelte";
  import { config, error, user, userBalances } from "./../ts/store";
  import Dots from "./Dots.svelte";
  import { stripePayment, stripePaymentMethod } from "../ts/services";
  import { loadStripe } from "@stripe/stripe-js/pure";
  import { API } from "../ts/api";
  import { plans } from "../types/contract";

  let stripe;
  let selectedPlan = 0;
  let seats = 1;
  let isSubmitting = false;
  let card; // HTML div to mount card
  let cardElement;
  let paymentProcessing = false;
  let showSuccess = false;

  $: {
    if (card) {
      if ($user.payment_method || paymentProcessing) {
        card.style.display = "none";
      } else {
        showSuccess = false;
        card.style.display = "block";
      }
    }
  }

  function createCardForm() {
    if(!cardElement) {
      console.log("called create");
      let elements = stripe.elements();
      cardElement = elements.create("card");
      cardElement.mount(card);
      cardElement.on("change", (e) => {
        if (e.error) {
          $error = e.error;
        }
      });
   }
  }

  const handleSubmit = async () => {
    paymentProcessing = true;
    isSubmitting = true;
    try {
      if ($user.payment_method) {
        await stripePayment(stripe, plans[selectedPlan].freq, seats, $user.payment_method);
      } else {
        await stripePaymentMethod(stripe, cardElement);
        await stripePayment(stripe, plans[selectedPlan].freq, seats, $user.payment_method);
      }
      showSuccess = true;
    } catch (e) {
      $error = e;
    } finally {
      paymentProcessing = false;
      isSubmitting = false;
    }
  };

  onMount(async () => {
    stripe = await loadStripe($config.stripePublicApi);
    const pc = await API.user.paymentCycle();
    if (pc) {
      seats = pc.seats == 0 ? 1 : pc.seats;
      selectedPlan = pc.freq == 365 ? 0 : 1;
    }
    createCardForm();
  });

</script>

<style>
    .small {
        font-size: x-small;
    }

    .child {
        margin: 0.5em;
        box-shadow: 0.25em 0.25em 0.25em #e1e1e3;
    }

    .w25 {
        width: 25rem;
    }

    @media (max-width: 36rem) {
        .page {
            flex-direction: column;
            display: flex;
        }

        .w25 {
            width: 20rem;
        }
    }

</style>
<h2 class="px-2">Credit Card</h2>

<div class="container-stretch">
  {#each plans as { title, desc }, i}
    <div
      class="child p-2 m-2 w1-2 card cursor-pointer border-primary-500 rounded {selectedPlan === i ? 'bg-green' : ''}"
      on:click="{() => (selectedPlan = i)}">
      <h3 class="text-center font-bold text-lg">{title}</h3>
      <div class="text-center">{@html desc}</div>
    </div>
  {/each}
</div>

<div class="container page">
  {#if $user.role === "ORG" }
    <div class="p-2">
      <span>How many seats?</span>
      <input size="5" type="number" min="1" bind:value={seats}>
    </div>
  {/if}

  <div class="p-2 m-2 w25 rounded border-primary-700" bind:this="{card}"></div>

  <div class="p-2">
    <form on:submit|preventDefault="{handleSubmit}">
      <button disabled="{isSubmitting}" type="submit">❤&nbsp;Support
        {#if isSubmitting}
          <Dots />
        {/if}
      </button>
    </form>
  </div>

  {#if showSuccess}
    <div class="p-2">Payment successful!</div>
  {/if}
  {#if paymentProcessing}
    <div class="p-2">Verifying payment
      <Dots />
    </div>
  {/if}

</div>

<div class="container-col">
  <div class="p-2">
    Total&nbsp;Donation:<span
    class="bold">$&nbsp;{plans[selectedPlan].price * ($user.role === "ORG" ? seats : 1)}</span>
  </div>

  <div class="border-primary-500 rounded small p-2 m-2">
    We request your permission that we initiate a payment or a series of {plans[selectedPlan].title.toLowerCase()}
    payments on your behalf of
    {plans[selectedPlan].price * ($user.role === "ORG" ? seats : 1)} USD.<br />
    By continuing, I authorize flatfeestack.io to send instructions to the financial institution that issued my card to
    take payments from my card account in accordance with the terms of my agreement with you.
  </div>
</div>

