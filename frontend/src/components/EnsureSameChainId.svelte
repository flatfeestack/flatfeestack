<script lang="ts">
  import { navigate } from "svelte-routing";
  import { chainId } from "../ts/ethStore";

  export let requiredChainId: number | undefined;

  $: {
    if (
      requiredChainId !== undefined &&
      $chainId !== null &&
      $chainId !== requiredChainId
    ) {
      navigate(
        `/differentChainId?required=${requiredChainId}&actual=${$chainId}`
      );
    }
  }
</script>
