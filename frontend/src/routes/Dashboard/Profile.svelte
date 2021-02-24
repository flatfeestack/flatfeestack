<script>
import DashboardLayout from "./DashboardLayout.svelte";
import Payment from "../../components/Payment.svelte";
import { user } from "ts/auth";
let checked = $user.mode != "ORG";
$: {
  if (checked == false) {
    $user.mode = "ORG";
  } else {
    $user.mode = "USR";
  }
}
</script>

<style>
    .container {
        display: flex;
        flex-direction: row;
        margin: 1em;
    }
    .wrap {
        display: flex;
        flex-wrap: wrap;
    }
</style>

<DashboardLayout>
  <h1 class = "px-2">Profile</h1>

  <div class = "container px-2">
    <label class = "">Are you an organization or an individual contributor? </label>
    <div class="onoffswitch">
    <input type="checkbox" bind:checked={checked} name="onoffswitch" class="onoffswitch-checkbox" id="myonoffswitch" tabindex="0" >
    <label class="onoffswitch-label" for="myonoffswitch">
      <span class="onoffswitch-inner"></span>
      <span class="onoffswitch-switch"></span>
    </label>
  </div>
  </div>

  {#if checked}
  <div class = "container px-2">
    <label class = "">What name should appear on your badge? </label>
    <input type="text">
    Upload your profile picture
    <input type="file">
  </div>
  {:else}
    <div class = "container px-2">
    <label class = "">What is the name of your organization? </label>
    <input type="text">
    Upload your logo
    <input type="file">
    </div>
  {/if}

  <div class="flex">
      <h2 class="Sponsoring">
        {#if $user.subscription_state}
          You are currently sponsoring 5 projects
          <span>{$user.subscription_state}</span>
        {:else}
          Support these awesome 5 projects
        {/if}
      </h2>

      {#if $user.subscription_state !== 'ACTIVE'}
        <div class="container bg-green rounded p-2 my-4">
          You don't have an active subscription.
          <br />Please choose a plan below to start sponsoring open source
          projects
        </div>
      {/if}
      <Payment />
    </div>


</DashboardLayout>
