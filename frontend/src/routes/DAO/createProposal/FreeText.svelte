<script lang="ts">
  import type { Call, ProposalFormProps } from "$lib/types/dao";
  import yup from "$lib/utils/yup";

  interface $$Props extends ProposalFormProps {}

  export let calls: $$Props["calls"];

  let formValues: Call[] = [];

  const schema = yup
    .array()
    .min(1)
    .of(
      yup.object().shape({
        target: yup.string().isEthereumAddress().required(),
        value: yup.number().min(0).required(),
        transferCallData: yup.string().required(),
      })
    );

  $: {
    try {
      schema.validateSync(formValues, { abortEarly: false });
      updateCalldata();
    } catch (err) {
      // ignore errors for now
    }
  }

  function updateCalldata() {
    calls = formValues;
  }

  function addAdditionalCall() {
    formValues = [
      ...formValues,
      {
        target: "another target",
        value: 0,
        transferCallData: "another set of calldata",
      },
    ];
  }
</script>

<style>
  .combine-column {
    grid-column: 1 / 3;
  }
</style>

{#each formValues as call, i}
  <h2 class="combine-column">Call {i + 1}</h2>

  <label for="target_{i}">Target</label>
  <input
    type="text"
    id="target_{i}"
    name="target[{i}]"
    bind:value={call.target}
    required
  />
  {#await schema.validateAt(`[${i}].target`, formValues) catch error}
    <p class="invalid combine-column" style="color:red">{error.errors[0]}</p>
  {/await}

  <label for="value_{i}">Value</label>
  <input
    type="number"
    id="value_{i}"
    name="value[{i}]"
    bind:value={call.value}
    required
  />
  {#await schema.validateAt(`[${i}].value`, formValues) catch error}
    <p class="invalid combine-column" style="color:red">{error.errors[0]}</p>
  {/await}

  <label class="combine-column" for="calldata_{i}">Calldata</label>
  <textarea
    class="combine-column"
    id="calldata_{i}"
    name="calldata[{i}]"
    bind:value={call.transferCallData}
    rows="4"
    cols="50"
  />

  {#await schema.validateAt(`[${i}].transferCallData`, formValues) catch error}
    <p class="combine-column invalid" style="color:red">{error.errors[0]}</p>
  {/await}
{/each}

<p />

<button class="button4" on:click={() => addAdditionalCall()}>Add call</button>
