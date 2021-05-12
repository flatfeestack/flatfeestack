<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import { error, user } from "../ts/store";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { faUpload } from "@fortawesome/free-solid-svg-icons";

  let checked = $user.role != "ORG";
  let nameOrig = $user.name;
  let timeoutName;
  let userModeOrig = $user.role;
  let timeoutUserMode;
  let fileInput;

  $: $user.role = checked === false ? "ORG": "USR";

  $: {
    if(timeoutUserMode) {
      clearTimeout(timeoutUserMode);
    }
    timeoutUserMode = setTimeout(() => {
      if ($user.role !== userModeOrig) {
        API.user.setUserMode($user.role);
        userModeOrig = $user.role;
      }
    }, 1000)
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

  const onFileSelected = (e) => {
    let image = e.target.files[0];
    let reader = new FileReader();
    reader.readAsDataURL(image);
    reader.onload = e => {
      if (typeof reader.result !== "string") {
        $error = "not a string?";
        return;
      }
      const data: string = reader.result as string;
      if (data.length > 200 * 1024) {
        $error = "image too large, max is 200KB";
        return;
      }
      API.user.setImage(data);
      $user.image = data;
    };
  };
</script>

<style>
    .container {
        display: flex;
        flex-direction: row;
        align-items: center;
        margin: 1rem 1rem 3rem;
    }

    .upload {
        display: flex;
        cursor: pointer;
        align-items: center;
    }

    /* on-off button layout */
    .onoffswitch {
        position: relative;
        width: 7rem;
    }
    .onoffswitch-checkbox {
        position: absolute;
        opacity: 0;
        pointer-events: none;
    }
    .onoffswitch-label {
        display: block; overflow: hidden; cursor: pointer;
        border: 1px solid #999999; border-radius: 20px;
    }
    .onoffswitch-inner {
        display: block; width: 200%; margin-left: -100%;
        transition: margin 0.3s ease-in 0s;
    }
    .onoffswitch-inner:before, .onoffswitch-inner:after {
        display: block; float: left; width: 50%; height: 1.5rem; padding: 0; line-height: 1.5rem;
        font-size: 0.8rem; color: white; font-weight: bold;
        box-sizing: border-box;
    }
    .onoffswitch-inner:before {
        content: "Contributor";
        padding-left: 10px;
        background-color: #438A5E; color: #FFFFFF;
    }
    .onoffswitch-inner:after {
        content: "Organization";
        padding-right: 10px;
        background-color: #285338; color: #FFFFFF;
        text-align: right;
    }
    .onoffswitch-switch {
        display: block; width: 0.75rem; margin: 6px;
        background: #FFFFFF;
        position: absolute; top: 0; bottom: 0;
        right: 5.25rem;
        border: 2px solid #999999; border-radius: 20px;
        transition: all 0.3s ease-in 0s;
    }
    .onoffswitch-checkbox:checked + .onoffswitch-label .onoffswitch-inner {
        margin-left: 0;
    }
    .onoffswitch-checkbox:checked + .onoffswitch-label .onoffswitch-switch {
        right: 0;
    }
</style>

<Navigation>
  <h1 class="px-2">Settings</h1>

  <div class="container">
    {#if checked}
      <label class="px-2">Name: </label>
      <input type="text" bind:value={$user.name} placeholder="Name on the badge">
    {:else}
      <label class="px-2">Organization name: </label>
      <input type="text" bind:value={$user.name} placeholder="My organization name">
    {/if}
  </div>

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
    <label class="px-2">Upload your profile picture:</label>
    <div class="upload" on:click={()=>{fileInput.click();}}>
      <Fa icon="{faUpload}" size="lg" class="icon, px-2" />
      <input style="display:none" type="file" accept=".jpg, .jpeg, .png" on:change={(e)=>onFileSelected(e)}
             bind:this={fileInput}>
      {#if $user.image}
        {#if checked}
          <img class="image-usr" src="{$user.image}" />
        {:else}
          <img class="image-org" src="{$user.image}" />
        {/if}
      {/if}
    </div>
  </div>

</Navigation>
