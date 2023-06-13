<script lang="ts">
  import { navigate } from "svelte-routing";
  import { chainId, lastEthRoute } from "../ts/ethStore";

  export let requiredChainId: number | undefined;

  $: {
    if (
      requiredChainId !== undefined &&
      $chainId !== null &&
      $chainId !== requiredChainId
    ) {
      $lastEthRoute = window.location.pathname;
      navigate(
        `/differentChainId?required=${requiredChainId}&actual=${$chainId}`
      );
    }
  }
</script>
