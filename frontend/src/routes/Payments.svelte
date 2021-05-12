<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import Payment from "../components/Payment.svelte";
  import { error, user, userBalances } from "../ts/store";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import type { Invitation } from "src/types/users.ts";
  import { faTrash } from "@fortawesome/free-solid-svg-icons";
  import type { Repo } from "src/types/users";
  import { links } from "svelte-routing";
  import { connectWs } from "../ts/services";
  import Dots from "../components/Dots.svelte";

  let checked = $user.role != "ORG";
  let invites: Invitation[] = [];
  let sponsoredRepos: Repo[] = [];
  let invite_email;
  let isAddInviteSubmitting = false;

  $: $user.role = checked === false ? "ORG" : "USR";

  $: {
    if ($userBalances.paymentCycle) {
      $user.paymentCycleId = $userBalances.paymentCycle.id;
    }
  }

  async function removeInvite(email: string) {
    try {
      await API.authToken.delInvite(email);
      invites = invites.filter((inv: Invitation) => {
        return inv.email !== email;
      });
    } catch (e) {
      $error = e;
    }
  }

  async function refreshInvite() {
    try {
      const response = await API.authToken.invites();
      if (response && response.length > 0) {
        invites = response;
      }
    } catch (e) {
      $error = e;
    }
  }

  async function addInvite() {
    try {
      isAddInviteSubmitting = true;
      const d = new Date().toISOString();
      await API.authToken.invite(invite_email, $user.email, $user.name, d);
      const inv: Invitation = { email: invite_email, createdAt: d, pending: true };
      invites = [...invites, inv];
    } catch (e) {
      $error = e;
    } finally {
      isAddInviteSubmitting = false;
    }
  }

  onMount(async () => {
    connectWs();
    try {
      const res1 = await API.authToken.invites();
      const res2 = await API.user.getSponsored();
      invites = res1 === null ? [] : res1;
      sponsoredRepos = res2 === null ? [] : res2;
    } catch (e) {
      $error = e;
    }
  });

</script>

<style>
    .container {
        display: flex;
        flex-direction: row;
        margin: 1em;
        align-items: center;
    }

    .bold {
        font-weight: bold;
    }
</style>

<Navigation>
  <h1 class="px-2">Payments</h1>

  <div class="container">
    <label class="px-2">Current Balance: </label>
    <span class="bold">{$userBalances && $userBalances.total > 0 ? $userBalances.total / 1000000 : 0} USD</span>
  </div>

  {#if checked}
    <div class="container">
      {#if sponsoredRepos.length > 0}
        <label class="px-2">Selected Projects:</label>
        <span class="bold">{sponsoredRepos.length} projects</span>
      {:else}
        <div class="bg-green rounded p-2 my-4" use:links>
          <p>You are not supporting any projects yet. Please go to the <a href="/user/search">Search</a>
            section where you can add your favorite projects.</p>
        </div>
      {/if}
    </div>
  {/if}

  <Payment />

  {#if !checked}
    <h2 class="px-2">Invitations</h2>
    <form on:submit|preventDefault="{addInvite}" class="container">
      <label class="p-2">Invite Email</label>
      <input class="p-2" size="24" maxlength="100" type="email" bind:value="{invite_email}" />&nbsp;
      <button type="submit" disabled="{isAddInviteSubmitting}">Invite to {$user.name ? $user.name : "to your org"}{#if isAddInviteSubmitting}<Dots />{/if}</button>
    </form>

    <div class="container">
      {#each invites as inv, key (inv.email)}
        {inv.email}
        {inv.pending}
        {inv.createdAt}
        <div class="cursor-pointer transform hover:scale-105 duration-200" on:click="{() => removeInvite(inv.email)}">
          <Fa icon="{faTrash}" size="md" />
        </div>
      {/each}
    </div>
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
            <td>{row.day}</td>
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
