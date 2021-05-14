<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import Payment from "../components/Payment.svelte";
  import { error, user, userBalances, token, config } from "../ts/store";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import { faSync } from "@fortawesome/free-solid-svg-icons";
  import type { Repo } from "src/types/users";
  import { connectWs, formatDate, parseJwt} from "../ts/services";
  import Dots from "../components/Dots.svelte";

  let checked = $user.role != "ORG";
  let sponsoredRepos: Repo[] = [];
  let invite_email;
  let isSubmitting = false;

  $: $user.role = checked === false ? "ORG" : "USR";

  async function topupInvite() {
    isSubmitting = true;
    try {
      await API.user.topup();
    } catch (e) {
      $error = e;
    } finally {
      isSubmitting = false;
    }
  }

  async function handleCancel() {
    isSubmitting = true;
    try {
      await API.user.cancelSub();
      $userBalances.paymentCycle.freq = 0;
    } catch (e) {
      $error = e;
    } finally {
      isSubmitting = false;
    }
  }

  async function updateSeats() {
    isSubmitting = true;
    try {
      await API.user.updateSeats($userBalances.paymentCycle.seats)
    } catch (e) {
      $error = e;
    } finally {
      isSubmitting = false;
    }
  }

  async function deletePaymentMethod() {
    isSubmitting = true;
    try {
      await API.user.deletePaymentMethod()
      $user.payment_method = null;
      $user.last4 = null;
    } catch (e) {
      $error = e;
    } finally {
      isSubmitting = false;
    }
  };

  onMount(async () => {
    try {
      await connectWs();
      const res2 = await API.user.getSponsored();
      sponsoredRepos = res2 === null ? [] : res2;
    } catch (e) {
      $error = e;
    }
  });
</script>

<Navigation>
  <h1 class="px-2">Payments</h1>

  <div class="container">
    <label class="px-2">Current Balance: </label>
    <span class="bold">{$userBalances && $userBalances.total > 0 ? $userBalances.total / 1000000 : 0} USD</span>
  </div>

  <div class="container">
    <label class="px-2">Current Support: </label>
    {#if $userBalances && $userBalances.paymentCycle && $userBalances.paymentCycle.seats > 0}
      <span class="bold">
        <input size="5" type="number" min="1" bind:value={$userBalances.paymentCycle.seats}> seats,
        {$config.plans.find(plan => plan.freq == $userBalances.paymentCycle.freq).title.toLocaleLowerCase()} recurring payments
        (${$userBalances.paymentCycle.seats * $userBalances.paymentCycle.freq * 330000 / 1000000})</span>
      <form class="p-2" on:submit|preventDefault="{updateSeats}">
        <button disabled="{isSubmitting}" type="submit">Update Seats
          {#if isSubmitting}
            <Dots />
          {/if}
        </button>
      </form>
      <form class="p-2" on:submit|preventDefault="{handleCancel}">
        <button disabled="{isSubmitting}" type="submit">Cancel&nbsp;Support
          {#if isSubmitting}
            <Dots />
          {/if}
        </button>
      </form>
    {:else}
      <span>n/a</span>
    {/if}
  </div>

  <div class="container">
    <label class="px-2">Credit card: </label>
    {#if $user.payment_method}
      <span>*** {$user.last4}</span>
      <form class="p-2" on:submit|preventDefault="{deletePaymentMethod}">
        <button disabled="{isSubmitting}" type="submit">Remove card
          {#if isSubmitting}
            <Dots />
          {/if}
        </button>
      </form>
    {:else}
      <span>n/a</span>
    {/if}
  </div>

  {#if checked && parseJwt($token).inviteEmails}
    <div class="container">
      <label class="px-2">Topup from invites: </label>
      <span class="cursor-pointer" on:click="{topupInvite}"><Fa icon="{faSync}" size="md" /></span>
    </div>
  {/if}

  {#if checked}
    <div class="container">
      <label class="px-2">Selected Projects:</label>
      <span class="bold">{sponsoredRepos.length} projects</span>
    </div>
  {/if}

  {#if !($userBalances && $userBalances.paymentCycle && $userBalances.paymentCycle.seats > 0)|| !checked}
    {#if !($userBalances && $userBalances.daysLeft>1) || !checked}
      <Payment />
    {/if}
  {/if}

  {#if $userBalances && $userBalances.userBalances}
    <div class="container">
      <h2 class="p-2">
        Payment History
      </h2>
    </div>
    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Payment Cycle</th>
          <th>Balance</th>
          <th>Type</th>
          <th>Day</th>
        </tr>
        </thead>
        <tbody>

        {#each $userBalances.userBalances as row}
          <tr>
            <td>{row.paymentCycleId}</td>
            <td>{row.balance / 1000000} USD</td>
            <td>{row.balanceType}</td>
            <td>{formatDate(new Date(row.createdAt))}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="5">No Data</td>
          </tr>
        {/each}

        </tbody>
      </table>
    </div>
  {/if}

</Navigation>
