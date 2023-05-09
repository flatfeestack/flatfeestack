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

  let fileInput;
  let username: undefined | string;

  let gitEmails: GitUser[] = [];
  let newEmail = "";

  $: {
    if (typeof username === "undefined" && $user.name) {
      username = $user.name;
    }
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

  function handleUsernameChangeg() {
    try {
      if (username === "") {
        API.user.clearName();
      } else {
        API.user.setName(username);
        $user.name = username;
      }
    } catch (e) {
      $error = e;
    }
  }

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
  .user-hint {
    grid-column: 2/2;
    font-size: 16px;
  }

  .settings-container {
    display: grid;
    grid-template-columns: min-content 1fr;
  }

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
  .col {
    display: flex;
    align-items: center;
    justify-content: left;
  }
  .col p {
    margin: 0;
  }
  .col-6 {
    width: 50%;
    min-width: 48%;
  }
  .col-3 {
    width: 30%;
    min-width: 26%;
  }
  .col-2 {
    width: 20%;
  }
  .wrapper {
    display: flex;
    width: 100%;
    border-bottom: solid white 1px;
  }
  .wrapper.row:nth-of-type(odd) {
    background-color: var(--primary-300);
  }
  .wrapper.row:nth-of-type(even) {
    background-color: var(--primary-100);
  }
  .b-r-w {
    border-right: solid white 1px;
  }
  .m-tb-2 {
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
  }
  @media screen and (max-width: 50em) {
    .settings-container {
      display: flex;
      flex-direction: column;
    }

    .wrapper {
      flex-direction: column;
      border-bottom: none;
    }

    .wrapper:last-child button {
      margin-top: 0.5rem;
    }

    .b-r-w {
      border-right: none;
    }
    .col-6,
    .col-3,
    .col-2 {
      width: unset;
    }
    .col {
      border-bottom: solid white 1px;
    }
    .m-tb-2 {
      margin-top: 0;
      margin-bottom: 0.25rem;
    }
    .image-container {
      display: flex;
      justify-content: center;
    }
  }
</style>

<Navigation>
  <h2 class="p-2 m-2">Account Settings</h2>
  <div class="settings-container m-2">
    <p class="p-2 m-0 m-tb-2 nobreak">Email:</p>
    <span class="p-2 m-0 m-tb-2">{$user.email}</span>

    <label for="username-input" class="p-2 m-tb-2 nobreak">Your name:</label>
    <input
      id="username-input"
      type="text"
      class="max-w20 m-2 m-tb-2"
      bind:value={username}
      on:change={handleUsernameChangeg}
      placeholder="Name on the badge"
    />
    {#if username === "" || typeof username === "undefined"}
      <p class="p-rl-2 m-0 user-hint">
        If no username is set, your email address will be used for your public
        badges.
      </p>
    {/if}

    <label for="profile-picture-upload" class="p-2 m-tb-2 nobreak"
      >Profile picture:</label
    >
    <div class="image-container m-tb-2">
      <button
        id="profile-picture-upload"
        class="upload accessible-btn"
        on:click={() => {
          fileInput.click();
        }}
      >
        <Fa icon={faUpload} size="lg" class="icon px-2" />
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

  <h2 class="p-2 m-2">Connect your Git Email to this Account</h2>
  <p class="p-2 m-2">
    If you have multiple git email addresses, you can connect these addresses to
    your FlatFeeStack account. You must verify your git email address. Once
    validated, the confirmed date will show the validation date.
  </p>

  <div class="custom-table p-2 m-2">
    <div class="wrapper header">
      <div class="p-2 col col-6 b-r-w">Email</div>
      <div class="p-2 col col-3 b-r-w">Confirmation</div>
      <div class="p-2 col col-2">Delete</div>
    </div>

    {#each gitEmails as email, key (email.email)}
      <div class="wrapper row">
        <div class="col-6 col p-2 b-r-w">
          <p>{email.email}</p>
        </div>

        {#if email.confirmedAt}
          <div
            class="col-3 col p-2 b-r-w"
            title={formatDate(new Date(email.confirmedAt))}
          >
            {timeSince(new Date(email.confirmedAt), new Date())} ago
          </div>
        {:else}
          <div class="col-3 col p-2 b-r-w"><Fa icon={faClock} size="md" /></div>
        {/if}
        <div class="col-2 col p-2">
          <button
            class="accessible-btn"
            on:click={() => removeEmail(email.email)}
          >
            <Fa icon={faTrash} size="md" />
          </button>
        </div>
      </div>
    {/each}
    <div class="wrapper row">
      <form class="p-2" on:submit|preventDefault={handleAddEmail}>
        <input
          id="email-input"
          name="email"
          type="email"
          required
          bind:value={newEmail}
          placeholder="Email"
        />
        <button class="p-2 button1" type="submit">Add Git Email</button>
      </form>
    </div>
  </div>
</Navigation>
