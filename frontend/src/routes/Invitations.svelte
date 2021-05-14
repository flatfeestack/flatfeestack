<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error, isSubmitting, user } from "../ts/store";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { onMount } from "svelte";
  import type { Invitation } from "src/types/users.ts";
  import { faTrash, faSync, faClock, faCheck } from "@fortawesome/free-solid-svg-icons";
  import { formatDate, timeSince } from "../ts/services";
  import Dots from "../components/Dots.svelte";
  import { plans } from "../types/contract";

  let invites: Invitation[] = [];
  let inviteEmail;
  let isAddInviteSubmitting = false;
  let selected;

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
    $isSubmitting = true;
    try {
      const res1 = await API.authToken.invites();
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
      const d = new Date();
      d.setTime(d.getTime() + (1000 * 60 * 60 * 24 * 7));
      await API.authToken.invite(inviteEmail, $user.name, d.toISOString(), selected);
      const inv: Invitation = { email: inviteEmail, meta: selected, createdAt: new Date().toISOString(), confirmedAt: null };
      invites = [...invites, inv];
    } catch (e) {
      $error = e;
    } finally {
      isAddInviteSubmitting = false;
    }
  }

  onMount(async () => {

    await refreshInvite();
  });

</script>

<Navigation>
  <h1 class="px-2">Invitations</h1>

  <div class="container bg-green rounded p-2 m-2">
    <div>
      <p>Invite your co-workers to your organization. If they accept, the co-worker will have a prefunded account from
        your organization.</p>
    </div>
  </div>

  <h2 class="px-2">Invite users to {$user.name ? $user.name : "your org"}</h2>
  <form on:submit|preventDefault="{addInvite}" class="container">
    <label class="p-2">Invite this email:</label>
    <input size="24" maxlength="100" type="email" bind:value="{inviteEmail}" />&nbsp;
    <select bind:value={selected}>
      {#each plans as plan, i}
        <option value="{plan.freq}">{plan.title}</option>
      {/each}
    </select>
    <button type="submit" disabled="{isAddInviteSubmitting}">Invite to {$user.name ? $user.name : "your org"}
      {#if isAddInviteSubmitting}
        <Dots />
      {/if}
    </button>
  </form>

  <div class="container">
    <table>
      <thead>
      <tr>
        <th>Email</th>
        <th>Status</th>
        <th>Date</th>
        <th>Plan</th>
        <th>Remove</th>
        <th><span class="cursor-pointer" on:click="{refreshInvite}"><Fa icon="{faSync}" size="md" /></span></th>
      </tr>
      </thead>
      <tbody>
      {#each invites as inv, key (inv.email)}
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
          <td>{plans.find(plan => plan.freq == inv.meta).title}</td>
          <td class="text-center">
            <span class="cursor-pointer" on:click="{() => removeInvite(inv.email)}"><Fa icon="{faTrash}" size="md" /></span>
          </td>
          <td />
        </tr>
      {/each}
      </tbody>
    </table>
  </div>

</Navigation>
