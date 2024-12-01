<script lang="ts">
    import type {Route} from "@mateothegreat/svelte5-router";
    import {Router} from "@mateothegreat/svelte5-router";

    import Landing from './Landing.svelte';
    import Login from "./auth/Login.svelte";
    import CatchAllRoute from "./CatchAllRoute.svelte";
    import DifferentChainId from "./DifferentChainId.svelte";
    import PublicBadges from "./PublicBadges.svelte";
    import Search from "./Search.svelte";
    import Payments from "./Payments.svelte";
    import Settings from "./Settings.svelte";
    import Income from "./Income.svelte";
    import Badges from "./Badges.svelte";
    import Invitations from "./Invitations.svelte";
    import Admin from "./Admin.svelte";
    import LoginWait from "./auth/LoginWait.svelte";
    import LoginConfirm from "./auth/LoginConfirm.svelte";
    import {API} from "./ts/api.ts";
    import {onMount} from "svelte";
    import {appState} from "./ts/state.ts";

    const routes: Route[] = [
        {path: "/user/search", component: Search},
        {path: "/user/payments", component: Payments},
        {path: "/user/settings", component: Settings},
        {path: "/user/income", component: Income},
        {path: "/user/badges", component: Badges},
        {path: "/user/invitations", component: Invitations},
        {path: "/user/admin", component: Admin},

        {path: "/differentChainId", component: DifferentChainId},
        {path: "/badges/:uuid", component: PublicBadges},
        {path: "/login-confirm", component: LoginConfirm},
        {path: "/login-wait", component: LoginWait},
        {path: "/login", component: Login},
        {path: "/", component: Landing},
        {path: "*", component: CatchAllRoute}
    ];

    onMount(async () => {
        if (!appState.$state.config) {
            appState.$state.config = await API.config.config();
        }
    });

</script>
<Router {routes}/>