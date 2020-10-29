<script>
import DashboardLayout from "../../layout/DashboardLayout.svelte";
import { user } from "src/store/auth";
import Payment from "../../components/Payment.svelte";
</script>

<DashboardLayout>
  <h1 class="text-primary-500">Dashboard</h1>
  <div class="flex">
    <div class="w-full sm:w-1/3 shadow-lg bg-white p-5">
      <h2 class="">Subscription</h2>

      {#if !$user.subscription}
        <div class="">You don't have an active subscription</div>
        <Payment />
      {:else}
        <div>Your subscription {$user.subscription}</div>
      {/if}
    </div>
  </div>
</DashboardLayout>
