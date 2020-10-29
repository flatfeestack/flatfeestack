<style type="text/scss">
.StripeElement {
  box-sizing: border-box;

  height: 40px;

  padding: 10px 12px;

  border: 1px solid transparent;
  border-radius: 4px;
  background-color: white;

  box-shadow: 0 1px 3px 0 #e6ebf1;
  -webkit-transition: box-shadow 150ms ease;
  transition: box-shadow 150ms ease;
  input {
    font-size: 1.5rem;
  }
}
</style>

<script>
import { API } from "src/api/api";

let stripe = Stripe("pk_test_8Qs51tLVL0qbzUUgo3YEQPgL");

//export let sku;

// Payment Intents

import { onMount } from "svelte";

let elements = stripe.elements();
let card; // HTML div to mount card
let cardElement;
let complete = false;

let plan = "price_1HhheeFlT4VRPYyK2hZryC8q";

onMount(async () => {
  await createCardForm();
});

// Step 2
async function createCardForm() {
  cardElement = elements.create("card");
  cardElement.mount(card);
  cardElement.on("change", (e) => {
    if (e.complete) {
      console.log("Form complete");
      complete = e.complete;
    }
  });
}

// Handle the submission of card details
const handleSubmit = async (event) => {
  event.preventDefault();

  // Create Payment Method
  const { paymentMethod, error } = await stripe.createPaymentMethod({
    type: "card",
    card: cardElement,
  });

  if (error) {
    alert(error.message);
    return;
  }

  // Create Subscription on the Server
  const subscription = await API.api.payments.createSubscription(
    plan,
    paymentMethod.id
  );

  // The subscription contains an invoice
  // If the invoice's payment succeeded then you're good,
  // otherwise, the payment intent must be confirmed

  const { latest_invoice } = subscription;

  if (latest_invoice.payment_intent) {
    const { client_secret, status } = latest_invoice.payment_intent;

    if (status === "requires_action") {
      const { error: confirmationError } = await stripe.confirmCardPayment(
        client_secret
      );
      if (confirmationError) {
        console.error(confirmationError);
        alert("unable to confirm card");
        return;
      }
    }

    // success
    alert("You are subscribed!");
  }
};
</script>

<div class="StripeElement" bind:this="{card}"></div>
<button
  class="bg-primary-500 p-2 text-white disabled:opacity-75 transition-all duration-150"
  type="submit"
  on:click="{handleSubmit}"
  disabled="{!complete}"
>Create Subscription</button>
