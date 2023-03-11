<script lang="ts">
  import {
    faCheck,
    faQuestion,
    faXmark,
  } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  export let disabled = false;
  export let onClick: (proposalId: string, voteValue: number) => void;
  export let proposalId: string;
  export let voteValue = 0;
</script>

<div>
  <button
    {disabled}
    class={voteValue == 0 ? "button4" : "button3"}
    on:click={() => onClick(proposalId, 0)}
  >
    <Fa icon={faXmark} size="sm" class="icon px-2" />
  </button>

  <button
    {disabled}
    class={voteValue == 1 ? "button4" : "button3"}
    on:click={() => onClick(proposalId, 1)}
  >
    <Fa icon={faCheck} size="sm" class="icon px-2" />
  </button>

  <button
    {disabled}
    class={voteValue == 3 ? "button4" : "button3"}
    on:click={() => onClick(proposalId, 2)}
  >
    <Fa icon={faQuestion} size="sm" class="icon px-2" />
  </button>
</div>
