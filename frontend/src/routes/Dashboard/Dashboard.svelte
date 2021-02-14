<script>
import DashboardLayout from "./DashboardLayout.svelte";
import { user } from "ts/auth";
import Payment from "../../components/Payment.svelte";
</script>



<DashboardLayout>
  <h1 class="text-primary-500">Dashboard</h1>
  <div class="flex">
    <div class="w-full sm:w-2/3 shadow-lg bg-white p-5">
      <h2 class="">
        Subscription
        {#if $user.subscription_state}
          <span
            class="bg-secondary-500 p-1 rounded-lg text-white"
          >{$user.subscription_state}</span>
        {/if}
      </h2>

      {#if $user.subscription_state !== 'ACTIVE'}
        <div class="text-gray-700 mb-10">
          You don't have an active subscription.
          <br />Please choose a plan below to start sponsoring open source
          projects
        </div>
      {/if}
      <Payment />
    </div>
  </div>
</DashboardLayout>
