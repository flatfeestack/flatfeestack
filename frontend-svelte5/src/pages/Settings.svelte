<script lang="ts">
  import {onMount} from "svelte";
  import { API } from "../ts/api.ts";
  import {appState} from "../ts/state.svelte.ts";
  import { formatDate, timeSince } from "../ts/services.svelte.ts";
  import type { GitUser } from "../types/backend.ts";
  import skaler, {debounce, emailValidationPattern} from "../utils.ts";
  import Main from "../Main.svelte";

  let username = $state("");
  let fileInput= $state<HTMLInputElement>();
  let gitEmails = $state<GitUser[]>([]);
  let newEmail = $state("");
  let isVisible = $state(false);
  let multiplierActive = $state(false);

  async function onFileSelected(e: Event) {
    try {
      const input = e.target as HTMLInputElement;
      if (!input.files || input.files.length === 0) {
        appState.setError('No file selected');
        return;
      }
      let image = input.files[0];

      const scaledFile = await skaler(image, {width: 480, quality:0.5});

      const reader = new FileReader();
      const base64 = await new Promise<string>((resolve, reject) => {
        reader.onload = () => {
          if (typeof reader.result !== 'string') {
            reject(new Error('Failed to convert to base64'));
            return;
          }
          resolve(reader.result);
        };
        reader.onerror = () => reject(new Error('Failed to read file'));
        reader.readAsDataURL(scaledFile);
      });
      console.log(base64.length);
      API.user.setImage(base64);
      appState.user.image = base64;
    } catch (e) {
      appState.setError(e);
    }
  }

  const handleUsernameChange = debounce(async () => {
    try {
      if (username !== "") {
        await API.user.setName(username);
        appState.user.name = username;
        isVisible = true;
        setTimeout(() => {
          isVisible = false;
        }, 3000);
      }
    } catch (e) {
      appState.setError(e);
    }
  }, 500);

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

  function handleMultiplierToggle() {
    try {
      if (multiplierActive) {
        API.user.setMultiplier(true);
      } else {
        API.user.setMultiplier(false);
      }
      appState.user.multiplier = multiplierActive;
    } catch (e) {
      appState.setError(e);
    }
  }

  onMount(async () => {
    try {
      const pr1 = API.user.gitEmails();
      const res1 = await pr1;
      gitEmails = res1 ? res1 : gitEmails;
    } catch (e) {
      appState.setError(e);
    }
  });

  $effect(() => {
    if (appState.user.name !== "undefined") {
      username = appState.user.name;
    }

    if (appState.user.multiplier) {
      multiplierActive = appState.user.multiplier;
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

  .success-checkmark {
    color: var(--primary-500);
    margin-left: 0.5rem;
    visibility: hidden;
    opacity: 0;
    transition: visibility 0s linear 300ms, opacity 300ms;
  }

  .success-checkmark.visible {
    visibility: visible;
    opacity: 1;
    transition: visibility 0s linear 0s, opacity 300ms;
  }

  .mul{
    width: 1.3rem;
  }
</style>

<Main>
  <h2>Account Settings</h2>
  <p class="p-2 m-2">
    Your email address is permanent and cannot be modified after registration.
    If you leave the name field empty, we'll automatically use the portion of
    your email address that appears before the "@" symbol as your display name.
  </p>

  <div class="grid-2">
    <span class="p-050 nobreak">Email:</span>
    <div class="p-050">{appState.user.email}</div>
    <label for="username-input" class="p-050 nobreak">Your name: </label>
    <div class="p-050">
    <input bind:value={username}
           required
           id="username-input"
           type="text"
           class="max-w20 required"
           oninput={handleUsernameChange}
           placeholder="Name on the badge"
           minlength=1
           aria-describedby="input-name"/>
      <i class="fas fa-check success-checkmark"
         class:visible={isVisible}
         aria-label="Name saved successfully"></i>
      <p id="input-name" class="p-rl-4 m-0 user-hint help-text">
        Add a valid user name
      </p>
      </div>
    <label for="profile-picture-upload" class="p-050 nobreak">
      Profile picture:
    </label>
    <div class="image-container p-050">
      {#if appState.user.image}
        <img class="image-org" src={appState.user.image} alt="profile img" />
        <button class="mx-050 px-050 upload button1" onclick={deleteImage} aria-label="Delete image">
          <i class ="fas fa-trash-can"></i>
        </button>
      {:else}
        <button id="profile-picture-upload" class="upload accessible-btn"
                onclick={() => {fileInput? fileInput.click():() => {}}} aria-label="Upload profile picture">
          <i class="fas fa-upload" aria-hidden="true"></i>
          <input style="display:none" type="file" accept=".jpg, .jpeg, .png"
                  onchange={(e) => onFileSelected(e)} bind:this={fileInput} aria-hidden="true"/>
        </button>
      {/if}
    </div>
  </div>

  <h2>Connect your Git Email to this Account</h2>
  <p class="p-2 m-2">
    If you have other git email addresses, you can connect these addresses to
    your FlatFeeStack account. In case you didn't receive a confirmation email,
    please remove and re-add your git email address.
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
        {#each gitEmails as email (email.email)}
          <tr>
            <td data-label="Email">{email.email}</td>

            {#if email.confirmedAt}
              <td data-label="Confirmation" title={formatDate(new Date(email.confirmedAt))}>
                {timeSince(new Date(email.confirmedAt), new Date())} ago
              </td>
            {:else}
              <td data-label="Confirmation">
                <i class="fas fa-clock"></i>
                {email.createdAt? timeSince(new Date(email.createdAt), new Date()) + ' ago': 'now'}
              </td>
            {/if}
            <td data-label="Delete" class="center-flex">
              <button class="button1" onclick={() => removeEmail(email.email)} aria-label="Remove email">
                <i class="fas fa-trash"></i>
              </button>
            </td>
          </tr>
        {/each}
        <tr>
          <td colspan="3">
            <div>
              <input id="email-input" name="email" type="email" pattern={emailValidationPattern}
                     required bind:value={newEmail} placeholder="Email"/>
              <button class="button1" onclick={handleAddEmail}
                      disabled={!newEmail || !newEmail.match(emailValidationPattern)} aria-label="Add Git Email">
                Add Git Email
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <h2>Boost Your Impact with Multipliers</h2>
  <p class="p-2 m-2">
    Multipliers let you amplify support for projects. When enabled, you'll see this icon
    <img class="mul" src="/images/no-multiplier-coin.svg" alt="No Multiplier Icon" />
    next to repositories in the search tab.

    To activate a multiplier, click that icon and the icon will change to
    <img class="mul" src="/images/multiplier-coin.svg" alt="Multiplier Icon" />.

    Here's how it works: When other FlatFeeStack users donate to a multiplied repository,
    you automatically contribute a small bonus.
  </p>

  <div class="grid-2">
    <span class="p-050 nobreak">Enable Multiplier:</span>
    <div class="p-050">
      <input type="checkbox" bind:checked={multiplierActive} onchange={handleMultiplierToggle}/>
    </div>
  </div>

</Main>
