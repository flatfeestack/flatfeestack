<script>
    import { config } from "../../ts/store";
    let currencies = [
        { id: 1, name: `Ethereum` },
        { id: 2, name: `NEO` },
        { id: 3, name: `Tezos` }
    ];

    let selected;

    async function handleSubmit() {
        alert(selected.name);
        let url = 'http://...'
        const response = await fetch(url, {
            method: 'POST'})
    }
</script>

<h2>Select your cryptocurrency</h2>

<form on:submit|preventDefault={handleSubmit}>
    <select bind:value={selected}>
        {#each $config.supportedCurrencies as currency}
            <option value={currency}>
                {currency.name}
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
</style>