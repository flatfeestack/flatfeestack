<script lang="ts">
    import type {Route} from "@mateothegreat/svelte5-router";
    import {Router} from "@mateothegreat/svelte5-router";
    import Landing from './Landing.svelte';
    import Login from "@/Login.svelte";
    import CatchAllRoute from "@/CatchAllRoute.svelte";
    import Forgot from "@/Forgot.svelte";
    import Signup from "@/Signup.svelte";
    import DifferentChainId from "@/DifferentChainId.svelte";
    import PublicBadges from "@/PublicBadges.svelte";
    import ConfirmForgot from "@/ConfirmForgot.svelte";
    import ConfirmSignup from "@/ConfirmSignup.svelte";
    import ForwardGitEmail from "@/ForwardGitEmail.svelte";
    import ConfirmInvite from "@/ConfirmInvite.svelte";
    import Search from "@/Search.svelte";
    import Payments from "@/Payments.svelte";
    import Settings from "@/Settings.svelte";
    import Income from "@/Income.svelte";
    import Badges from "@/Badges.svelte";
    import Invitations from "@/Invitations.svelte";
    import Admin from "@/Admin.svelte";

    const routes: Route[] = [
        {path: "/confirm/reset/:email/:token", component: ConfirmForgot},
        {path: "/confirm/signup/:email/:token", component: ConfirmSignup},
        {path: "/confirm/git-email/:email/:token", component: ForwardGitEmail},
        {path: "/confirm/invite/:email/:emailToken/:inviteByEmail", component: ConfirmInvite},

        {path: "/user/search", component: Search},
        {path: "/user/payments", component: Payments},
        {path: "/user/settings", component: Settings},
        {path: "/user/income", component: Income},
        {path: "/user/badges", component: Badges},
        {path: "/user/invitations", component: Invitations},
        {path: "/user/admin", component: Admin},

        {path: "/differentChainId", component: DifferentChainId},
        {path: "/badges/:uuid", component: PublicBadges},
        {path: "/forgot", component: Forgot},
        {path: "/signup", component: Signup},
        {path: "/login", component: Login},
        {path: "/", component: Landing},
        {path: "*", component: CatchAllRoute}
    ];
</script>
<Router {routes}/>