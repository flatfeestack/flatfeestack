<script>
import DashboardLayout from "./DashboardLayout.svelte";
import Payment from "../../components/Payment.svelte";
import { user } from "ts/auth";
import {faUpload} from "@fortawesome/free-solid-svg-icons";
import Fa from "svelte-fa";
import { API } from "ts/api.ts";
let checked = $user.mode != "ORG";

$: {
  if (checked === false) {
    $user.mode = "ORG";
  } else {
    $user.mode = "USR";
  }
}

let nameOrig = $user.name;
let timeoutName;

$: {
  if(timeoutName) {
    clearTimeout(timeoutName);
  }
  timeoutName = setTimeout(() => {
    if ($user.name !== nameOrig) {
      API.user.setName($user.name);
      nameOrig = $user.name;
    }
  }, 1000)
}
//onDestroy(()=> clearTimeout(timeout)) -> always store

let fileinput, error;
const onFileSelected =(e)=> {
  let image = e.target.files[0];
  let reader = new FileReader();
  reader.readAsDataURL(image);
  reader.onload = e => {
    if(e.target.result.length > 200 * 1024) {
      error = "image too large, max is 200KB";
      console.log(":::::::::::::::::::::::::::")
      return;
    }
    API.user.setImage(e.target.result)
    $user.image = e.target.result
  };
}

</script>

<style>
    .container {
        display: flex;
        flex-direction: row;
        margin: 1em;
    }
    .wrap {
        display: flex;
        flex-wrap: wrap;
    }
    .upload{
        display:flex;
        height:50px;
        width:50px;
        cursor:pointer;
    }
    .image-usr {
        height:10em;
        width:10em;
        border-radius: 50%;
        object-fit: cover;
    }
    .image-org {
        display:flex;
        max-height:10em;
        max-width:10em;
        width:auto;
    }
</style>

{#if error}
  <div class="bg-red-500 text-white p-3 my-5">{error}</div>
{/if}
<DashboardLayout>
  <h1 class = "px-2">Profile</h1>

  <div class = "container px-2">
    <label class = "">Are you an organization or an individual contributor? </label>
    <div class="onoffswitch">
    <input type="checkbox" bind:checked={checked} name="onoffswitch" class="onoffswitch-checkbox" id="myonoffswitch" tabindex="0" >
    <label class="onoffswitch-label" for="myonoffswitch">
      <span class="onoffswitch-inner"></span>
      <span class="onoffswitch-switch"></span>
    </label>
  </div>
  </div>


  <div class = "container px-2">
    {#if checked}
    <label class = "">What name should appear on your badge? </label>
    {:else}
      <label class = "">What is the name of your organization? </label>
    {/if}
    <input type="text" bind:value={$user.name}>

    Upload your profile picture

    {#if $user.image}
      {#if checked}
        <img class="image-usr" src="{$user.image}" alt="d" />
      {:else}
        <img class="image-org" src="{$user.image}" alt="d" />
      {/if}
    {:else}
      <img class="image" src="https://cdn4.iconfinder.com/data/icons/small-n-flat/24/user-alt-512.png" alt="" />
    {/if}

    <span class="upload" on:click={()=>{fileinput.click();}}>
      <Fa icon="{faUpload}" size="lg"  class="icon" />
    </span>

    <div class="chan" on:click={()=>{fileinput.click();}}>Choose Image</div>
    <input style="display:none" type="file" accept=".jpg, .jpeg, .png" on:change={(e)=>onFileSelected(e)} bind:this={fileinput} >

  </div>

  <div class="flex">
      <h2 class="Sponsoring">
        {#if $user.subscription_state}
          You are currently sponsoring 5 projects
          <span>{$user.subscription_state}</span>
        {:else}
          Support these awesome 5 projects
        {/if}
      </h2>

      {#if $user.subscription_state !== 'ACTIVE'}
        <div class="container bg-green rounded p-2 my-4">
          You don't have an active subscription.
          <br />Please choose a plan below to start sponsoring open source
          projects
        </div>
      {/if}
      <Payment />
    </div>


</DashboardLayout>
