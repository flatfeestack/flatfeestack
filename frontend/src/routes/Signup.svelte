<script lang="ts">
import { Form, Input, Select, Choice } from "sveltejs-forms";
import { Link } from "svelte-routing";
import * as yup from "yup";
import { API } from "../api/api.ts";
import Spacer from "../components/UI/Spacer.svelte";

const schema = yup.object().shape({
  email: yup.string().required().email(),
  password: yup.string().required().min(8),
});

async function handleSubmit({
  detail: {
    values: { password, email },
    setSubmitting,
    resetForm,
  },
}) {
  try {
    const res = await API.auth.signup(email, password);
    setSubmitting(false);
    console.log(res);
  } catch (e) {
    console.log(e);
  }
}

function handleReset() {
  console.log("reset");
}
</script>

<div class="container">
  <div class="lead">Signup</div>
  <div class="subtitle">
    Already have an account already?
    <Link to="login">Login</Link>
  </div>
  <Spacer x5 />
  <Form
    schema="{schema}"
    on:submit="{handleSubmit}"
    on:reset="{handleReset}"
    let:isSubmitting
    let:isValid
  >
    <label>Email</label>
    <Input name="email" type="text" />
    <label>Password</label>
    <Input name="password" type="password" />
    <button type="submit" disabled="{isSubmitting}">Sign in</button>
  </Form>
</div>
