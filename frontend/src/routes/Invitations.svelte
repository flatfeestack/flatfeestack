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

      const res1 = API.invite.inviteAuth(inviteEmail);
      const res2 = API.invite.invite(inviteEmail);
      const inv: Invitation = {
        email: $user.email,
        inviteEmail,
        createdAt: new Date().toISOString(),
        confirmedAt: null,
      };
      invites = [...invites, inv];
      await res1;
      await res2;
    } catch (e) {
      $error =
        "Duplicate email address. Can't invite the same address multiple times.";
    } finally {
      isAddInviteSubmitting = false;
    }
  }

  onMount(async () => {
    const pr1 = refreshInvite();
    const pr2 = API.user.statusSponsoredUsers();
    const res2 = await pr2;
    statusSponsoredUsers = res2 === null ? [] : res2;
    await pr1;
  });
</script>

<style>
  .custom-table {
    display: flex;
    flex-wrap: wrap;
  }
  .header {
    background-color: var(--primary-300);
    color: #000;
    text-align: left;
    font-weight: bold;
  }
  .wrapper:not(.header):nth-of-type(odd) {
    background-color: var(--primary-300);
  }
  .wrapper:not(.header):nth-of-type(even) {
    background-color: var(--primary-100);
  }
  .wrapper {
    display: flex;
    width: 100%;
    border-bottom: solid white 1px;
  }
  .wrapper > div {
    display: flex;
    align-items: center;
    justify-content: left;
  }
  .col-6 {
    width: 50%;
  }
  .col-2 {
    width: 20%;
  }

  .col-1 {
    width: 10%;
  }

  .break-word {
    word-break: break-all;
  }

  div.just-end {
    justify-content: end;
  }

  .form-container {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
  }

  .form-container {
    margin: 0.5rem;
  }

  .wrapper .accessible-btn + .accessible-btn {
    margin-left: 0.5rem;
  }

  @media screen and (max-width: 54em) {
    .wrapper {
      flex-direction: row;
      flex-wrap: wrap;
    }
    .col-6 {
      width: 100%;
    }
    .col-1 {
      width: 15%;
    }
    .col-2 {
      width: 35%;
    }
  }
</style>

<Navigation>
  <h2 class="p-2 m-2">Invite Users</h2>
  <p class="p-2 m-2">
    Invite your friends or co-workers. They will be charged from your account on
    a daily basis.
  </p>

  <div class="custom-table p-2 m-2">
    <div class="wrapper header">
      <div class="p-2 col-6">Invited</div>
      <div class="p-2 col-1">Status</div>
      <div class="p-2 col-2">Date</div>
      <div class="p-2 col-1">Remove</div>
      <div class="p-2 col-1 just-end">
        <button class="accessible-btn" on:click={refreshInvite}>
          <Fa icon={faSync} size="md" />
        </button>
      </div>
    </div>
    {#each invites as inv, key (inv.email + inv.inviteEmail)}
      {#if inv.email === $user.email}
        <div class="wrapper">
          <div class="col-6 p-2 break-word">{inv.inviteEmail}</div>
          <div class="col-1 p-2">
            {#if inv.confirmedAt}
              <Fa icon={faCheck} size="md" />
            {:else}
              <Fa icon={faClock} size="md" />
            {/if}
          </div>
          <div title={formatDate(new Date(inv.createdAt))} class="col-2 p-2">
            {timeSince(new Date(inv.createdAt), new Date())} ago
          </div>
          <div class="col-2 p-2">
            <button
              class="accessible-btn"
              on:click={() => removeMyInvite(inv.email, inv.inviteEmail)}
              ><Fa icon={faTrash} size="md" /></button
            >
          </div>
        </div>
      {/if}
    {/each}
    <div class="wrapper">
      <form on:submit|preventDefault={addInvite} class="form-container">
        <label for="invite-mail-input" class="p-2">Invite by email:</label>
        <input
          id="invite-mail-input"
          class="m-2"
          size="24"
          maxlength="50"
          type="email"
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
    </div>
  </div>
  <h2 class="p-2 m-2">Invited By</h2>
  <p class="p-2 m-2">Accept your invitation and fund your account.</p>

  <div class="custom-table p-2 m-2">
    <div class="wrapper header">
      <div class="p-2 col-6 b-r-w">Invited By</div>
      <div class="p-2 col-1 b-r-w">Status</div>
      <div class="p-2 col-2 b-r-w">Date</div>
      <div class="p-2 col-1 b-r-w">Action</div>
      <div class="p-2 col-1 just-end">
        <button class="accessible-btn" on:click={refreshInvite}
          ><Fa icon={faSync} size="md" /></button
        >
      </div>
    </div>
    {#each invites as inv, key (inv.email + inv.inviteEmail)}
      {#if inv.inviteEmail === $user.email}
        <div class="wrapper">
          <div class="col-6 p-2 break-word">{inv.email}</div>
          <div class="col-1 p-2">
            {#if inv.confirmedAt}
              <Fa icon={faCheck} size="md" />
            {:else}
              <Fa icon={faClock} size="md" />
            {/if}
          </div>
          <div class="col-2 p-2" title={formatDate(new Date(inv.createdAt))}>
            {timeSince(new Date(inv.createdAt), new Date())} ago
          </div>
          <div class="col-2 p-2">
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
          </div>
        </div>
      {/if}
    {/each}
  </div>
</Navigation>
