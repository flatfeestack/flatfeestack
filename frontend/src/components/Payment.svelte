<script lang="ts">
  import { API } from "ts/api";
  import { onMount } from "svelte";
  import { user, paymentCycle } from "ts/auth";
  import { loadStripe } from "@stripe/stripe-js/pure";
  import { User, PaymentCycle, Repo } from "../types/user";


  import Spinner from "./Spinner.svelte";
  import Dots from "./Dots.svelte";

  let stripe;
  let selectedPlan = 0;
  let seats = 1;

  let isSubmitting = false;

  $: {
    if (card) {
      if ($user.payment_method) {
        card.style.display = "none";
      } else {
        card.style.display = "block";
      }
    }
  }

  const plans = [
    {
      title: "Yearly",
      price: 120,
      freq: 365,
      desc: "By paying yearly <b>120&nbsp;USD</b>, you help us to keep payment processing costs low and more money will reach your sponsored projects"
    },
    {
      title: "Quarterly",
      price: 30,
      freq: 90,
      desc: "You want to support Open Source software with a quarterly flat fee of <b>30&nbsp;USD</b>"
    }
  ];

  let card; // HTML div to mount card
  let cardElement;
  let complete = false;
  let paymentProcessing = false;
  let submitted = false;
  let error = "";
  let showSuccess = false;

  function createCardForm() {
    let elements = stripe.elements();
    cardElement = elements.create("card");
    cardElement.mount(card);
    cardElement.on("change", (e) => {
      if (e.complete) {
        complete = e.complete;
        finishSetup();
      }
    });
  }

  const finishSetup = async () => {
    const cs = await API.user.setupStripe();
    if (!cs) {
      error = "could not setup stripe";
      return;
    }
    console.log(cs.data);
    console.log(cs.data.client_secret);
    stripe.confirmCardSetup(
      cs.data.client_secret,
      {
        payment_method: {
          card: cardElement
        }
      }
    ).then(async function(result) {
      if (result.error) {
        console.log(result.error);
      } else {
        $user.payment_method = result.setupIntent.payment_method;
        const res = await API.user.updatePaymentMethod(result.setupIntent.payment_method);
        if (!res.data) {
          console.log("could not verify in email");
          return;
        }
        $user = res.data;
        console.log("OOKKK" + $user.payment_method);
      }
    });
  };

  const deletePaymentMethod = async () => {
    const res = await API.user.deletePaymentMethod()
    if (res.status === 200) {
      $user.payment_method = null;
      createCardForm();
    }
  };

  const handleCancel = async (event) => {
    try {
      const res = await API.user.cancelSub();
      if (res.status === 200) {
        $paymentCycle.seats = 0;
      }
    } catch (e) {
      console.log(e);
    }
  }

  // Handle the submission of card details
  const handleSubmit = async (event) => {
    paymentProcessing = true;
    try {
      const res = await API.user.stripePayment(plans[selectedPlan].freq, seats);

      stripe.confirmCardPayment(res.data.client_secret, {
        payment_method: $user.payment_method
      }).then(function(result) {
        if (result.error) {
          // Show error to your customer
          console.log(result.error.message);
        } else {
          if (result.paymentIntent.status === "succeeded") {
            console.log("yesssss");
          }
        }
      });

      showSuccess = true;
    } catch (e) {
      console.log(e);
      error = "The payment failed. The subscription could not be created.";
    } finally {
      paymentProcessing = false;
    }
  };

  onMount(async () => {

    stripe = await loadStripe("pk_test_51ITqIGItjdVuh2paNpnIUSWtsHJCLwY9fBYtiH2leQh2BvaMWB4de40Ea0ntC14nnmYcUyBD21LKO9ldlaXL6DJJ00Qm1toLdb");
    createCardForm();
    const res = await API.user.getCurrentPayment();
    if (res.status === 200) {
      console.log("setting payment cycle: "+res.data);
      if (res.data) {
        $paymentCycle = res.data;
        selectedPlan = $paymentCycle.freq === 365? 0:1
        seats = $paymentCycle.seats
      } else {
        console.log("empty paymentcily")
      }
    } else {
      console.log(res.data);
      console.log(res.status);
    }
  });

</script>

<style>
    .StripeElement {
        box-sizing: border-box;
        height: 40px;
        padding: 10px 12px;

        border: 1px solid transparent;
        border-radius: 4px;
        background-color: white;

        box-shadow: 1px 1px 7px 2px rgba(0, 0, 0, 0.1);
        -webkit-transition: box-shadow 150ms ease;
        transition: box-shadow 150ms ease;
    }

    .card {
        cursor:pointer;
    }
    .card:hover {
        @apply bg-gradient-to-tr from-secondary-400 to-primary-400 cursor-pointer transform scale-105 text-white;
    }
    .card:active {
        @apply transform from-secondary-600 to-primary-600;
    }
    .price {
        @apply text-3xl font-bold text-center my-5;
    }
    .container {
        display: flex;
        flex-direction: row;
        margin-left: 1em;
        margin-right: 1em;
        align-items: center;
    }
</style>
{#if error}
  <div class="bg-red-500 text-white p-3 my-5">{error}</div>
{/if}

{#if !submitted}
  <div class="container">
    {#each plans as { title, desc }, i}
      <div class="h-100 p-2 m-2 w1-2 card border-primary-500 rounded {selectedPlan === i ? 'bg-green' : ''}"
        on:click="{() => (selectedPlan = i)}">
        <h3 class="text-center font-bold text-lg">{title}</h3>
        <div class="text-center">{@html desc}</div>
      </div>
    {/each}
  </div>
{/if}

<div class="container">
  <div class="p-2 m-2 w-100 StripeElement" bind:this="{card}"></div>
  {#if $user.payment_method}
    <p class="p-2">Credit card: xxxx xxxx xxxx {$user.last4}</p>
    <form class="p-2" on:submit|preventDefault="{deletePaymentMethod}">
      <button disabled="{isSubmitting}" type="submit">Remove card{#if isSubmitting}<Dots />{/if}
      </button>
    </form>
  {/if}
  {#if $paymentCycle != null && $paymentCycle.seats > 0}
    <form class="p-2" on:submit|preventDefault="{handleCancel}">
      <button class="btn my-4" disabled="{isSubmitting}" type="submit">Cancel&nbsp;support
        {#if isSubmitting}
          <Dots />
        {/if}
      </button>
    </form>
  {:else}
    (currently not supporting)
  {/if}

</div>

<div class="container">
  <div class="p-2">
    {#if $user.role == "ORG" }
      How many seats? <input size="5" type="number" min="1" bind:value="{seats}">
    {/if}
  </div>
  <div class="p-2">
    Total Donation: $
    {plans[selectedPlan].price * ($user.role === "ORG" ? seats : 1)}
  </div>
  <div class="flex w-full justify-end">
    <form on:submit|preventDefault="{handleSubmit}">
      <button class="btn my-4" disabled="{isSubmitting}" type="submit">❤ Support
        {#if isSubmitting}
          <Dots />
        {/if}
      </button>
    </form>
  </div>
</div>

{#if paymentProcessing}
  <div class="container">
    <h2 class="p-2">One sec while we're verifying your payment.</h2>
    <Spinner />
  </div>
{/if}
{#if showSuccess}
  <div class="container">
    <h2 class="p-2">Payment Successful!</h2>
  </div>
{/if}
