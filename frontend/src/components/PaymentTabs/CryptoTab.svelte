<script>
    import {config, error} from "../../ts/store";
    import {API} from "../../ts/api";

    export let total;
    export let selectedPlan;
    export let seats;

    let selected;

    async function handleSubmit() {
        try {
            let response = await API.user.nowpaymentsPayment(selected.shortName, $config.plans[selectedPlan].freq, seats);
            let json = await response.json();
            window.open(json.invoice_url, "", "width=800,height=800");
        } catch (e) {
            $error = e;
        }
    }
</script>

<form on:submit|preventDefault={handleSubmit}>
    <div class="container">
        <div class="p-2">
            <select bind:value={selected}>
                {#each $config.supportedCurrencies || [] as currency}
                    <option value={currency}>
                        {currency.name} - {currency.shortName}
                    </option>
                {/each}
            </select>
        </div>
        <div class="p-2">
            <button class="button1" type="submit" disabled="{total < 10}">❤&nbsp;Support</button>
        </div>
    </div>
</form>

<style>
    input {
        display: block;
        width: 500px;
        max-width: 100%;
    }
</style>