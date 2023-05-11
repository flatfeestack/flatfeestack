<script lang="ts">
  import queryString from "query-string";

  export let location;

  let queryParams;
  $: queryParams = queryString.parse(location.search);
</script>

<div class="centerContainer">
  <div class="w-50">
    <h1 class="text-secondary-900">Different chain IDs detected</h1>
    <p>
      Your MetaMask browser plugin is currently connected to the network with ID {queryParams.actual}.
      However, FlatFeeStack requires to use {queryParams.required}. Please
      change your network in MetaMask and
      <a href="javascript:history.back()">click here</a> to return to the previous
      page.
    </p>
  </div>
</div>
