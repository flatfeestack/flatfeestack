<script lang="ts">
    import { login } from "ts/auth";
    import { goto } from "@mateothegreat/svelte5-router";
    import AuthLayout from "./AuthLayout.svelte";
    import AuthForm from "./AuthForm.svelte";

    let error = $state("");
    let isSubmitting = $state(false);

    async function handleSubmit(email: string, password: string) {
        try {
            error = "";
            isSubmitting = true;
            await login(email, password);
            goto("/user/search");
        } catch (e) {
            error = String(e);
        } finally {
            isSubmitting = false;
        }
    }
</script>

<AuthLayout
        title="Sign in to FlatFeeStack"
        bottomText="New to FlatFeeStack?"
        bottomLinkText="Sign up"
        bottomLinkHref="/signup"
>
    <AuthForm
            showForgotPassword={true}
            showConfirmPassword={true}
            submitButtonText="Sign in"
            onSubmit={handleSubmit}
            error={error}
    />
</AuthLayout>