<script>
import DashboardLayout from "../../layout/DashboardLayout.svelte";
import { API } from "src/api/api";

const logout = async () => {
  return await API.auth.logout()
}

</script>

<DashboardLayout>
  <h1>Profile</h1>

  <button class="py-2 px-3 bg-primary-500 rounded-md text-white mt-4 disabled:opacity-75" on:click={logout}>
    Logout
  </button>
</DashboardLayout>
