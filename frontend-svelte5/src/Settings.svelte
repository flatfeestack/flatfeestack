<script lang="ts">
  import { onMount } from "svelte";
  import Navigation from "./Navigation.svelte";
  import { API } from "./ts/api.ts";
  import {appState} from "./ts/state.svelte.ts";
  import { formatDate, timeSince } from "./ts/services.svelte.ts";
  import type { GitUser } from "./types/backend";
  import { emailValidationPattern } from "./utils";

  let fileInput;
  let username: undefined | string;

  let gitEmails: GitUser[] = [];
  let newEmail = "";

  $: {
    if (typeof username === "undefined" && appState.user.name) {
      username = appState.user.name;
    }
  }

  const onFileSelected = (e:any) => {
    let image = e.target.files[0];
    let reader = new FileReader();
    reader.readAsDataURL(image);
    reader.onload = () => {
      if (typeof reader.result !== "string") {
        appState.setError("not a string?");
        return;
      }
      const data: string = reader.result as string;
      if (data.length > 200 * 1024) {
        appState.setError("image too large, max is 200KB");
        return;
      }
      API.user.setImage(data);
      appState.user.image = data;
    };
  };

  function handleUsernameChangeg() {
    try {
      if (username === "") {
        API.user.clearName();
      } else {
        API.user.setName(username);
        appState.user.name = username;
      }
    } catch (e) {
      appState.setError(e);
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
      appState.setError(e);
    }
  }

  async function removeEmail(email: string) {
    try {
      await API.user.removeGitEmail(email);
      gitEmails = gitEmails.filter((e) => e.email !== email);
    } catch (e) {
      appState.setError(e);
    }
  }

  const deleteImage = async () => {
    try {
      await API.user.deleteImage();
      appState.user.image = null;
    } catch (e) {
      appState.setError(e);
    }
  };

  onMount(async () => {
    try {
      const pr1 = API.user.gitEmails();
      const res1 = await pr1;
      gitEmails = res1 ? res1 : gitEmails;
    } catch (e) {
      appState.setError(e);
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

  .image-container {
    display: flex;
    align-items: center;
  }

  @media screen and (max-width: 600px) {
    .grid-2 {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
    }
    .grid-2 p,
    .grid-2 span,
    .grid-2 label,
    .grid-2 input {
      padding: 0;
      margin: 0;
    }

    .grid-2 input {
      width: 100%;
      padding: 0.25em;
    }

    .grid-2 span {
      margin: 10px 0;
    }

    .grid-2 label {
      margin: 15px 0;
    }
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
    table form {
      display: flex;
      flex-direction: column;
    }
    table form button {
      margin-top: 15px;
      margin-left: 0;
    }
  }
</style>

<Navigation>
  <h2 class="p-2 m-2">Account Settings</h2>
  <div class="grid-2">
    <p class="p-2 m-2 nobreak">Email:</p>
    <span class="p-2 m-2">{appState.user.email}</span>
    <label for="username-input" class="p-2 m-2 nobreak">Your name: </label>
    <input
      id="username-input"
      type="text"
      class="m-4 max-w20"
      bind:value={username}
      on:change={handleUsernameChangeg}
      placeholder="Name on the badge"
    />
    {#if username === "" || typeof username === "undefined"}
      <p class="p-rl-4 m-0 user-hint">
        If no username is set, your email address will be used for your public
        badges.
      </p>
    {/if}
    <label for="profile-picture-upload" class="p-2 m-2 nobreak"
      >Profile picture:</label
    >
    <div>
      {#if appState.user.image}
        <div class="image-container">
          <button class="upload accessible-btn" on:click={deleteImage}>
            <i class ="fa-trash-can"></i>
          </button>
          <img class="image-org" src={appState.user.image} alt="profile img" />
        </div>
      {:else}
        <button
          id="profile-picture-upload"
          class="upload accessible-btn"
          on:click={() => {
            fileInput.click();
          }}
        >
          <i class="fa-upload"></i>
          <input
            style="display:none"
            type="file"
            accept=".jpg, .jpeg, .png"
            on:change={(e) => onFileSelected(e)}
            bind:this={fileInput}
          />
        </button>
      {/if}
    </div>
  </div>

  <h2 class="p-2 ml-5 mb-0">Connect your Git Email to this Account</h2>
  <p class="p-2 m-2">
    If you have multiple git email addresses, you can connect these addresses to
    your FlatFeeStack account. You must verify your git email address. Once
    validated, the confirmed date will show the validation date. In case you
    didn't receive a confirmation email, please remove and re-add your git email
    address.
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
            <td data-label="Email">{email.email}</td>

            {#if email.confirmedAt}
              <td
                data-label="Confirmation"
                title={formatDate(new Date(email.confirmedAt))}
              >
                {timeSince(new Date(email.confirmedAt), new Date())} ago
              </td>
            {:else}
              <td data-label="Confirmation">
                <i class="fa-clock"></i>
              </td>
            {/if}
            <td data-label="Delete">
              <button
                class="accessible-btn"
                on:click={() => removeEmail(email.email)}
              >
                <i class="fa-trash"></i>
              </button>
            </td>
          </tr>
        {/each}
        <tr>
          <td colspan="3">
            <form class="p-2" on:submit|preventDefault={handleAddEmail}>
              <input
                id="email-input"
                name="email"
                type="email"
                pattern={emailValidationPattern}
                required
                bind:value={newEmail}
                placeholder="Email"
              />
              <button class="ml-5 p-2 button1" type="submit"
                >Add Git Email
              </button>
            </form>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</Navigation>
