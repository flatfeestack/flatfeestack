<script lang="ts">
    import {
        faSquareCheck,
        faSquareXmark,
    } from "@fortawesome/free-solid-svg-icons";
    import { getContext, onMount } from "svelte";
    import Fa from "svelte-fa";
    import { membershipContract, signer } from "../../ts/daaStore";
    import { error } from "../../ts/mainStore";

    let membershipStatus = 0;
    let membershipFeePaid = false;
    let nextMembershipFeePayment = 0;

    const { close } = getContext("simple-modal");

    const approvalProcessSteps = [
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

            if (membershipStatus == 3) {
                const ethereumAddress = await $signer.getAddress();
                nextMembershipFeePayment =
                    await $membershipContract.nextMembershipFeePayment(
                        ethereumAddress
                    );

                membershipFeePaid =
                    nextMembershipFeePayment > Math.floor(Date.now() / 1000);
            }
        }
    });

    async function payMembershipFees() {
        try {
            const membershipFee = await $membershipContract.membershipFee();
            await $membershipContract.payMembershipFee({
                value: membershipFee,
            });
        } catch (e) {
            showErrorAndCloseModal(e);
        }
    }

    async function requestMembership() {
        try {
            await $membershipContract.requestMembership();
        } catch (e) {
            showErrorAndCloseModal(e);
        }
    }

    function showErrorAndCloseModal(e: Error) {
        $error = e.message;
        close();
    }
</script>

<div>
    <h1 class="text-secondary-900">Membership approval process</h1>

    <ul>
        {#each Array(3) as _, i}
            <li
                class={membershipStatus >= i + 1
                    ? "text-primary-500"
                    : "text-red"}
            >
                {#if membershipStatus >= i + 1}
                    <Fa icon={faSquareCheck} size="sm" class="icon" />
                {:else}
                    <Fa icon={faSquareXmark} size="sm" class="icon" />
                {/if}
                {approvalProcessSteps[i]}
            </li>
        {/each}

        <li class={membershipFeePaid ? "text-primary-500" : "text-red"}>
            {#if membershipFeePaid}
                <Fa icon={faSquareCheck} size="sm" class="icon" />
            {:else}
                <Fa icon={faSquareXmark} size="sm" class="icon" />
            {/if}
            Membership fees paid {#if membershipFeePaid}
                (until {new Date(
                    Number(nextMembershipFeePayment) * 1000
                ).toLocaleDateString()}){/if}
        </li>
    </ul>

    <div class="py-2 right">
        {#if membershipStatus == 0}
            <button on:click={requestMembership} class="button1"
                >Request membership</button
            >
        {/if}

        {#if membershipStatus == 3 && !membershipFeePaid}
            <button on:click={payMembershipFees} class="button1"
                >Pay membership fees</button
            >
        {/if}
    </div>
</div>

<style>
    ul {
        list-style-type: none;
        padding: 0;
        margin: 0;
    }
</style>
