<script>
    import FiatTab from "./FiatTab.svelte";
    import CryptoTab from "./CryptoTab.svelte";
    import Tabs from "./Tabs.svelte";

    // List of tab items with labels, values and assigned components
    let items = [
        { label: "Fiat",
            value: 1,
            component: FiatTab
        },
        { label: "Crypto",
            value: 2,
            component: CryptoTab
        }
    ];
</script>

<Tabs {items} />