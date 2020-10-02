<script lang="ts">
import { Form, Input, Select, Choice } from "sveltejs-forms";
import { Link } from "svelte-routing";
import * as yup from "yup";
import { API } from "../api/api.ts";
import Spacer from "../components/UI/Spacer.svelte";
import { token } from "../store/auth.ts";
import { navigate } from "svelte-routing";

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
    const res = await API.auth.login(email, password);
    const t = res.headers["token"];
    if (t) {
      token.set(t);
    }
    setSubmitting(false);
    console.log();
    navigate("dashboard");
    resetForm();
  } catch (e) {
    console.log(e);
  }
}

function handleReset() {
  console.log("reset");
}
</script>

<div class="page container">
  <div class="lead">Login</div>
  <div class="subtitle">
    Don't have an account already?
    <Link to="signup">Sign up</Link>
  </div>
  <Spacer x5 />
  <div class="flex justify-center">
    <div class="card">
      <Form
        schema="{schema}"
        on:submit="{handleSubmit}"
        on:reset="{handleReset}"
        let:isSubmitting
        let:isValid
      >
        <label for="email-input">Email </label>
        <Input id="email-input" name="email" type="text" />
        <Spacer x2 />
        <label for="password-input">Password </label>
        <Input id="password-input" name="password" type="password" />
        <button type="submit" disabled="{isSubmitting}">Sign in</button>
      </Form>
    </div>
  </div>
</div>
