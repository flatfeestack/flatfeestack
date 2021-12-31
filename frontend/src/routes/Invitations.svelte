<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error, isSubmitting, user, config } from "../ts/store";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import type { Invitation, UserStatus } from "../types/users.ts";
  import { faTrash, faSync, faClock, faCheck } from "@fortawesome/free-solid-svg-icons";
  import { formatDate, timeSince } from "../ts/services";
  import Dots from "../components/Dots.svelte";

  let invites: Invitation[] = [];
  let inviteEmail;
  let isAddInviteSubmitting = false;
  let selected;
  let statusSponsoredUsers: UserStatus[] = [];

  async function removeMyInvite(email: string, inviteEmail: string) {
    try {
      await API.invite.delMyInvite(inviteEmail);
      invites = invites.filter((inv: Invitation) => {
        return inv.email !== email || inv.inviteEmail !== inviteEmail;
      });
    } catch (e) {
      $error = e;
    }
  }

  async function removeByInvite(email: string, inviteEmail: string) {
    try {
      await API.invite.delByInvite(email);
      invites = invites.filter((inv: Invitation) => {
        return inv.email !== email || inv.inviteEmail !== inviteEmail;
      });
    } catch (e) {
      $error = e;
    }
  }

  async function acceptInvite(email: string) {
    try {
      await API.invite.confirmInvite(email);
      await refreshInvite();
    } catch (e) {
      $error = e;
    }
  }

  async function refreshInvite() {
    $isSubmitting = true;
    try {
      const res1 = await API.invite.invites();
      invites = res1 === null ? [] : res1;
    } catch (e) {
      $error = e;
    } finally {
      $isSubmitting = false;
    }
  }

  async function addInvite() {
    try {
      isAddInviteSubmitting = true;
      const res1 = API.invite.inviteAuth(inviteEmail, selected);
      const res2 = API.invite.invite(inviteEmail, selected);
      const inv: Invitation = { email: $user.email, inviteEmail, freq: selected, createdAt: new Date().toISOString(), confirmedAt: null };
      invites = [...invites, inv];
      await res1
      await res2;
    } catch (e) {
      $error = e;
    } finally {
      isAddInviteSubmitting = false;
    }
  }

  function daysLeft(email) {
    const result = statusSponsoredUsers.find(e => e.email === email);
    if(!result) {
      return "?"
    }
    return result;
  }

  onMount(async () => {
    const pr1 = refreshInvite();
    const pr2 = API.user.statusSponsoredUsers();
    const res2 = await pr2;
    statusSponsoredUsers = res2 === null ? [] : res2;
    await pr1;
  });

</script>

<Navigation>
  <h2 class="p-2 m-2">Invite Users</h2>
  <p class="p-2 m-2">Invite your friends or co-workers. They will be prefunded from your account on a regular basis.</p>

  <div class="container">
    <table>
      <thead>
      <tr>
        <th>Invited</th>
        <th>Status</th>
        <th>Date</th>
        <th>Plan</th>
        <th>Days Left</th>
        <th>Remove</th>
        <th><span class="cursor-pointer" on:click="{refreshInvite}"><Fa icon="{faSync}" size="md" /></span></th>
      </tr>
      </thead>
      <tbody>
      {#each invites as inv, key (inv.email+inv.inviteEmail)}
        {#if (inv.email === $user.email)}
        <tr>
          <td>{inv.inviteEmail}</td>
          <td class="text-center">
            {#if inv.confirmedAt}
              <Fa icon="{faCheck}" size="md" />
            {:else}
              <Fa icon="{faClock}" size="md" />
            {/if}
          </td>
          <td title="{formatDate(new Date(inv.createdAt))}">
            {timeSince(new Date(inv.createdAt), new Date())} ago
          </td>
          <td>{$config.plans.find(plan => plan.freq == inv.freq).title}</td>
          <td>{daysLeft(inv.email)}</td>
          <td class="text-center" colspan="2">
            <span class="cursor-pointer" on:click="{() => removeMyInvite(inv.email, inv.inviteEmail)}"><Fa icon="{faTrash}" size="md" /></span>
          </td>
        </tr>
        {/if}
      {/each}
      <tr>
        <td colspan="7">
          <form on:submit|preventDefault="{addInvite}" class="container-small">
            <label class="p-2">Invite by email:</label>
            <input size="24" maxlength="50" type="email" bind:value="{inviteEmail}" />&nbsp;
            <select bind:value={selected}>
              {#each $config.plans as plan, i}
                <option value="{plan.freq}">{plan.title}</option>
              {/each}
            </select>
            <button class="ml-5 p-2 button1" type="submit" disabled="{isAddInviteSubmitting}">Invite
              {#if isAddInviteSubmitting}
                <Dots />
              {/if}
            </button>
          </form>
        </td>
      </tr>
      </tbody>
    </table>
  </div>

  <h2 class="p-2 m-2">Invited By</h2>
  <p class="p-2 m-2">Accept your invitation and fund your account.</p>

  <div class="container">
    <table>
      <thead>
      <tr>
        <th>Invited By</th>
        <th>Status</th>
        <th>Date</th>
        <th>Plan</th>
        <th>Days Left</th>
        <th>Action</th>
        <th><span class="cursor-pointer" on:click="{refreshInvite}"><Fa icon="{faSync}" size="md" /></span></th>
      </tr>
      </thead>
      <tbody>
      {#each invites as inv, key (inv.email+inv.inviteEmail)}
        {#if (inv.inviteEmail === $user.email)}
        <tr>
          <td>{inv.email}</td>
          <td class="text-center">
            {#if inv.confirmedAt}
              <Fa icon="{faCheck}" size="md" />
            {:else}
              <Fa icon="{faClock}" size="md" />
            {/if}
          </td>
          <td title="{formatDate(new Date(inv.createdAt))}">
            {timeSince(new Date(inv.createdAt), new Date())} ago
          </td>
          <td>{$config.plans.find(plan => plan.freq == inv.freq).title}</td>
          <td>{daysLeft(inv.email)}</td>
          <td class="text-center" colspan="2">
            <span class="cursor-pointer" on:click="{() => removeByInvite(inv.email, inv.inviteEmail)}"><Fa icon="{faTrash}" size="md" /></span>
            {#if !inv.confirmedAt} <span class="cursor-pointer" on:click="{() => acceptInvite(inv.email)}"><Fa icon="{faCheck}" size="md" /></span> {/if}
          </td>
        </tr>
        {/if}
      {/each}
      </tbody>
    </table>
  </div>

</Navigation>
