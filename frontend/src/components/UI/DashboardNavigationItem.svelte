<style type="text/scss">
</style>

<script lang="ts">
import Fa from "svelte-fa";
export let icon;
export let route;
export let label;

import { getContext } from "svelte";
import { Link } from "svelte-routing";
import { ROUTER } from "svelte-routing/src/contexts";

const { activeRoute } = getContext(ROUTER);
const active = $activeRoute?.uri;
</script>

<li class="mr-3 flex-1">
  <Link to="{route}">
    <div
      class="block py-1 md:py-3 pl-1 align-middle text-white no-underline hover:text-white border-b-2 border-grey-darkest hover:border-pink"
    >
      <div
        class="flex align-center hover:bg-primary-100 hover:text-gray-600 active:bg-secondary-500 p-3 rounded-lg {active === route ? 'bg-primary-500 text-white' : 'text-gray-500'}"
      >
        <Fa icon="{icon}" size="lg" />
        <div class="pl-2">{label}</div>
      </div>
    </div>
  </Link>
</li>
