<script>
  import { getContext } from "svelte";
  export let title;
  export let message;
  export let onCancel = () => {};
  export let onConfirm = () => {};

  const { close } = getContext("simple-modal");

  let value;
  let onChange = () => {};

  function _onCancel() {
    onCancel();
    close();
  }

  function _onConfirm() {
    onConfirm(value);
    close();
  }

  $: onChange(value);
</script>

<style>
  .buttons {
    display: flex;
    justify-content: space-between;
  }
</style>

<h2>{title}</h2>
<p>{message}</p>

<div class="buttons">
  <button class="button1" on:click={_onCancel}> Cancel </button>
  <button class="button1" on:click={_onConfirm}> Confirm </button>
</div>
