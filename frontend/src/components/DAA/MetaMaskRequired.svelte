<script lang="ts">
  import { onMount } from "svelte";
  import detectEthereumProvider from "@metamask/detect-provider";
  import { navigate } from "svelte-routing";

  onMount(async () => {
    try {
      const provider = await detectEthereumProvider();
      if (provider !== null) {
        navigate("/daa/home");
      }
    } catch {}
  });
</script>

<div class="centerContainer">
  <div class="w-50">
    <h1 class="text-secondary-900">MetaMask missing</h1>
    <p>
      You need MetaMask to access this site (even if you choose to not connect
      your account with FlatFeeStack DAA). Please <a
        href="https://metamask.io/"
        rel="noopener noreferrer"
        target="_blank">download the extension</a
      > and activate it.
    </p>
  </div>
</div>
