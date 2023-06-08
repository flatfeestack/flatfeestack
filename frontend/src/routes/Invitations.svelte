<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error, isSubmitting, user } from "../ts/mainStore";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import type { Invitation, UserStatus } from "../types/backend";
  import {
    faTrash,
    faSync,
    faClock,
    faCheck,
  } from "@fortawesome/free-solid-svg-icons";
  import { formatDate, timeSince } from "../ts/services";
  import Dots from "../components/Dots.svelte";
  import { emailValidationPattern } from "../ts/utils";

  let invites: Invitation[] = [];
  let inviteEmail: string;
  let isAddInviteSubmitting = false;
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
      invites = res1 || [];
    } catch (e) {
      $error = e;
    } finally {
      $isSubmitting = false;
    }
  }

  async function addInvite() {
    try {
      isAddInviteSubmitting = true;
      if (inviteEmail === $user.email) {
        throw "Oops something went wrong. You aren't able to invite yourself.";
      }
      await API.invite.invite(inviteEmail);
      await API.invite.inviteAuth(inviteEmail);
      const inv: Invitation = {
        email: $user.email,
        inviteEmail,
        createdAt: new Date().toISOString(),
        confirmedAt: null,
      };
      invites = [...invites, inv];
      inviteEmail = "";
    } catch (e) {
      $error = e;
    } finally {
      isAddInviteSubmitting = false;
    }
  }

  onMount(async () => {
    await refreshInvite();
    const res2 = await API.user.statusSponsoredUsers();
    statusSponsoredUsers = res2 || [];
  });
</script>

<style>
  @media screen and (max-width: 600px) {
    table {
      width: 100%;
    }
    table thead {
      border: none;
      clip: rect(0 0 0 0);
      height: 1px;
      margin: -1px;
      overflow: hidden;
      padding: 0;
      position: absolute;
      width: 1px;
    }

    table tr {
      border-bottom: 3px solid #fff;
      display: block;
    }

    table td {
      border-bottom: 1px solid #fff;
      display: block;
      font-size: 0.8em;
      text-align: right;
    }

    table td::before {
      content: attr(data-label);
      float: left;
      font-weight: bold;
      text-transform: uppercase;
    }

    table td:last-child {
      border-bottom: 0;
    }
    table form {
      text-align: center;
      display: flex;
      flex-direction: column;
    }
    table form button {
      margin: 0.5rem 0;
    }
  }
</style>

<Navigation>
  <h2 class="p-2 m-2">Invite Users</h2>
  <p class="p-2 m-2">
    Invite your friends or co-workers. They will be charged from your account on
    a daily basis.
  </p>

  <div class="container">
    <table>
      <thead>
        <tr>
          <th>Invited</th>
          <th>Status</th>
          <th>Date</th>
          <th>Remove</th>
          <th>
            <button class="accessible-btn" on:click={refreshInvite}>
              <Fa icon={faSync} size="md" />
            </button>
          </th>
        </tr>
      </thead>
      <tbody>
        {#each invites as inv, key (inv.email + inv.inviteEmail)}
          {#if inv.email === $user.email}
            <tr>
              <td data-label="Invited">{inv.inviteEmail}</td>
              <td data-label="Status" class="text-center">
                {#if inv.confirmedAt}
                  <Fa icon={faCheck} size="md" />
                {:else}
                  <Fa icon={faClock} size="md" />
                {/if}
              </td>
              <td data-label="Date" title={formatDate(new Date(inv.createdAt))}>
                {timeSince(new Date(inv.createdAt), new Date())} ago
              </td>
              <td data-label="Remove" class="text-center" colspan="2">
                <button
                  class="accessible-btn"
                  on:click={() => removeMyInvite(inv.email, inv.inviteEmail)}
                  ><Fa icon={faTrash} size="md" /></button
                >
              </td>
            </tr>
          {/if}
        {/each}
        <tr>
          <td colspan="5">
            <form on:submit|preventDefault={addInvite}>
              <label for="invite-mail-input" class="p-2">Invite by email:</label
              >
              <input
                id="invite-mail-input"
                size="24"
                maxlength="50"
                type="email"
                pattern={emailValidationPattern}
                required
                bind:value={inviteEmail}
              />
              <button
                class="ml-5 p-2 button1"
                type="submit"
                disabled={isAddInviteSubmitting}
                >Invite
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
          <th>Action</th>
          <th
            ><button class="accessible-btn" on:click={refreshInvite}
              ><Fa icon={faSync} size="md" /></button
            ></th
          >
        </tr>
      </thead>
      <tbody>
        {#each invites as inv, key (inv.email + inv.inviteEmail)}
          {#if inv.inviteEmail === $user.email}
            <tr>
              <td data-label="Invited By">{inv.email}</td>
              <td data-label="Status" class="text-center">
                {#if inv.confirmedAt}
                  <Fa icon={faCheck} size="md" />
                {:else}
                  <Fa icon={faClock} size="md" />
                {/if}
              </td>
              <td data-label="Date" title={formatDate(new Date(inv.createdAt))}>
                {timeSince(new Date(inv.createdAt), new Date())} ago
              </td>
              <td data-label="Action" class="text-center" colspan="2">
                <button
                  class="accessible-btn"
                  on:click={() => removeByInvite(inv.email, inv.inviteEmail)}
                  ><Fa icon={faTrash} size="md" /></button
                >
                {#if !inv.confirmedAt}
                  <button
                    class="accessible-btn"
                    on:click={() => acceptInvite(inv.email)}
                    ><Fa icon={faCheck} size="md" /></button
                  >
                {/if}
              </td>
            </tr>
          {/if}
        {/each}
      </tbody>
    </table>
  </div>
</Navigation>
