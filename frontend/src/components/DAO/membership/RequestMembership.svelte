<script lang="ts">
  import { getContext } from "svelte";
  import { bylawsUrl } from "../../../ts/daoStore";

  let isAgreed = false;

  const { close } = getContext("simple-modal");

  export let onMembershipCancel = () => {};
  export let onMembershipConfirm = () => {};

  function _onCancel() {
    onMembershipCancel();
    close();
  }

  function _onConfirm() {
    onMembershipConfirm();
    close();
  }
</script>

<style>
  .buttons {
    display: flex;
    justify-content: space-between;
  }
</style>

<div>
  <h1 class="text-secondary-900">Request Membership</h1>

  <div class="py-2">
    <div class="py-2">
      <label>
        <input type="checkbox" bind:checked={isAgreed} />
        I agree with the current
        {#if $bylawsUrl === null}bylaws{:else}<a
            href={$bylawsUrl}
            target="_blank"
            rel="noreferrer">bylaws</a
          >{/if}.
      </label>
    </div>
  </div>

  <div class="buttons">
    <button class="button4" on:click={_onCancel}> Cancel </button>
    <button class="button4" on:click={_onConfirm} disabled={!isAgreed}>
      Request membership
    </button>
  </div>
</div>
