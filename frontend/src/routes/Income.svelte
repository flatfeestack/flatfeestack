<script lang="ts">
    import Navigation from "../components/Navigation.svelte";
    import {onMount} from "svelte";
    import {API} from "../ts/api";
    import {error} from "../ts/store";
    import type {UserBalanceCore} from "../types/users.ts";
    import {formatMUSD, formatDay} from "../ts/services";
    import {navigate} from "svelte-routing";
    import {Contributions} from "../types/users.ts";
    import Spinner from "../components/Spinner.svelte";

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
    <h2 class="p-2 ml-5 mb-0">Income</h2>
    <div class="container-small">
        <div class="container-col-small">
            <p>Realized income (transferred to your account)</p>
            {#await API.user.totalRealizedIncome()}
                <Spinner/>
            {:then res}
                <table>
                    <thead>
                    <tr>
                        <th>Currency</th>
                        <th>Amount</th>
                    </tr>
                    </thead>
                    <tbody>
                    {#if res && res.length > 0}
                        {#each res as row}
                            <tr>
                                <td>{row.currency}</td>
                                <td>{row.balance}</td>
                            </tr>
                        {:else}
                            <tr>
                                <td colspan="4">No Data</td>
                            </tr>
                        {/each}
                    {:else}
                        <tr>
                            <td colspan="4">No Data</td>
                        </tr>
                    {/if}
                    </tbody>
                </table>
            {:catch err}
                {error.set(err)}
            {/await}
        </div>

        <div class="container-col-small">
            <p>Pending income (will be transfered)</p>
            {#await API.user.pendingDailyUserPayouts()}
                <Spinner/>
            {:then res}
                <table>
                    <thead>
                    <tr>
                        <th>Currency</th>
                        <th>Amount</th>
                    </tr>
                    </thead>
                    <tbody>
                    {#if res && res.length > 0}
                        {#each res as row}
                            <tr>
                                <td>{row.currency}</td>
                                <td>{row.balance}</td>
                            </tr>
                        {:else}
                            <tr>
                                <td colspan="4">No Data</td>
                            </tr>
                        {/each}
                    {:else}
                        <tr>
                            <td colspan="4">No Data</td>
                        </tr>
                    {/if}
                    </tbody>
                </table>
            {:catch err}
                {error.set(err)}
            {/await}
        </div>
        <!--<PayoutSelection />-->

    </div>

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
                        <td>{contribution.repoName}</td>
                        <td>{contribution.userName}</td>
                        <td>{contribution.contributorWeight * 100}%</td>
                        <td>
                            {#if contribution.contributorUserId}
                                Realized
                            {:else}
                                Unclaimed
                            {/if}
                        </td>
                        <td>{formatMUSD(contribution.balance)}</td>
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
