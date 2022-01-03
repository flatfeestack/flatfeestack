<script lang="ts">
    import {onMount} from "svelte";
    import {error, user, config, userBalances} from "../../ts/store";
    import Dots from "../Dots.svelte";
    import {formatBalance, stripePayment, stripePaymentMethod} from "../../ts/services";
    import {loadStripe} from "@stripe/stripe-js/pure";
    import {API} from "../../ts/api";

    export let remaining: number;
    export let current: number;
    export let seats: number;
    export let freq: number;

    let stripe;
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
        if (!cardElement) {
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
            if (!$user.payment_method) {
                await stripePaymentMethod(stripe, cardElement);
            }
            await stripePayment(stripe, freq, seats, $user.payment_method);
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
            const p1 = API.user.deletePaymentMethod()
            const p2 = API.user.cancelSub()
            $user.payment_method = null;
            $user.last4 = null;
            await p1;
            await p2;
        } catch (e) {
            $error = e;
        } finally {
            isSubmitting = false;
        }
    }

    onMount(async () => {
        stripe = await loadStripe($config.stripePublicApi);
        createCardForm();
    });

</script>

<style>
    .w25 {
        width: 25rem;
    }

    @media (max-width: 36rem) {
        .w25 {
            width: 20rem;
        }
    }
</style>

{#if $user.payment_method}
<div class="container">
    <label class="nobreak">Credit card: </label>
    <div class="container">
        <span>*** *** *** {$user.last4}</span>
        <form class="p-2" on:submit|preventDefault="{deletePaymentMethod}">
            <button class="button3" disabled="{isSubmitting}" type="submit">Cancel
                {#if isSubmitting}
                    <Dots/>
                {/if}
            </button>
        </form>
    </div>
</div>
{/if}

{#if $userBalances && $userBalances.total}
<div class="container">
<label class="nobreak">Current balance: </label>
    <div class="container">
        {formatBalance($userBalances.total["USD"],"USD")}
    </div>
</div>
{/if}

<div class="container">
    <div class="p-2 m-2 w25 rounded border-primary-700" bind:this="{card}"></div>
    <div class="p-2">
        <form on:submit|preventDefault="{handleSubmit}">
            <button class="button1" disabled="{isSubmitting || remaining < (current / 2)}" type="submit">❤&nbsp;Support
                {#if isSubmitting}
                    <Dots/>
                {/if}
            </button>
            {#if remaining < (current / 2)}
                (No need to top-up your account, you still funds)
            {:else}
                for ${remaining.toFixed(2)}
            {/if}
            </form>
    </div>
</div>

{#if showSuccess}<div class="p-2">Payment successful sent</div>{/if}
{#if paymentProcessing}<div class="p-2">Verifying payment<Dots/></div>{/if}
