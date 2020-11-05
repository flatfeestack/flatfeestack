<style type="text/scss">
.nav-item {
}
</style>

<script type="ts">
import {
  faThList,
  faUser,
  faCodeBranch,
  faHandHoldingUsd,
  faCogs,
} from "@fortawesome/free-solid-svg-icons";
import DashboardNavigationItem from "../components/UI/DashboardNavigationItem.svelte";

import { ROUTES } from "src/types/routes";
import { Link } from "svelte-routing";
</script>

<div class="flex flex-col md:flex-row">
  <div
    class="md:fixed bg-white shadow-lg md:w-56 md:h-screen md:block hidden"
  ></div>
  <div
    class="bg-white h-16 fixed bottom-0 md:relative md:h-screen z-10 w-full md:w-56"
  >
    <div
      class="md:mt-12 md:w-48 md:fixed md:pin-l md:pin-t content-center md:content-start text-left justify-between"
    >
      <div class="flex flex-shrink-0 text-white mx-2 mb-10">
        <Link to="/">
          <img src="/assets/images/logo.svg" alt="Flatfeestack" />
        </Link>
      </div>
      <ul
        class="list-reset flex flex-row md:flex-col py-0 md:py-3 px-1 md:px-2 text-center md:text-left"
      >
        <DashboardNavigationItem
          icon="{faThList}"
          route="{ROUTES.DASHBOARD_OVERVIEW}"
          label="Overview"
        />
        <DashboardNavigationItem
          icon="{faUser}"
          route="{ROUTES.DASHBOARD_PROFILE}"
          label="Profile"
        />
        <DashboardNavigationItem
          icon="{faCodeBranch}"
          route="{ROUTES.DASHBOARD_SPONSORING}"
          label="Sponsoring"
        />
        <DashboardNavigationItem
          icon="{faHandHoldingUsd}"
          route="{ROUTES.DASHBOARD_INCOME}"
          label="Income"
        />
        <DashboardNavigationItem
          icon="{faCogs}"
          route="{ROUTES.DASHBOARD_SETTINGS}"
          label="Settings"
        />
      </ul>
    </div>
  </div>

  <div
    class="main-content flex-1 bg-grey-lightest mt-12 md:mt-2 pb-24 md:pb-5 px-3 md:px-0 md:pl-10"
  >
    <slot />
  </div>
</div>
