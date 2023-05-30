<script>
  import { Route } from "svelte-routing";
  import PrivateRouteGuard from "./PrivateRouteGuard.svelte";

  export let path;
  export let admin = false;
</script>

<Route {path} {admin} let:params let:location>
  <PrivateRouteGuard {admin}>
    <slot {params} {location} />
  </PrivateRouteGuard>
</Route>
