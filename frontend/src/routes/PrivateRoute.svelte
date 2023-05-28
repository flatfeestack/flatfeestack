<script>
  import { Route } from "svelte-routing";
  import PrivateRouteGuard from "./PrivateRouteGuard.svelte";

  export let path;
</script>

<Route {path} let:params let:location>
  <PrivateRouteGuard>
    <slot {params} {location} />
  </PrivateRouteGuard>
</Route>
