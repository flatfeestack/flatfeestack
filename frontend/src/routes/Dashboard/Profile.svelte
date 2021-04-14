<script lang="ts">
  import DashboardLayout from "./DashboardLayout.svelte";
  import Payment from "../../components/Payment.svelte";
  import { user } from "ts/auth";
  import Fa from "svelte-fa";
  import { API } from "ts/api.ts";
  import { onMount } from "svelte";
  import type { Invitation } from "types/invitation.type";
  import { faTrash, faUpload } from "@fortawesome/free-solid-svg-icons";
  import { Repo } from "../../types/repo.type";
  import { links } from "svelte-routing";

  let checked = $user.role != "ORG";
  let nameOrig = $user.name;
  let timeoutName;
  let invites: Invitation[] = [];
  let sponsoredRepos: Repo[] = [];
  let invite_email;

  $: {
    if (checked === false) {
      $user.role = "ORG";
    } else {
      $user.role = "USR";
    }
  }

  $: {
    if (timeoutName) {
      clearTimeout(timeoutName);
    }
    timeoutName = setTimeout(() => {
      if ($user.name !== nameOrig) {
        API.user.setName($user.name);
        nameOrig = $user.name;
      }
    }, 1000);
  }

  let fileinput, error;
  const onFileSelected = (e) => {
    let image = e.target.files[0];
    let reader = new FileReader();
    reader.readAsDataURL(image);
    reader.onload = e => {
      if (typeof reader.result !== "string") {
        console.log("not a string?");
        return;
      }
      const data: string = reader.result as string;
      if (data.length > 200 * 1024) {
        error = "image too large, max is 200KB";
        console.log(":::::::::::::::::::::::::::");
        return;
      }
      API.user.setImage(data);
      $user.image = data;
    };
  };

  async function removeInvite(email: string) {
    try {
      await API.authToken.delInvite(email);
      const response = await API.authToken.invites();
      if (response?.data && response.data.length > 0) {
        invites = response.data;
      }
      console.log(invites);
    } catch (e) {
      console.log(e);
    }
  }

  async function invite() {
    try {
      const d = new Date();
      await API.authToken.invite(invite_email, $user.email, $user.name, new Date().toISOString());
      const response = await API.authToken.invites();
      if (response?.data && response.data.length > 0) {
        invites = response.data;
      }
      console.log(invites);
    } catch (e) {
      console.log(e);
    }
  }

  //onDestroy(()=> clearTimeout(timeout)) -> always store
  onMount(async () => {
    try {
      const res1 = await API.authToken.invites();
      const res2 = await API.user.getSponsored();
      invites = res1.data === null ? [] : res1.data;
      sponsoredRepos = res2.data === null ? [] : res2.data;

    } catch (e) {
      error = e;
      console.log(e);
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

    .upload {
        display: flex;
        cursor: pointer;
    }

    .image-usr {
        height: 10em;
        width: 10em;
        border-radius: 50%;
        object-fit: cover;
    }

    .image-org {
        display: flex;
        max-height: 10em;
        max-width: 10em;
        width: auto;
    }
</style>

{#if error}
  <div class="bg-red-500 text-white p-3 my-5">{error}</div>
{/if}
<DashboardLayout>
  <h1 class="px-2">Profile</h1>

  <div class="container">
    <label class="px-2">Email:&nbsp;</label>
    <input type="email" disabled="true" value="{$user.email}">
  </div>

  <div class="container">
    <label class="px-2">Are you an organization or an individual contributor?&nbsp;</label>
    <div class="onoffswitch">
      <input type="checkbox" bind:checked={checked} name="onoffswitch" class="onoffswitch-checkbox" id="myonoffswitch"
             tabindex="0">
      <label class="onoffswitch-label" for="myonoffswitch">
        <span class="onoffswitch-inner"></span>
        <span class="onoffswitch-switch"></span>
      </label>
    </div>
  </div>

  <div class="container">
    {#if checked}
      <label class="px-2">User name: </label>
      <input type="text" bind:value={$user.name} placeholder="Name on the badge">
    {:else}
      <label class="px-2">Organization name: </label>
      <input type="text" bind:value={$user.name} placeholder="My organization name">
    {/if}
  </div>

  <div class="container">
    <label class="px-2">Upload your profile picture:</label>
    <div class="upload" on:click={()=>{fileinput.click();}}>
      <Fa icon="{faUpload}" size="lg" class="icon, px-2" />
      <input style="display:none" type="file" accept=".jpg, .jpeg, .png" on:change={(e)=>onFileSelected(e)}
             bind:this={fileinput}>
      {#if $user.image}
        {#if checked}
          <img class="image-usr" src="{$user.image}" />
        {:else}
          <img class="image-org" src="{$user.image}" />
        {/if}
      {/if}
    </div>
  </div>

  {#if checked}
    <div class="container">
      {#if sponsoredRepos.length > 0}
        <h2 class="Sponsoring">
          Support {sponsoredRepos.length} projects
        </h2>
      {:else}
        <div class="bg-green rounded p-2 my-4" use:links>
          <p>You are not supporting any projects yet. Please go to the <a href="/dashboard/sponsoring">Find Repos</a>
            section
            where you can add your favorite projects.</p>
        </div>
      {/if}
    </div>
  {/if}

  {#if sponsoredRepos.length > 0 || !checked}
    <h2 class="px-2">Donation</h2>
    <Payment />
  {/if}


  {#if !checked}
    <div class="container">
      {#each invites as inv}
        {inv.email}
        {inv.pending}
        {inv.createdAt}
        <div class="cursor-pointer transform hover:scale-105 duration-200" on:click="{() => removeInvite(inv.email)}">
          <Fa icon="{faTrash}" size="md" />
        </div>
      {/each}
    </div>

    <form on:submit|preventDefault="{invite}" class="container">
      <label class="p-2">Invite Email</label>
      <input class="p-2" size="24" maxlength="100" type="email" bind:value="{invite_email}" />&nbsp;
      <button type="submit" class="button p-2">Invite to your org</button>
    </form>

  {/if}

</DashboardLayout>
