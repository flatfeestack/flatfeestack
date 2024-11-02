<script lang="ts">
  import {
    faClock,
    faTrash,
    faUpload,
    faTrashCan,
  } from "@fortawesome/free-solid-svg-icons";
  import { onMount } from "svelte";
  import Fa from "svelte-fa";
  import Navigation from "../components/Navigation.svelte";
  import { API } from "../ts/api";
  import { error, user } from "../ts/mainStore";
  import { formatDate, timeSince } from "../ts/services";
  import type { GitUser } from "../types/backend";
  import { emailValidationPattern } from "../ts/utils";
  import { fade } from 'svelte/transition';

  let fileInput;
  let username: undefined | string;

  let gitEmails: GitUser[] = [];
  let newEmail = "";

  let multiplierActive = false; // add this to a store when using it for other pages
  let showMultiplierInfo = false;
  let dailyLimit = 120; // should refer to backend
  let newDailyLimit;

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

  const deleteImage = async () => {
    try {
      await API.user.deleteImage();
      $user.image = null;
    } catch (e) {
      $error = e.message;
    }
  };

  function handleMultiplierToggle() {
    if (multiplierActive) {
      console.log("Switch is ON");
      // call backend
    } else {
      console.log("Switch is OFF");
      // call backend
    }
  }

  function toggleMultiplierInfoVisibility() {
    showMultiplierInfo = !showMultiplierInfo;
  }

  function setDailyLimit () {
    console.log(dailyLimit);
    dailyLimit = newDailyLimit // call backend
  }

  function handleLimitKeyDown(event) {
    if (event.key === "Enter") {
      setDailyLimit();
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

  .image-container {
    display: flex;
    align-items: center;
  }

  label.switch {
      position: relative;
      display: inline-block;
      flex-shrink: 0;
      width: 60px;
      height: 34px;
      margin: 1rem 1rem 1rem 0;
  }
  label.switch input {
      opacity: 0;
      width: 0;
      height: 0;
  }
  .slider {
      position: absolute;
      cursor: pointer;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background-color: #ccc;
      -webkit-transition: .4s;
      transition: .4s;
  }

  .slider:before {
      position: absolute;
      content: "";
      height: 26px;
      width: 26px;
      left: 4px;
      bottom: 4px;
      background-color: white;
      -webkit-transition: .4s;
      transition: .4s;
  }

  input:checked + .slider {
      background-color: var(--primary-500);
  }

  input:focus + .slider {
      box-shadow: 0 0 1px var(--primary-500);
  }

  input:checked + .slider:before {
      -ms-transform: translateX(26px);
      transform: translateX(26px);
  }
  .slider.round {
      border-radius: 34px;
  }

  .slider.round:before {
      border-radius: 50%;
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
    <span class="p-2 m-2">{$user.email}</span>
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
      {#if $user.image}
        <div class="image-container">
          <button class="upload accessible-btn" on:click={deleteImage}>
            <Fa icon={faTrashCan} size="lg" class="icon, px-2" />
          </button>
          <img class="image-org" src={$user.image} alt="profile img" />
        </div>
      {:else}
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
                <Fa icon={faClock} size="md" />
              </td>
            {/if}
            <td data-label="Delete">
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
  <div class="container-col m-4">
    <div class="container-small">
      <label class="switch">
        <input type="checkbox" bind:checked={multiplierActive} on:change={handleMultiplierToggle}>
        <span class="slider round"></span>
      </label>
      <div class="container-small">
        <p class="m-0"><strong>enable multiplier options</strong> <br></p>
        <button class="button1 ml-5" on:click={toggleMultiplierInfoVisibility}>?</button>
      </div>
    </div>
    {#if showMultiplierInfo}
      <p class="m-0" transition:fade={{ duration: 250 }}>
        What are multiplier options? Multiplier options allow you to boost your support for your favorite projects.
        When enabled, a special icon (ICON HERE) appears in the search tab. <br>
        By clicking the multiplier icon next to a repository, you activate a multiplier to support it.
        This means that each time another FlatFeeStack user donates to that repository, you'll automatically contribute up to 0.9% of their initial donation as well. <br>
        If you're interested in the exact mechanism of this feature, you can read it <strong>here</strong>.
      </p>
    {/if}
  </div>
  {#if multiplierActive}
    <div class="container-col" id="tipping-limit-div">
      <p>Your tipping limit is set to <strong>${ dailyLimit }</strong> per day.</p>
      <div class="container">
        <label for="daily-limit-input">Daily Limit </label>
        <input
          id="daily-limit-input"
          type="number"
          class="m-4 max-w20"
          bind:value={newDailyLimit}
          on:keydown={handleLimitKeyDown}
          placeholder="$"
        />
        <button on:click={setDailyLimit} class="ml-5 p-2 button1">Set Daily Limit</button>
      </div>
    </div>
  {/if}
</Navigation>
