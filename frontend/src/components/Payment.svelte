<style type="text/scss">
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
  @apply shadow-md p-2 mx-3 rounded-lg transition-all duration-150 bg-gray-100 p-5;
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
</style>

<script>
import { API } from "src/api/api";
import format from "date-fns/format";
import { LottiePlayer } from "@lottiefiles/svelte-lottie-player";

let stripe = Stripe("pk_test_8Qs51tLVL0qbzUUgo3YEQPgL");

import { onMount } from "svelte";
import Spinner from "./UI/Spinner.svelte";
import { updateUser } from "src/store/authService";
import { user } from "src/store/auth";

let selectedPlan = 1;

const plans = [
  {
    title: "Monthly",
    price: "$ 10.-",
    desc:
      "You wan't to support Open Source software with a monthly flat fee of.",
    stripe_id: "price_1HhheeFlT4VRPYyK2hZryC8q",
  },
  {
    title: "Yearly",
    price: "$ 120.-",
    desc:
      "By paying yearly, you help us to keep payment processing costs low and more money will reach your sponsored projects",
    stripe_id: "price_1HhhefFlT4VRPYyKqaH4eQuC",
  },
  {
    title: "Quarterly",
    price: "$ 30.-",
    desc:
      " If you're not cool with paying yearly but still want us to keep payment processing costs low :)",
    stripe_id: "price_1HhhefFlT4VRPYyKuS7gWwPw",
  },
];

let elements = stripe.elements();
let card; // HTML div to mount card
let cardElement;
let complete = false;
let paymentProcessing = false;
let submitted = false;
let error = "";
let showSuccess = false;

let interval;

async function createCardForm() {
  cardElement = elements.create("card");
  cardElement.mount(card);
  cardElement.on("change", (e) => {
    if (e.complete) {
      complete = e.complete;
    }
  });
}

// Handle the submission of card details
const handleSubmit = async (event) => {
  try {
    event.preventDefault();
    paymentProcessing = true;
    submitted = true;

    // Create Payment Method
    const { paymentMethod, err } = await stripe.createPaymentMethod({
      type: "card",
      card: cardElement,
    });

    if (err) {
      error = error.message;
      paymentProcessing = false;
      return;
    }

    // Create Subscription on the Server
    const res = await API.api.payments.createSubscription(
      plans[selectedPlan].stripe_id,
      paymentMethod.id
    );
    const subscription = res.data.data;

    // The subscription contains an invoice
    // If the invoice's payment succeeded then you're good,
    // otherwise, the payment intent must be confirmed

    const { latest_invoice } = subscription;

    if (latest_invoice && latest_invoice.payment_intent) {
      const { client_secret, status } = latest_invoice.payment_intent;

      if (status === "requires_action") {
        const { error: confirmationError } = await stripe.confirmCardPayment(
          client_secret
        );
        if (confirmationError) {
          console.error(confirmationError);
          error =
            "Unable to confirm card. The subscription could not be created.";
          return;
        }
      }
    }
    error = "";
  } catch (e) {
    error = "The payment failed. The subscription could not be created.";
  } finally {
    paymentProcessing = false;
  }
};

$: if (submitted && !error && !paymentProcessing) {
  console.log("starting to fetch");
  interval = setInterval(() => updateUser(), 1000);
}

$: if ($user.subscription_state === "active" && interval) {
  clearInterval(interval);
}

$: if ($user.subscription_state === "active") {
  showSuccess = true;
}

onMount(async () => {
  if ($user.subscription_state !== "active") {
    await createCardForm();
  }
});
</script>

{#if error}
  <div class="bg-red-500 text-white p-3 my-5">{error}</div>
{/if}
{#if !submitted && $user.subscription_state !== 'active'}
  <div class="flex items-end mb-10">
    {#each plans as { title, price, desc }, i}
      <div
        class="w-1/3 card {selectedPlan === i ? 'text-white bg-gradient-to-tr from-secondary-600 to-primary-600' : 'text-black'}"
        on:click="{() => (selectedPlan = i)}"
      >
        <h3 class="text-center font-bold text-lg">{title}</h3>
        <div class="price">{price}</div>
        <div class="opacity-50 text-sm text-center">{desc}</div>
      </div>
    {/each}
  </div>
{/if}
{#if $user.subscription_state !== 'active'}
  <div class="w-2/3 mx-auto {submitted ? 'hidden' : ''}">
    <div class="font-semibold mb-5">
      Selected Plan:
      {plans[selectedPlan].title}
    </div>
    <div class="font-semibold mb-5">
      Next Payment:
      {plans[selectedPlan].price}
      at
      {format(new Date(), 'do LLLL yyyy')}
    </div>
    <div class="StripeElement" bind:this="{card}"></div>
    <div class="flex w-full justify-end">
      <button
        class="bg-primary-500 p-2 text-white disabled:opacity-75 transition-all duration-150 mt-2"
        type="submit"
        on:click="{handleSubmit}"
        disabled="{!complete}"
      >Create Subscription</button>
    </div>
  </div>
{/if}

{#if paymentProcessing || (submitted && !paymentProcessing && !error && $user.subscription_state !== 'active')}
  <div class="w-full flex flex-col items-center">
    <h2>One sec while we're verifying your payment.</h2>
    <Spinner />
  </div>
{/if}
{#if showSuccess && submitted}
  <div class="w-full flex flex-col items-center">
    <h2>Success! Welcome onboard!</h2>
    <LottiePlayer
      src="/assets/animations/payment-success.json"
      autoplay="{true}"
      loop="{false}"
      controls="{false}"
      renderer="svg"
      background="transparent"
      height="{300}"
      width="{300}"
    />
  </div>
{/if}
