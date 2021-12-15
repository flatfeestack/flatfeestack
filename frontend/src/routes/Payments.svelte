<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error, user, userBalances, token, config } from "../ts/store";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import { faSync } from "@fortawesome/free-solid-svg-icons";
  import type {Repo} from "../types/users";
  import { connectWs, formatDate, parseJwt} from "../ts/services";
  import Dots from "../components/Dots.svelte";
  import { navigate } from "svelte-routing";
  import PaymentSelection from "../components/PaymentSelection.svelte";

  let sponsoredRepos: Repo[] = [];
  let invite_email;
  let isSubmitting = false;

  let current;
  $: {
    current = $userBalances && $userBalances.total > 0 ? $userBalances.total : 0
  }

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
  }

  onMount(async () => {
    try {
      const pr1 = connectWs();
      const pr2 = API.user.getSponsored();
      const res2 = await pr2;
      sponsoredRepos = res2 === null ? [] : res2;
      await pr1;
    } catch (e) {
      $error = e;
    }
  });
</script>

<Navigation>
  <h2 class="p-2 m-2">Sponsor Summary</h2>
  {#if $userBalances && $userBalances.total}
    <div class="container">
      <h2 class="p-2">
        Current Balance:
      </h2>
    </div>
    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Currency</th>
          <th>Balance</th>
        </tr>
        </thead>
        <tbody>

        {#each $userBalances.total as row}
          <tr>
            <td>{row.currency}</td>
            <td>{row.balance}</td>
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


  <div class="grid-2">
    {#if $userBalances && $userBalances.paymentCycle && $userBalances.paymentCycle.seats > 0 && $userBalances.paymentCycle.freq > 0}
      <label class="nobreak">Current Recurring Support: </label>
      <div class="container">

      <span class="bold">
        <input size="5" type="number" min="1" bind:value={$userBalances.paymentCycle.seats}> seats,
        {$config.plans.find(plan => plan.freq == $userBalances.paymentCycle.freq).title.toLocaleLowerCase()} recurring payments
        (${$userBalances.paymentCycle.seats * $config.plans.find(plan => plan.freq == $userBalances.paymentCycle.freq).price})
      </span>
          <form class="p-2" on:submit|preventDefault="{updateSeats}">
            <button class="button2" disabled="{isSubmitting}" type="submit">Update Seats
              {#if isSubmitting}
                <Dots/>
              {/if}
            </button>
          </form>
        <form class="p-2" on:submit|preventDefault="{handleCancel}">
          <button class="button2" disabled="{isSubmitting}" type="submit">Cancel&nbsp;Support
            {#if isSubmitting}
              <Dots/>
            {/if}
          </button>
        </form>
      </div>
    {/if}

    {#if $user.payment_method}
      <label class="nobreak">Credit card: </label>
      <div class="container">
        <span>*** {$user.last4}</span>
        <form class="p-2" on:submit|preventDefault="{deletePaymentMethod}">
          <button class="button2" disabled="{isSubmitting}" type="submit">Remove card
            {#if isSubmitting}
              <Dots/>
            {/if}
          </button>
        </form>
      </div>
    {/if}

    {#if parseJwt($token).inviteEmails}
      <label class="nobreak">Topup from invites: </label>
      <div class="container">
        <span class="cursor-pointer" on:click="{topupInvite}"><Fa icon="{faSync}" size="md" /></span>
      </div>
    {/if}


    <label class="nobreak">Selected Projects:</label>
    <div class="container">
      <span class="bold">{sponsoredRepos.length} projects</span>
    </div>

  </div>

  {#if !$userBalances || !$userBalances.paymentCycle || $userBalances.paymentCycle.freq === 0}
      <PaymentSelection />
  {/if}

  {#if $userBalances && $userBalances.userBalances}
    <h2 class="p-2 m-2">Payment History</h2>
    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Payment Cycle</th>
          <th>Balance</th>
          <th>Currency</th>
          <th>Type</th>
          <th>Day</th>
        </tr>
        </thead>
        <tbody>

        {#each $userBalances.userBalances as row}
          <tr>
            <td>{row.paymentCycleId}</td>
            <td>{row.balance}</td>
            <td>{row.currency}</td>
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
