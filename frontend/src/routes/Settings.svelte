<script lang="ts">
  import {
    faClock,
    faTrash,
    faUpload,
  } from "@fortawesome/free-solid-svg-icons";
  import { onMount } from "svelte";
  import Fa from "svelte-fa";
  import Navigation from "../components/Navigation.svelte";
  import { API } from "../ts/api";
  import { error, user } from "../ts/mainStore";
  import { formatDate, timeSince } from "../ts/services";
  import type { GitUser } from "../types/backend";

  let nameOrig = $user.name;
  let timeoutName;
  let fileInput;

  let gitEmails: GitUser[] = [];
  let newEmail = "";

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
    reader.onload = (e) => {
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

  async function handleAddEmail() {
    try {
      await API.user.addEmail(newEmail);
      let ge: GitUser = {
        confirmedAt: null,
        createdAt: null,
        email: newEmail,
      };
      gitEmails = [...gitEmails, ge];
      newEmail = "";
    } catch (e) {
      $error = "Duplicate email address. Email can only be used once.";
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
      const pr1 = API.user.gitEmails();
      const res1 = await pr1;
      gitEmails = res1 ? res1 : gitEmails;
    } catch (e) {
      $error = e;
    }
  });
</script>

<style>
  .upload {
    display: flex;
    cursor: pointer;
    align-items: center;
  }
</style>

<Navigation>
  <h2 class="p-2 m-2">Account Settings</h2>
  <div class="grid-2">
    <p class="p-2 m-2 nobreak">Email:&nbsp;</p>
    <span class="p-2 m-2">{$user.email}</span>
    <label for="username-input" class="p-2 m-2 nobreak">Your name: </label>
    <input
      id="username-input"
      type="text"
      class="max-w20"
      bind:value={$user.name}
      placeholder="Name on the badge"
    />
    <label for="profile-picture-upload" class="p-2 m-2 nobreak"
      >Profile picture:</label
    >
    <div>
      <button
        id="profile-picture-upload"
        class="upload accessible-btn"
        on:click={() => {
          fileInput.click();
        }}
      >
        <Fa icon={faUpload} size="lg" class="icon, px-2" />
        <input
          style="display:none"
          type="file"
          accept=".jpg, .jpeg, .png"
          on:change={(e) => onFileSelected(e)}
          bind:this={fileInput}
        />
        {#if $user.image}
          <img class="image-org" src={$user.image} alt="profile img" />
        {/if}
      </button>
    </div>
  </div>

  <h2 class="p-2 mt-40 ml-5 mb-0">Connect your Git Email to this Account</h2>
  <p class="p-2 m-2">
    If you have multiple git email addresses, you can connect these addresses to
    your FlatFeeStack account. You must verify your git email address. Once it
    has been validated, the confirm date will show the date of validation.
  </p>

  <div class="min-w20 container">
    <table>
      <thead>
        <tr>
          <th>Email</th>
          <th>Confirmation</th>
          <th>Delete</th>
        </tr>
      </thead>
      <tbody>
        {#each gitEmails as email, key (email.email)}
          <tr>
            <td>{email.email}</td>

            {#if email.confirmedAt}
              <td title={formatDate(new Date(email.confirmedAt))}>
                {timeSince(new Date(email.confirmedAt), new Date())} ago
              </td>
            {:else}
              <td><Fa icon={faClock} size="md" /></td>
            {/if}
            <td>
              <button
                class="accessible-btn"
                on:click={() => removeEmail(email.email)}
              >
                <Fa icon={faTrash} size="md" />
              </button>
            </td>
          </tr>
        {/each}
        <tr>
          <td colspan="3">
            <div class="container-small">
              <form class="p-2" on:submit|preventDefault={handleAddEmail}>
                <input
                  id="email-input"
                  name="email"
                  type="email"
                  required
                  bind:value={newEmail}
                  placeholder="Email"
                />
                <button class="ml-5 p-2 button1" type="submit"
                  >Add Git Email</button
                >
              </form>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</Navigation>
