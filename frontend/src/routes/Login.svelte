<style>
.sveltejs-forms {
  @apply w-2/3 mx-auto py-10;
}
.message {
  @apply text-red-500 pt-1;
}
</style>

<script lang="ts">
import { Form, Input } from "sveltejs-forms";
import { Link } from "svelte-routing";
import * as yup from "yup";
import { navigate } from "svelte-routing";
import { login } from "../store/authService";
import PageLayout from "../layout/PageLayout.svelte";

let error = "";
const schema = yup.object().shape({
  //TOOD: add validation again
  email: yup.string().required(), //.email(),
  password: yup.string().required(), //.min(8),
});

async function handleSubmit({
  detail: {
    values: { password, email },
    setSubmitting,
    resetForm,
  },
}) {
  try {
    await login(email, password);
    setSubmitting(false);
    navigate("dashboard");
    resetForm();
  } catch (e) {
    if (e?.response?.status === 400) {
      error = "No match found for username / password combination";
    } else {
      error = "Something went wrong. Please try again.";
    }
    setSubmitting(false);
    resetForm();
  }
}

function handleReset() {
  console.log("reset");
}

let isSubmitting = false;
</script>

<PageLayout>
  <h1 class="text-primary-500">Login</h1>
  <div class="subtitle">
    Don't have an account already?
    <span class="text-secondary-500">
      <Link to="signup">Sign up</Link>
    </span>
  </div>
  <div class="bg-white shadow-xl flex mx-auto w-1/3">
    <div class="mx-auto py-5 w-64">
      <Form
        schema="{schema}"
        on:submit="{handleSubmit}"
        on:reset="{handleReset}"
        let:isSubmitting
        let:isValid
      >
        <label
          for="email-input"
          class="block text-grey-darker text-sm font-bold mb-2 w-full"
        >Email
        </label>
        <Input id="email-input" name="email" type="text" class="input" />
        <label
          for="password-input"
          class="block text-grey-darker text-sm font-bold mb-2 w-full mt-5"
        >Password
        </label>
        <Input
          id="password-input"
          name="password"
          type="password"
          class="input w-100"
        />
        <button
          class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75"
          disabled="{isSubmitting}"
          type="submit"
        >Sign in{#if isSubmitting}...{/if}</button>

        {#if error}
          <div class="bg-red-500 p-2 text-white mt-2">{error}</div>
        {/if}
      </Form>
    </div>
  </div>
</PageLayout>
