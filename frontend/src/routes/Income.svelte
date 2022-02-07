<script lang="ts">
    import Navigation from "../components/Navigation.svelte";
    import {onMount} from "svelte";
    import {API} from "../ts/api";
    import {error} from "../ts/store";
    import type {UserBalanceCore} from "../types/users.ts";
    import {formatBalance, formatDay} from "../ts/services";
    import {Contributions} from "../types/users.ts";

    let address = "";
    let isSubmitting = false;
    let contributions: Contributions[] = [];
    let pendingPayouts: UserBalanceCore;

    onMount(async () => {
        try {
            const pr1 = API.user.contributionsRcv();
            const res1 = await pr1;
            contributions = res1 ? res1 : contributions;
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
                    <th>Contribution</th>
                    <th>Realized</th>
                    <th>Balance USD</th>
                    <th>Date</th>
                </tr>
                </thead>
                <tbody>
                {#each contributions as contribution}
                    <tr>
                        <td>{contribution.repoName}/{contribution.repoUrl}</td>
                        <td>{contribution.userName}</td>
                        <td>{contribution.contributorWeight * 100}%</td>
                        <td>
                            {#if contribution.contributorUserId}
                                Realized
                            {:else}
                                Unclaimed
                            {/if}
                        </td>
                        <td>{formatBalance(contribution.balance, "TODO")}</td>
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
