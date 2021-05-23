<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import Fa from "svelte-fa";
  import { onMount } from "svelte";
  import { API } from "../ts/api";
  import { faTrash, faClock } from "@fortawesome/free-solid-svg-icons";
  import Web3 from "../components/Web3.svelte";
  import { error, user } from "../ts/store";
  import type { GitUser } from "../types/users.ts";
  import { formatDate } from "../ts/services";

  let address = "";
  let gitEmails: GitUser[] = [];
  let newEmail = "";
  let isSubmitting = false;

  async function updatePayout(e) {
    try {
      if (!$user.payout_eth || !$user.payout_eth.match(/^0x[a-fA-F0-9]{40}$/g)) {
        $error = "Invalid ethereum address";
      }
      //TODO: no button, wait 1sec
      await API.user.updatePayoutAddress($user.payout_eth);
    } catch (e) {
      $error = e;
    }
  }

  async function handleSubmit() {
    try {
      await API.user.addEmail(newEmail);
      let ge: GitUser = {
        confirmedAt: null, createdAt: null, email: newEmail
      };
      gitEmails = [...gitEmails, ge];
      newEmail = "";
    } catch (e) {
      $error = e;
    }
  }

  async function removeEmail(email: string) {
    try {
      await API.user.removeGitEmail(email);
      gitEmails = gitEmails.filter((e) => e.email !== email);
    } catch (e) {
      $error = e;
    }
  }

  onMount(async () => {
    try {
      const res = await API.user.gitEmails();
      gitEmails = res ? res:gitEmails;
    } catch (e) {
      $error = e;
    }
  });

</script>

<Navigation>
  <h1 class="px-2">Income</h1>

  {#if !gitEmails || gitEmails.length === 0}
    <div class="container bg-green rounded p-2 m-2">
      Please add your git e-mail addresses to generate income
    </div>
  {/if}

  <div class="container">
    <label class="px-2">Add Git Email:</label>
    <input id="email-input" name="email" type="text" bind:value={newEmail} placeholder="Email" />
    <form class="p-2" on:submit|preventDefault="{handleSubmit}">
      <button type="submit">Add Email</button>
    </form>
  </div>

  {#if gitEmails && gitEmails.length > 0}

    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Email</th>
          <th>Confirm Date</th>
          <th>Delete</th>
        </tr>
        </thead>
        <tbody>
        {#each gitEmails as email, key (email.email)}
          <tr>
            <td>{email.email}</td>
            <td>
              {#if email.confirmedAt}
                {formatDate(new Date(email.confirmedAt))}
              {:else }
                <Fa icon="{faClock}" size="md" />
              {/if}
            </td>
            <td class="cursor-pointer" on:click="{() => removeEmail(email.email)}">
              <Fa icon="{faTrash}" size="md" />
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="3">No Data</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <div class="container">
    <label class="px-2">Payout Address:</label>
    <input type="text" bind:value="{$user.payout_eth}" placeholder="Ethereum Address" />
    <form class="p-2" on:submit|preventDefault="{updatePayout}">
      <button type="submit">Update Address</button>
    </form>
  </div>

  <Web3 />
</Navigation>
