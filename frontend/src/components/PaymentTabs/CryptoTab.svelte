<script>
    import {config, user, error} from "../../ts/store";
    import {onMount} from "svelte";
    import {loadStripe} from "@stripe/stripe-js/pure";
    import {API} from "../../ts/api";

    $: {
        if (card) {
            if ($user.payment_method) {
                card.style.display = "none";
            } else {
                card.style.display = "block";
            }
        }
    }

    let selected;
    let cardElement;
    let stripe;
    let card; // HTML div to mount card
    let seats = 1;
    let selectedPlan = 0;

    async function handleSubmit() {
        let response = await API.user.nowpaymentsPayment(selected.shortName, $config.plans[selectedPlan].freq, seats);
        let json = await response.json();
        window.open(json.invoice_url, "", "width=800,height=800");
    }

    function createCardForm() {
        if(!cardElement) {
            cardElement.on("change", (e) => {
                if (e.error) {
                    $error = e.error;
                }
            });
        }
    }

    onMount(async () => {
        createCardForm();
    });
</script>

<h2>Select your cryptocurrency</h2>

<form on:submit|preventDefault={handleSubmit}>
    <div class="container-stretch">
        {#each $config.plans as { title, desc, disclaimer }, i}
            <div
                    class="child p-2 m-2 w1-2 card cursor-pointer border-primary-500 rounded {selectedPlan === i ? 'bg-green' : ''}"
                    on:click="{() => (selectedPlan = i)}">
                <h3 class="text-center font-bold text-lg">{title}</h3>
                <div class="text-center">{@html desc}</div>
                <div class="small text-center">{@html disclaimer}</div>
            </div>
        {/each}
    </div>

    {#if $user.role === "ORG" }
        <div class="p-2">
            <span>How many seats?</span>
            <input size="5" type="number" min="1" bind:value={seats}>
        </div>
    {/if}
    <select bind:value={selected}>
        {#each $config.supportedCurrencies as currency}
            <option value={currency}>
                {currency.name} - {currency.shortName}
            </option>
        {/each}
    </select>

    <button disabled={!selected} type=submit>
        Submit
    </button>
</form>

<style>
    input {
        display: block;
        width: 500px;
        max-width: 100%;
    }

    .small {
        font-size: x-small;
    }
</style>