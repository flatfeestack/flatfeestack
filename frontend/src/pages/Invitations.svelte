<script lang="ts">
  import { appState } from "ts/state.svelte.ts";
  import { API } from "../ts/api.ts";
  import { onMount } from "svelte";
  import type { Invitation } from "../types/backend.ts";
  import { formatDate, timeSince } from "../ts/services.svelte.ts";
  import Dots from "../Dots.svelte";
  import { emailValidationPattern } from "../utils.ts";
  import Main from "../Main.svelte";

  let invites = $state<Invitation[]>([]);
  let inviteEmail= $state("");
  let isAddInviteSubmitting = $state(false);

  async function removeMyInvite(email: string, inviteEmail: string) {
    try {
      await API.invite.delMyInvite(inviteEmail);
      invites = invites.filter((inv: Invitation) => {
        return inv.email !== email || inv.inviteEmail !== inviteEmail;
      });
    } catch (e) {
      appState.setError(e);
    }
  }

  async function removeByInvite(email: string, inviteEmail: string) {
    try {
      await API.invite.delByInvite(email);
      invites = invites.filter((inv: Invitation) => {
        return inv.email !== email || inv.inviteEmail !== inviteEmail;
      });
    } catch (e) {
      appState.setError(e);
    }
  }

  async function acceptInvite(email: string) {
    try {
      await API.invite.confirmInvite(email);
      await refreshInvite();
    } catch (e) {
      appState.setError(e);
    }
  }

  async function refreshInvite() {
    appState.isSubmitting = true;
    try {
      const res1 = await API.invite.invites();
      invites = res1 || [];
    } catch (e) {
      appState.setError(e);
    } finally {
      appState.isSubmitting = false;
    }
  }

  async function addInvite() {
    try {
      isAddInviteSubmitting = true;
      if (inviteEmail === appState.user.email) {
        appState.setError("Oops something went wrong. You aren't able to invite yourself.");
        return;
      }
      await API.invite.invite(inviteEmail);
      await API.invite.inviteAuth(inviteEmail);
      const inv: Invitation = {
        email: appState.user.email,
        inviteEmail,
        createdAt: new Date().toISOString(),
        confirmedAt: null,
      };
      invites = [...invites, inv];
      inviteEmail = "";
    } catch (e) {
      appState.setError(e);
    } finally {
      isAddInviteSubmitting = false;
    }
  }

  onMount(async () => {
    await refreshInvite();
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
  }
</style>

<Main>
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
            <button class="accessible-btn" onclick={refreshInvite} aria-label="Refresh invitation">
              <i class="fa-sync"></i>
            </button>
          </th>
        </tr>
      </thead>
      <tbody>
        {#each invites as inv (inv.email + inv.inviteEmail)}
          {#if inv.email === appState.user.email}
            <tr>
              <td data-label="Invited">{inv.inviteEmail}</td>
              <td data-label="Status" class="text-center">
                {#if inv.confirmedAt}
                  <i class="fa-check"></i>
                {:else}
                  <i class="fa-clock"></i>
                {/if}
              </td>
              <td data-label="Date" title={formatDate(new Date(inv.createdAt))}>
                {timeSince(new Date(inv.createdAt), new Date())} ago
              </td>
              <td data-label="Remove" class="text-center" colspan="2">
                <button class="accessible-btn" aria-label="Remove invitation"
                        onclick={() => removeMyInvite(inv.email, inv.inviteEmail)}>
                  <i class="fa-trash"></i>
                </button
                >
              </td>
            </tr>
          {/if}
        {/each}
        <tr>
          <td colspan="5">
            <div class="flex items-center">
              <label for="invite-mail-input" class="p-2">Invite by email:</label>
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
                      onclick={addInvite}
                      disabled={isAddInviteSubmitting || !inviteEmail || !inviteEmail.match(emailValidationPattern)}
                      aria-label="Send invitation"
              >
                Invite
                {#if isAddInviteSubmitting}
                  <Dots />
                {/if}
              </button>
            </div>
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
          <th>
            <button class="accessible-btn" onclick={refreshInvite} aria-label="Refresh invitations">
              <i class="fas fa-sync" aria-hidden="true"></i>
            </button>
          </th>
        </tr>
      </thead>
      <tbody>
        {#each invites as inv (inv.email + inv.inviteEmail)}
          {#if inv.inviteEmail === appState.user.email}
            <tr>
              <td data-label="Invited By">{inv.email}</td>
              <td data-label="Status" class="text-center">
                {#if inv.confirmedAt}
                  <i class="fa-check"></i>
                {:else}
                  <i class="fa-clock"></i>
                {/if}
              </td>
              <td data-label="Date" title={formatDate(new Date(inv.createdAt))}>
                {timeSince(new Date(inv.createdAt), new Date())} ago
              </td>
              <td data-label="Action" class="text-center" colspan="2">
                <button
                        class="accessible-btn"
                        onclick={() => removeByInvite(inv.email, inv.inviteEmail)}
                        aria-label="Remove invitation"
                >
                  <i class="fas fa-trash" aria-hidden="true"></i>
                </button>

                {#if !inv.confirmedAt}
                  <button
                          class="accessible-btn"
                          onclick={() => acceptInvite(inv.email)}
                          aria-label="Accept invitation"
                  >
                    <i class="fas fa-check" aria-hidden="true"></i>
                  </button>
                {/if}
              </td>
            </tr>
          {/if}
        {/each}
      </tbody>
    </table>
  </div>
</Main>
