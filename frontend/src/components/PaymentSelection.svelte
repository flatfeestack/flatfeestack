<script>
    import FiatTab from "./PaymentTabs/FiatTab.svelte";
    import CryptoTab from "./PaymentTabs/CryptoTab.svelte";
    import Tabs from "./Tabs.svelte";

    // List of tab items with labels, values and assigned components
    let items = [
        { label: "Credit Card",
            value: 1,
            component: FiatTab
        },
        { label: "Crypto Currencies",
            value: 2,
            component: CryptoTab
        }
    ];


    import { user, config, userBalances } from "../ts/store";


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

    let total;
    $: {
        total = $config.plans[selectedPlan].price * seats;
    }

    let current;
    $: {
        current = $userBalances && $userBalances.total > 0 ? $userBalances.total / 1000000 : 0
    }

    let remaining;
    $: {
        const rem = total - current;
        remaining = rem > 0 ? rem:0;
    }


</script>

<style>
    .small {
        font-size: x-small;
    }
    .child {
        margin: 0.5em;
        box-shadow: 0.25em 0.25em 0.25em #e1e1e3;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
    }
</style>

<h2 class="p-2 m-2">Payment</h2>
<p class="p-2 m-2">We request your permission that we initiate a payment or a series of {$config.plans[selectedPlan].title.toLowerCase()}
    payments on your behalf of
    {total.toFixed(2)} USD. By continuing, I authorize FlatFeeStack to send instructions to the financial institution that issued my card to
    take payments from my card account in accordance with the terms of my agreement with you.</p>
<div class="container-stretch">
    {#each $config.plans as { title, desc, disclaimer }, i}
        <div class="flex-grow child p-2 m-2 w1-2 card cursor-pointer border-primary-500 rounded {selectedPlan === i ? 'bg-green' : ''}"
                on:click="{() => (selectedPlan = i)}">
            <h3 class="text-center font-bold text-lg">{title}</h3>
            <div class="text-center">{@html desc}</div>
            <div class="small text-center">{@html disclaimer}</div>
        </div>
    {/each}
</div>

<div class="container page">
    <div class="p-2">
        <input size="5" type="number" min="1" bind:value={seats}> Seats
    </div>
    <div class="p-2">
        {#if remaining >= 10}
            Total&nbsp;Sponsoring:<span class="bold">${remaining.toFixed(2)}</span>
            {#if current.toFixed(2) > 0}
                (${total.toFixed(2)} - ${current.toFixed(2)} [current balance] = ${remaining.toFixed(2)} [remaining
                payment])
            {/if}
        {/if}
    </div>
</div>

<div class="p-2 m-2">
    <Tabs {items} total={remaining.toFixed(2)} {selectedPlan} {seats}/>
</div>