<script lang="ts">
  import type { ProposalFormProps } from "../../../types/daa";
  import yup from "../../../utils/yup";

  interface $$Props extends ProposalFormProps {}

  export let targets: $$Props["targets"];
  export let values: $$Props["values"];
  export let transferCallData: $$Props["transferCallData"];

  let formValues = {
    targets: ["example address"],
    values: [0],
    transferCallData: ["example call data"],
  };

  const schema = yup.object().shape({
    targets: yup.array().min(1).of(yup.string().isEthereumAddress()).required(),
    values: yup.array().min(1).of(yup.number()).required(),
    transferCallData: yup.array().min(1).of(yup.string()).required(),
  });

  $: {
    try {
      schema.validateSync(formValues, { abortEarly: false });
      updateCalldata();
    } catch (err) {
      // ignore errors for now
    }
  }

  function updateCalldata() {
    values = formValues.values;
    targets = formValues.targets;
    transferCallData = formValues.transferCallData;
  }

  function addAdditionalCall() {
    formValues.targets = [...formValues.targets, "another call"];
    formValues.values = [...formValues.values, 0];
    formValues.transferCallData = [
      ...formValues.transferCallData,
      "another set of calldata",
    ];
  }
</script>

<style>
  .combine-column {
    grid-column: 1 / 3;
  }
</style>

{#each formValues.targets as _targets, i}
  <h2 class="combine-column">Call {i + 1}</h2>

  <label for="target_{i}">Target</label>
  <input
    type="text"
    id="target_{i}"
    name="target[{i}]"
    bind:value={formValues.targets[i]}
    required
  />
  {#await schema.validateAt(`targets[${i}]`, formValues)}{:catch error}
    <p class="invalid combine-column" style="color:red">{error.errors[0]}</p>
  {/await}

  <label for="value_{i}">Value</label>
  <input
    type="number"
    id="value_{i}"
    name="value[{i}]"
    bind:value={formValues.values[i]}
    required
  />
  {#await schema.validateAt(`values[${i}]`, formValues)}{:catch error}
    <p class="invalid combine-column" style="color:red">{error.errors[0]}</p>
  {/await}

  <label class="combine-column" for="calldata_{i}">Calldata</label>
  <textarea
    class="combine-column"
    id="calldata_{i}"
    name="calldata[{i}]"
    bind:value={formValues.transferCallData[i]}
    rows="4"
    cols="50"
  />

  {#await schema.validateAt(`transferCallData[${i}]`, formValues)}{:catch error}
    <p class="combine-column invalid" style="color:red">{error.errors[0]}</p>
  {/await}
{/each}

<p />

<button class="button1" on:click={() => addAdditionalCall()}>Add call</button>
