<script lang="ts">
    import {config, error} from "../../ts/store";
    import {API} from "../../ts/api";

    export let remaining: number;
    export let current: number;
    export let seats: number;
    export let freq: number;

    let selected;
    let payInAddress = "";

    async function handleSubmit() {
        try {
            const paymentResponse = await API.user.nowpaymentsPayment(selected, freq, seats);
            payInAddress = paymentResponse.pay_address;
        } catch (e) {
            $error = e;
        }
    }
</script>

<style>
    input {
        display: block;
        width: 500px;
        max-width: 100%;
    }
</style>

<form on:submit|preventDefault={handleSubmit}>
    <div class="container">
        <div class="p-2">
            <select bind:value={selected}>
                {#each Object.entries($config.supportedCurrencies) as [key, value]}
                    {#if value.isCrypto}
                        <option value={key}>
                            {value.name}
                        </option>
                    {/if}
                {/each}
            </select>
        </div>
        <div class="p-2">
            <button class="button1" type="submit" disabled="{remaining < (current / 2)}">❤&nbsp;Support</button>
            {#if remaining < (current / 2)}
                (No need to top-up your account, you still funds)
            {:else}
                for ${remaining.toFixed(2)}
            {/if}
        </div>
    </div>
</form>
{#if payInAddress != ""}<div class="p-2">Pay in to this address: <b>{payInAddress}</b></div>{/if}

