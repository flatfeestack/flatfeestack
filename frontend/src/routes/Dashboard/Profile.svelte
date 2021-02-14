<script>
import DashboardLayout from "./DashboardLayout.svelte";
import { API } from "ts/api";
import { user } from "ts/auth";


const logout = async () => {
  return await API.auth.logout()
}

let checked = $user.mode != "ORGANIZATION";
$: {
  if (checked == false) {
    $user.mode = "ORGANIZATION"
  } else {
    $user.mode = "CONTRIBUTOR"
  }
}

</script>

<DashboardLayout>
  <h1>Profile</h1>

  <div class="onoffswitch">
    <input type="checkbox" bind:checked={checked} name="onoffswitch" class="onoffswitch-checkbox" id="myonoffswitch" tabindex="0" >
    <label class="onoffswitch-label" for="myonoffswitch">
      <span class="onoffswitch-inner"></span>
      <span class="onoffswitch-switch"></span>
    </label>
  </div>

  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={logout}>
    Logout
  </button>
</DashboardLayout>
