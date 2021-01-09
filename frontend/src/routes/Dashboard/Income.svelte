<style>
.sveltejs-forms {
  display: flex;
}
</style>

<script lang="ts">
import DashboardLayout from "../../layout/DashboardLayout.svelte";
import Fa from "svelte-fa";
import { onMount } from "svelte";
import { API } from "src/api/api";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import * as yup from "yup";
import { Form, Input } from "sveltejs-forms";
import Web3 from "../../components/Web3.svelte";
import CryptoAddressForm from "../../components/CryptoAddressForm.svelte";

let emails = [];
onMount(async () => {
  try {
    const response = await API.user.connectedEmails();

    if (response.data?.data?.length !== 0) {
      console.log(response.data);
      emails = response.data;
    }
  } catch (e) {
    console.log(e);
  }
});

let error = "";
const schema = yup.object().shape({
  email: yup.string().email(),
});

async function handleSubmit({
  detail: {
    values: { email },
    setSubmitting,
    resetForm,
  },
}) {
  try {
    error = "";
    await API.user.addEmail(email);
    setSubmitting(false);
    resetForm();
    emails = [...emails, email];
  } catch (e) {
    error = e.response?.data?.message || "Something went wrong";
    setSubmitting(false);
  }
}

let isSubmitting = false;

async function removeEmail(email: string) {
  try {
    error = "";
    console.log("remove email:", email);
    await API.user.removeEmail(email);
    emails = emails.filter((e) => e !== email);
  } catch (e) {
    console.log("in catch", e.message);
    error = e?.message || "Something went wrong";
  }
}
</script>

<DashboardLayout>
  <h1>Income</h1>
  <hr class="mb-10 w-64" />
  {#if emails.length === 0}
    <div class="flex mb-5">
      <div class="bg-red-500 text-white p-5">
        Please add your git e-mail addresses to generate income
      </div>
    </div>
  {/if}
  <h2 class="mb-5">Connected Git Emails</h2>
  {#each emails as email}
    <div class="div mb-2 flex flex-row items-center">
      <div class="w-64">
        <input type="text" class="input" value="{email}" disabled />
      </div>
      <div
        class="ml-5 hover:text-red-500 cursor-pointer transform hover:scale-105 duration-200"
        on:click="{() => removeEmail(email)}"
      >
        <Fa icon="{faTrash}" size="md" />
      </div>
    </div>
  {/each}
  <h2 class="my-5">Add Git Email</h2>
  <Form
    schema="{schema}"
    on:submit="{handleSubmit}"
    let:isSubmitting
    let:isValid
    class="sveltejs-forms"
  >
    <div class="flex flex-row items-end">
      <div class="w-64">
        <label
          for="email-input"
          class="block text-grey-darker text-sm font-bold mb-2 w-full"
        >Email
        </label>
        <Input id="email-input" name="email" type="text" class="input" />
      </div>
      <div class="ml-5">
        <button
          class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75"
          disabled="{isSubmitting}"
          type="submit"
        >Add Email{#if isSubmitting}...{/if}</button>
      </div>
      {#if error}
        <div class="bg-red-500 p-2 text-white mt-2">{error}</div>
      {/if}
    </div>
  </Form>
  <h2 class="my-5">Payout Address</h2>
  <CryptoAddressForm />

  <h2 class="my-5">Request Funds</h2>

  <Web3 />
</DashboardLayout>
