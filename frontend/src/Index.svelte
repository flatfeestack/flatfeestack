<script lang="ts">
    import { Router, type Routes } from 'preveltekit';
    import Landing from './pages/Landing.svelte';
    import Login from "./auth/Login.svelte";
    import CatchAllRoute from "./CatchAllRoute.svelte";
    import DifferentChainId from "./DifferentChainId.svelte";
    import PublicBadges from "./PublicBadges.svelte";
    import Search from "./pages/Search.svelte";
    import Payments from "./pages/Payments.svelte";
    import Settings from "./pages/Settings.svelte";
    //import Income from "./pages/Income.svelte";
    import Badges from "./pages/Badges.svelte";
    import Invitations from "./pages/Invitations.svelte";
    import Admin from "./pages/admin/Admin.svelte";
    import LoginWait from "./auth/LoginWait.svelte";
    import LoginConfirm from "./auth/LoginConfirm.svelte";
    import {API} from "./ts/api.ts";
    import {onMount} from "svelte";
    import {appState} from "./ts/state.svelte.ts";
    import TermsAndConditions from "./auth/TermsAndConditions.svelte";
    import ChallengeTask from "./pages/ChallengeTask.svelte";

    const routes: Routes = { dynamicRoutes:[
        {path: "/user/search", component: Search},
        {path: "/user/payments", component: Payments},
        {path: "/user/settings", component: Settings},
        //{path: "/user/income", component: Income},
        {path: "/user/badges", component: Badges},
        {path: "/user/invitations", component: Invitations},
        {path: "/user/admin", component: Admin},
        {path: "/user/admin", component: Admin},
        {path: "/user/healthy-repos", component: Admin},

        {path: "/ChallengeTask", component: ChallengeTask},

        {path: "/differentChainId", component: DifferentChainId},
        {path: "/badges/:uuid", component: PublicBadges},
        {path: "/toc2", component: TermsAndConditions},
        {path: "/login-confirm", component: LoginConfirm},
        {path: "/login-wait/:email", component: LoginWait},
        {path: "/login", component: Login},
        {path: "/", component: Landing},
        {path: "*", component: CatchAllRoute}
    ]};

    onMount(async () => {
        if (!appState.config) {
            appState.config = await API.config.config();
        }
    });

</script>
<Router {routes}/>