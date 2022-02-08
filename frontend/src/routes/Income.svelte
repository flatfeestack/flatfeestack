<script lang="ts">
    import Navigation from "../components/Navigation.svelte";
    import {onMount} from "svelte";
    import {API} from "../ts/api";
    import {error} from "../ts/store";
    import type {UserBalanceCore} from "../types/users.ts";
    import {formatBalance, formatDate, formatDay, timeSince} from "../ts/services";
    import {Contributions} from "../types/users.ts";

    let address = "";
    let isSubmitting = false;
    let contributions: Contributions[] = [];
    let pendingPayouts: UserBalanceCore;

    function aggregate(contributions: Contributions[]): (string[]) {
        return [];
    }

    onMount(async () => {
        try {
            const pr1 = API.user.contributionsRcv();
            const res1 = await pr1;
            contributions = res1 ? res1 : contributions;
            contributions = aggregate(contributions)
        } catch (e) {
            $error = e;
        }
    });

</script>

<Navigation>
    {#if contributions && contributions.length > 0}
        <div class="container">
            <table>
                <thead>
                <tr>
                    <th>Repository</th>
                    <th>From</th>
                    <th>Balance</th>
                    <th>Currency</th>
                    <th>Realized</th>
                    <th>Date</th>
                </tr>
                </thead>
                <tbody>
                {#each contributions as contribution}
                    <tr>
                        <td><a href="{contribution.repoUrl}">{contribution.repoName}</a></td>
                        <td>{contribution.sponsorName?contribution.sponsorName:"[no name]"}</td>
                        <td>{contribution.balance}</td>
                        <td>{contribution.currency}</td>
                        <td>
                            {#if contribution.contributorUserId}
                                Realized
                            {:else}
                                Unclaimed
                            {/if}
                        </td>
                        <td title="{formatDate(new Date(contribution.day))}">
                            {timeSince(new Date(contribution.day), new Date())} ago
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

</Navigation>
