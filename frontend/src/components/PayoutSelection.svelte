<script>
  import NEO from "./PayoutTabs/NEO.svelte";
  import Tezos from "./PayoutTabs/Tezos.svelte";
  import Web3 from "./PayoutTabs/Web3.svelte";
  import Tabs from "./Tabs.svelte";

  // List of tab items with labels, values and assigned components
  let items = [
    {
      label: "Ethereum",
      value: 1,
      component: Web3,
    },
    {
      label: "NEO",
      value: 2,
      component: NEO,
    },
    {
      label: "Tezos",
      value: 3,
      component: Tezos,
    },
  ];
</script>

<Tabs {items} />
