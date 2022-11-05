<script lang="ts">
    import { membershipContract, provider, signer } from "../../ts/daaStore";
    import { onMount } from "svelte";
    import Fa from "svelte-fa";
    import {
        faSquareCheck,
        faSquareXmark,
    } from "@fortawesome/free-solid-svg-icons";

    export let membershipStatus = 0;
    export let nextMembershipFeePayment: Number | null = null;
    export let membershipFeePaid = false;

    export const approvalProcessSteps = [
        "Membership to the DAA has been requested",
        "Membership has been approved by one whitelister",
        "Membership has been approved by a second whitelister",
    ];

    onMount(async () => {
        if ($signer !== null || $membershipContract !== null) {
            const ethereumAddress = await $signer.getAddress();
            membershipStatus = await $membershipContract.getMembershipStatus(
                ethereumAddress
            );

            // is member, check if membership fees have been paid
            if (membershipStatus === 3) {
                nextMembershipFeePayment = (
                    await $provider.getBlock(
                        await $membershipContract.nextMembershipFeePayment(
                            ethereumAddress
                        )
                    )
                ).timestamp;
                membershipFeePaid =
                    nextMembershipFeePayment > Math.floor(Date.now() / 1000);
            }
        }
    });
</script>

<div>
    <h1 class="text-secondary-900">Membership approval process</h1>

    <ul>
        {#each Array(3) as _, i}
            <li class={membershipStatus >= i ? "text-primary-500" : "text-red"}>
                {#if membershipStatus >= i}
                    <Fa icon={faSquareCheck} size="sm" class="icon" />
                {:else}
                    <Fa icon={faSquareXmark} size="sm" class="icon" />
                {/if}
                {approvalProcessSteps[i]}
            </li>
        {/each}

        {#if membershipStatus >= 3}
            <li
                class={nextMembershipFeePayment
                    ? "text-primary-500"
                    : "text-red"}
            >
                {#if nextMembershipFeePayment}
                    <Fa icon={faSquareCheck} size="sm" class="icon" />
                {:else}
                    <Fa icon={faSquareXmark} size="sm" class="icon" />
                {/if}
                Membership fees paid
            </li>
        {/if}
    </ul>
</div>

<style>
    ul {
        list-style-type: none;
        padding: 0;
        margin: 0;
    }
</style>
