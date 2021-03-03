
<script lang="ts">
import DashboardLayout from "./DashboardLayout.svelte";
import Fa from "svelte-fa";
import { onMount } from "svelte";
import { API } from "ts/api";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import Web3 from "../../components/Web3.svelte";
import { user } from "ts/auth";
import { GitUser } from "../../types/git-email.type";

let address = "";
let gitEmails: Array<GitUser> = [];
let newEmail = "";
onMount(async () => {
  try {
    const response = await API.user.gitEmails();
    if (response?.data && response.data.length > 0) {
      console.log(response.data);
      gitEmails = response.data;
    }
  } catch (e) {
    console.log(e);
  }
});

let error = "";

async function updatePayout(e) {
  try {
    if (!$user.payout_eth || !$user.payout_eth.match(/^0x[a-fA-F0-9]{40}$/g)) {
      console.log("error not valid eth address");
      error = "Invalid ethereum address";
    }
    //TODO: no button, wait 1sec
    await API.user.updatePayoutAddress($user.payout_eth);
    error = "";
  } catch (e) {
    error = String(e);
    console.log(e);
  }
}

async function handleSubmit() {
  try {
    error = "";
    await API.user.addEmail(newEmail);
    let ge: GitUser = {
      confirmedAt: "", createdAt: "", email: newEmail
    }
    gitEmails = [...gitEmails, ge];
    newEmail = "";
  } catch (e) {
    error = e.response?.data?.message || "Something went wrong:" + e;
  }
}

let isSubmitting = false;

async function removeEmail(email: string) {
  try {
    error = "";
    console.log("remove email:", email);
    await API.user.removeGitEmail(email);
    gitEmails = gitEmails.filter((e) => e.email !== email);
  } catch (e) {
    console.log("in catch", e.message);
    error = e?.message || "Something went wrong";
  }
}
</script>

<style>

</style>


<DashboardLayout>
  <h1>Income</h1>
  <hr class="mb-10 w-64" />
  {#if gitEmails && gitEmails.length === 0}
    <div class="flex mb-5">
      <div class="bg-red-500 text-white p-5">
        Please add your git e-mail addresses to generate income
      </div>
    </div>
  {/if}
  <h2 class="mb-5">Connected Git Emails</h2>
  {#each gitEmails as email}
    <div class="div mb-2 flex flex-row items-center">
      <div class="w-64">
        <input type="text" class="input" value="{email.email}" disabled />
      </div>
      <input type="text" class="input" value="{email.confirmedAt}" disabled />
      <div
        class="ml-5 hover:text-red-500 cursor-pointer transform hover:scale-105 duration-200"
        on:click="{() => removeEmail(email.email)}"
      >
        <Fa icon="{faTrash}" size="md" />
      </div>
    </div>
  {/each}
  <h2 class="my-5">Add Git Email</h2>
  <form on:submit|preventDefault="{handleSubmit}">
    <div class="flex flex-row items-end">
      <div class="w-64">
        <label for="email-input" class="block text-grey-darker text-sm font-bold mb-2 w-full">Email</label>
        <input id="email-input" name="email" type="text" class="input" bind:value={newEmail}/>
      </div>
      <div class="ml-5">
        <button
          class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75"
          disabled="{isSubmitting}"
          type="submit"
        >Add Email{#if isSubmitting}...{/if}</button>
      </div>
      {#if error}
        <div class="bg-red-500 p-2 text-white mt-2">{error}</div>
      {/if}
    </div>
  </form>
  <h2 class="my-5">Payout Address</h2>

  <form on:submit|preventDefault="{updatePayout}">
    <div class="w-64">
      <label class="block text-grey-darker text-sm font-bold mb-2 w-full">Ethereum
        Address</label>
      <input type="text" class="input" bind:value="{$user.payout_eth}" />
    </div>
    <div><button type="submit" class="button ml-5">Update Address</button></div>
  </form>
  {#if error}
    <div class="flex mt-5">
      <p class="bg-red-500 p-2 block text-white rounded">{error}</p>
    </div>
  {/if}

  <h2 class="my-5">Request Funds</h2>

  <Web3 />
</DashboardLayout>
