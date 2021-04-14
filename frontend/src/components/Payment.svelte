<script lang="ts">
  import { API } from "ts/api";
  import { onMount } from "svelte";
  import { token, user } from "ts/auth";
  import { loadStripe } from "@stripe/stripe-js/pure";

  import Spinner from "./Spinner.svelte";
  import Dots from "./Dots.svelte";
  import { get } from "svelte/store";
  import { UserBalance } from "../types/user";

  let stripe;
  let selectedPlan = 0;
  let seats = 1;

  let isSubmitting = false;
  let payments: UserBalance[];

  const plans = [
    {
      title: "Yearly",
      price: 120,
      desc: "By paying yearly <b>120&nbsp;USD</b>, you help us to keep payment processing costs low and more money will reach your sponsored projects"
    },
    {
      title: "Quarterly",
      price: 30,
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

  $: {
    if (card) {
      if ($user.payment_method) {
        card.style.display = "none";
      } else {
        card.style.display = "block";
      }
    }
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
        console.log(cardElement);
        console.log("test");
        console.log(result.setupIntent);
        console.log(result.setupIntent.payment_method.card);
        const res = await API.user.updatePaymentMethod(result.setupIntent.payment_method);
        if (!res.data) {
          console.log("could not verify in email");
          return;
        }
        $user = res.data;
        console.log("OOKKK");
      }
    });
  };

  const deletePaymentMethod = async () => {
    console.log(card);
    $user.payment_method = null;
    createCardForm();
  };

  const handleCancel = async (event) => {
    try {
      const res = await API.user.cancelSub();
      if (res.status === 200) {
        $user.freq=0;
      }
    } catch (e) {
      console.log(e);
    }
  }

  // Handle the submission of card details
  const handleSubmit = async (event) => {
    paymentProcessing = true;
    try {
      console.log("HERE");
      const res = await API.user.stripePayment("yearly", 1);

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

  /*if (submitted && !error && !paymentProcessing) {
  console.log("starting to fetch");
  interval = setInterval(() => updateUser(), 1000);
}

if (user.subscription_state === "ACTIVE" && interval) {
  clearInterval(interval);
}

if (user.subscription_state === "ACTIVE") {
  showSuccess = true;
}*/

  onMount(async () => {
    connectWs();
    stripe = await loadStripe("pk_test_51ITqIGItjdVuh2paNpnIUSWtsHJCLwY9fBYtiH2leQh2BvaMWB4de40Ea0ntC14nnmYcUyBD21LKO9ldlaXL6DJJ00Qm1toLdb");
    if (!$user.payment_method) {
      createCardForm();
    }
  });

  const connectWs = () => {
    const t = get(token);
    if (!t) {
      console.log("bump")
    }

    const ws = new WebSocket('ws://localhost/ws/users/me/payment', ["access_token", t]);
    ws.onmessage = function (event) {
      console.log(event.data);
      payments = JSON.parse(event.data);
    };
    ws.onclose = function(e) {
      console.log('Socket is closed. Reconnect will be attempted in 1 second.', e.reason);
      setTimeout(function() {
        connectWs();
      }, 1000);
    };
    ws.onerror = function(err) {
      console.error('Socket encountered error: ', err, 'Closing socket');
      ws.close();
    };
  }

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
    {#if $user.freq > 0}
    <form class="p-2" on:submit|preventDefault="{handleCancel}">
      <button class="btn my-4" disabled="{isSubmitting}" type="submit">Cancel current support
        {#if isSubmitting}
          <Dots />
        {/if}
      </button>
    </form>
    {:else}
     (currently not supporting)
    {/if}
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

<div class="container">
  <h2 class="p-2">
    Payment History
  </h2>
</div>
<div class="container">
<table>
  <thead>
  <tr>
    <th>Payment Cycle</th>
    <th>Balance</th>
    <th>Type</th>
    <th>Day</th>
  </tr>
  </thead>
  <tbody>
  {#if payments}
    {#each payments as row}
      <tr>
        <td>{row.paymentCycleId}</td>
        <td>{row.balance / 1000000} USD</td>
        <td>{row.balanceType}</td>
        <td>{row.day}</td>
      </tr>
    {:else}
      <tr><td colspan="5">No Data</td></tr>
    {/each}
  {:else}
    <tr><td colspan="5">No Data</td></tr>
  {/if}
  </tbody>
</table>
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
