<script lang="ts">
  import queryString from "query-string";
  import {route} from "@mateothegreat/svelte5-router";

  let { location } = $props<{ location: any }>();

  let queryParams = $state<any>();
  $effect(() => { queryParams = queryString.parse(location.search);});
</script>

<div class="centerContainer">
  <div class="w-50">
    <h1 class="text-secondary-900">Different chain IDs detected</h1>
    <p>
      Your MetaMask browser plugin is currently connected to the network with ID {queryParams.actual}.
      However, FlatFeeStack requires to use {queryParams.required}. Please
      change your network in MetaMask and
      <a use:route href="/">click here</a> to return to the previous page.
    </p>
  </div>
</div>
