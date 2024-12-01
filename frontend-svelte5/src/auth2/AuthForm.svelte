<script lang="ts">
    import Dots from "Dots.svelte";
    import { route } from "@mateothegreat/svelte5-router";
    import {emailValidationPattern} from "../utils.ts";

    let {
        showConfirmPassword = false,
        submitButtonText = "Submit",
        showForgotPassword = false,
        onSubmit,
        error,
    }:{
        showConfirmPassword: boolean,
        submitButtonText: string,
        showForgotPassword: boolean,
        onSubmit: (email: string, password: string) => void,
        error:string} = $props();

    let email = $state("");
    let password = $state("");
    let confirmPassword = $state("");
    let isSubmitting = $state(false);

    function handleSubmit() {
        if (showConfirmPassword && password !== confirmPassword) {
            error = "Password and confirmation password do not match.";
            return;
        }
        onSubmit(email, password);
    }
</script>

<form onsubmit = {handleSubmit}>
    <label for="email" class="py-1">Email address</label>
    <input
            required
            size="100"
            maxlength="100"
            type="email"
            id="email"
            pattern={emailValidationPattern}
            name="email"
            bind:value={email}
    />

    <div class="flex py-1">
        <label for="password">Password</label>
        {#if showForgotPassword}
            <a href="/forgot" use:route tabindex="-1">Forgot password?</a>
        {/if}
    </div>

    <input
            required
            size="100"
            maxlength="100"
            type="password"
            id="password"
            minlength="8"
            bind:value={password}
    />

    {#if showConfirmPassword}
        <label for="confirmPassword" class="flex py-1">Confirm Password</label>
        <input
                required
                size="100"
                maxlength="100"
                type="password"
                id="confirmPassword"
                minlength="8"
                bind:value={confirmPassword}
        />
    {/if}

    <button class="button1 my-4" disabled={isSubmitting} type="submit">
        {submitButtonText}
        {#if isSubmitting}<Dots />{/if}
    </button>

    {#if error}
        <div class="bg-red rounded p-2">{error}</div>
    {/if}
</form>

<style>
    form {
        display: flex;
        flex-direction: column;
    }
    label {
        color: var(--primary-900);
    }
</style>