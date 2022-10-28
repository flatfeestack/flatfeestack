<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import {error, user, config} from "../ts/store";
  import Fa from "svelte-fa";
  import { API } from "../ts/api";
  import { faUpload } from "@fortawesome/free-solid-svg-icons";
  import {faTrash, faClock} from "@fortawesome/free-solid-svg-icons";
  import type {GitUser, PayoutAddress, Currencies} from "../types/users";
  import {onMount} from "svelte";
  import {formatDate, timeSince} from "../ts/services";

  let nameOrig = $user.name;
  let timeoutName;
  let fileInput;

  let gitEmails: GitUser[] = [];
  let newEmail = "";

  let payoutAddresses: PayoutAddress[] = [];
  let currenciesWithoutWallet: Map<string, Currencies>;
  let newPayoutCurrency: string;
  let newPayoutAddress: ""

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

  $: {
    let tmp:Map<string, Currencies>=new Map<string, Currencies>();
    //https://stackoverflow.com/questions/34913675/how-to-iterate-keys-values-in-javascript
    const e = Object.entries($config.supportedCurrencies);
    for (const [key, value] of e) {
      if (!payoutAddresses.find(e => e.currency === key) && value.isCrypto) {
        tmp.set(key, value)
      }
    }
    currenciesWithoutWallet = tmp;
  }

  async function handleAddPayoutAddress() {
    try {
      let regex;
      switch (newPayoutCurrency) {
        case "ETH":
          regex = /^0x[a-fA-F0-9]{40}$/g
          break;
        case "GAS":
          break;
        case "XTZ":
          break;
        default:
          $error = "Invalid currency";
      }

      if (!newPayoutCurrency || (regex && !newPayoutAddress.match(regex))) {
        $error = "Invalid ethereum address";
      }

      let confirmedPayoutAddress: PayoutAddress = await API.user.addPayoutAddress(newPayoutCurrency, newPayoutAddress);
      payoutAddresses = [...payoutAddresses, confirmedPayoutAddress];
      newPayoutAddress = "";
    } catch (e) {
      if (e.response.status === 409) {
        $error = "Wallet Address is already used by someone else. Please use one Wallet per user."
        return
      }
      $error = e;
    }
  }

  async function removePaymentAddress(addressNumber: number) {
    try {
      await API.user.removePayoutAddress(addressNumber);
      payoutAddresses = payoutAddresses.filter((e) => e.id !== addressNumber);
    } catch (e) {
      $error = e;
    }
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

  async function handleAddEmail() {
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
      const pr1 = API.user.gitEmails();
      const pr2 = API.user.getPayoutAddresses();
      const res1 = await pr1;
      const res2 = await pr2;
      payoutAddresses = res2 ? res2: payoutAddresses;
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
    <label class="p-2 m-2 nobreak">Email:&nbsp;</label>
    <label class="p-2 m-2">{$user.email}</label>
    <label class="p-2 m-2 nobreak">Your name: </label>
    <input type="text" class="max-w20" bind:value={$user.name} placeholder="Name on the badge">
    <label class="p-2 m-2 nobreak">Profile picture:</label>
    <div class="upload" on:click={()=>{fileInput.click();}}>
      <Fa icon="{faUpload}" size="lg" class="icon, px-2" />
      <input style="display:none" type="file" accept=".jpg, .jpeg, .png" on:change={(e)=>onFileSelected(e)}
             bind:this={fileInput}>
      {#if $user.image}
        <img class="image-org" src="{$user.image}" />
      {/if}
    </div>
  </div>

  <h2 class="p-2 mt-40 ml-5 mb-0">Connect your Git Email to this Account</h2>
  <p class="p-2 m-2">If you have multiple git email addresses, you can connect these addresses to your FlatFeeStack
    account. You must
    verify your git email address. Once it has been validated, the confirm date will show the data of validation.
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
              <td title="{formatDate(new Date(email.confirmedAt))}">
                {timeSince(new Date(email.confirmedAt), new Date())} ago
              </td>
            {:else }
              <td><Fa icon="{faClock}" size="md"/></td>
            {/if}

          <td class="cursor-pointer" on:click="{() => removeEmail(email.email)}">
            <Fa icon="{faTrash}" size="md"/>
          </td>
        </tr>
      {/each}
      <tr>
        <td colspan="3">
          <div class="container-small">
            <input input-size="24" id="email-input" name="email" type="text" bind:value={newEmail} placeholder="Email"/>
            <form class="p-2" on:submit|preventDefault="{handleAddEmail}">
              <button class="ml-5 p-2 button1" type="submit">Add Git Email</button>
            </form>
          </div>
        </td>
      </tr>
      </tbody>
    </table>
  </div>

  <h2 class="p-2 ml-5 mb-5 mt-60">Add Your Payout Address</h2>
  <p class="p-2 m-2">You need to add a wallet address for each currency to receive the funds. In the pending income
    you
    can see which currencies have been sent to you. The payout will happen every month. For an Ethereum wallet you
    can use <a href="https://metamask.io/">Metamask</a>, for NEO you can use <a
            href="https://neoline.io/en/">NeoLine</a>.</p>

  <div class="container">
    <table>
      <thead>
      <tr>
        <th>Currency</th>
        <th>Payout Address</th>
        <th>Delete</th>
      </tr>
      </thead>
      <tbody>
      {#each payoutAddresses as address}
        <tr>
          <td><strong>{address.currency}</strong></td>
          <td>{address.address}</td>
          <td class="cursor-pointer" on:click="{() => removePaymentAddress(address.id)}">
            <Fa icon="{faTrash}" size="md"/>
          </td>
        </tr>
      {/each}
      <tr>
        {#if [...currenciesWithoutWallet].length > 0}
          <td colspan="3">
            <div class="container-small">
              <select bind:value={newPayoutCurrency}>
                {#each [...currenciesWithoutWallet] as [key, value]}
                  <option value={key}>
                    {value.name}
                  </option>
                {/each}
              </select>
              <input input-size="32" id="address-input" name="address" type="text" bind:value={newPayoutAddress}
                     placeholder="Address"/>
              <form class="p-2" on:submit|preventDefault="{handleAddPayoutAddress}">
                <button class="ml-5 p-2 button1" type="submit">Add address</button>
              </form>
            </div>
          </td>
        {/if}
      </tr>
      </tbody>
    </table>
  </div>

</Navigation>
