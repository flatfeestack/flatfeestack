<script lang="ts">
  import Navigation from "../components/Navigation.svelte";
  import Fa from "svelte-fa";
  import {onMount} from "svelte";
  import {API} from "../ts/api";
  import {faTrash, faClock} from "@fortawesome/free-solid-svg-icons";
  import {error, user, firstTime, config} from "../ts/store";
  import type {GitUser, UserBalanceCore} from "../types/users.ts";
  import {formatDate, formatMUSD, formatDay} from "../ts/services";
  import {navigate} from "svelte-routing";
  import {Contributions, PayoutAddress} from "../types/users.ts";
  import PayoutSelection from "../components/PayoutSelection.svelte";
  import {CryptoCurrency} from "../types/crypto";

  let address = "";
  let gitEmails: GitUser[] = [];
  let newEmail = "";
  let isSubmitting = false;
  let contributions: Contributions[] = [];
  let pendingPayouts: UserBalanceCore;
  let newPayoutAddress: ""
  let newPayoutCurrency: CryptoCurrency = { name: "Ethereum", shortName: "eth" }
  let payoutAddresses: PayoutAddress [] = [];

  async function handleAddPayoutAddress() {
    debugger
    try {
      let regex;
      switch (newPayoutCurrency.shortName) {
        case "ETH":
          regex = /^0x[a-fA-F0-9]{40}$/g
          break;
        case "NEO":
          break;
        case "XTZ":
          break;
        default:
          $error = "Invalid currency";
      }

      if (!newPayoutCurrency || (regex && !newPayoutAddress.match(regex))) {
        $error = "Invalid ethereum address";
      }

      let confirmedPayoutAddress: PayoutAddress = await API.user.addPayoutAddress(newPayoutCurrency.shortName, newPayoutAddress);
      payoutAddresses = [...payoutAddresses, confirmedPayoutAddress];
      newPayoutAddress = ""
      newPayoutCurrency = { name: "Ethereum", shortName: "eth" }
    } catch (e) {
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

  async function updatePayout(e) {
    try {
      if (!$user.payout_eth || !$user.payout_eth.match(/^0x[a-fA-F0-9]{40}$/g)) {
        $error = "Invalid ethereum address";
      }
      //TODO: no button, wait 1sec
      await API.user.updatePayoutAddress($user.payout_eth);
    } catch (e) {
      $error = e;
    }
  }

  async function handleSubmit() {
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
      const pr2 = API.user.contributionsRcv();
      const pr3 = API.user.pendingDailyUserPayouts();
      const pr4 = API.user.getPayoutAddresses();
      const res1 = await pr1;
      gitEmails = res1 ? res1 : gitEmails;
      const res2 = await pr2;
      contributions = res2 ? res2 : contributions;
      pendingPayouts = await pr3;
      //payoutAddresses = [{id: "1", currency: "eth", address: "abc..."}, {id: "2", currency: "neo", address: "cde..."}, {id: "3", currency: "tez", address: "fgh..."}]
      payoutAddresses = await pr4;
    } catch (e) {
      $error = e;
    }
  });

</script>

<Navigation>
  <h1 class="px-2">Income</h1>

  {#if !gitEmails || gitEmails.length === 0}
    <div class="container bg-green rounded p-2 m-2">
      Please add your git e-mail addresses to generate income
    </div>
  {/if}

  <div class="container">
    <label class="px-2">Add Git Email:</label>
    <input id="email-input" name="email" type="text" bind:value={newEmail} placeholder="Email" />
    <form class="p-2" on:submit|preventDefault="{handleSubmit}">
      <button class="button2" type="submit">Add Email</button>
    </form>
  </div>

  {#if gitEmails && gitEmails.length > 0}

    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Email</th>
          <th>Confirm Date</th>
          <th>Delete</th>
        </tr>
        </thead>
        <tbody>
        {#each gitEmails as email, key (email.email)}
          <tr>
            <td>{email.email}</td>
            <td>
              {#if email.confirmedAt}
                {formatDate(new Date(email.confirmedAt))}
              {:else }
                <Fa icon="{faClock}" size="md" />
              {/if}
            </td>
            <td class="cursor-pointer" on:click="{() => removeEmail(email.email)}">
              <Fa icon="{faTrash}" size="md" />
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="3">No Data</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <div class="container">
    <label class="px-2">Add Payout Address:</label>
    <select bind:value={newPayoutCurrency}>
      {#each $config.supportedCurrencies as currency}
        <option value={currency}>
          {currency.name}
        </option>
      {/each}
    </select>
    <input id="address-input" name="address" type="text" bind:value={newPayoutAddress} placeholder="Address" />
    <form class="p-2" on:submit|preventDefault="{handleAddPayoutAddress}">
      <button class="button2" disabled={!newPayoutCurrency} type="submit">Add address</button>
    </form>
  </div>
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
      {#each payoutAddresses as address, key (address.id)}
        <tr>
          <td><strong>{address.currency}</strong></td>
          <td>{address.address}</td>
          <td class="cursor-pointer" on:click="{() => removePaymentAddress(address.id)}">
            <Fa icon="{faTrash}" size="md" />
          </td>
        </tr>
      {:else}
        <tr>
          <td colspan="3">No Data</td>
        </tr>
      {/each}
      </tbody>
    </table>
  </div>

  <div class="container">
    <label class="px-2">Total realized income:</label>
  </div>

  <div class="container">
    <label class="px-2">Pending income:</label>{pendingPayouts?pendingPayouts.balance:0}
  </div>


  <!--<PayoutSelection />-->

  <div class="border-primary-500 rounded small p-2 m-2">
    Our commit evaluation engine analyzes within a timeframe of 2 month, and due to potential chargebacks,
    we cannot convert the funds immediately. Thus, after 2 month, the pending income can be withdrawn if
    the pending amount is larger than $25.
  </div>

  {#if $firstTime}
    <div class="container">
      <button class="button1 px-2" on:click="{() => {navigate(`/user/badges`)}}">Last step: View your track record</button>
    </div>
  {/if}

  {#if contributions && contributions.length > 0}
    <div class="container">
      <table>
        <thead>
        <tr>
          <th>Repository</th>
          <th>From</th>
          <th>Contribution</th>
          <th>Realized</th>
          <th>Balance USD</th>
          <th>Date</th>
        </tr>
        </thead>
        <tbody>
        {#each contributions as contribution}
          <tr>
            <td>{contribution.repoName}</td>
            <td>{contribution.userName}</td>
            <td>{contribution.contributorWeight * 100}%</td>
            <td>
              {#if contribution.contributorUserId}
                Realized
              {:else}
                Unclaimed
              {/if}
            </td>
            <td>{formatMUSD(contribution.balance)}</td>
            <td>{formatDay(new Date(contribution.day))}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="3">No Data</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
  {/if}

</Navigation>
